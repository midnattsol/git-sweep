package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/midnattsol/git-sweep/internal/sweep"
)

// PickerItem represents an item in the picker
type PickerItem struct {
	Name          string
	Category      sweep.Category
	LastCommit    time.Time
	Selected      bool
	Disabled      bool
	DisableReason string
}

// PickerModel is a bubbletea model for multi-select
type PickerModel struct {
	items     []PickerItem
	cursor    int
	quitting  bool
	confirmed bool
}

// NewPicker creates a new picker from sweep results
func NewPicker(result *sweep.Result) PickerModel {
	items := make([]PickerItem, 0, len(result.Branches))

	for _, br := range result.Branches {
		item := PickerItem{
			Name:       br.Branch.Name,
			Category:   br.Category,
			LastCommit: br.Branch.LastCommit,
		}

		if br.Skip != "" {
			item.Disabled = true
			item.DisableReason = string(br.Skip)
		} else {
			// Pre-select suggested branches
			item.Selected = br.Suggested
		}

		items = append(items, item)
	}

	// Sort by category: suggested first, then gone, orphan, active, protected
	sort.SliceStable(items, func(i, j int) bool {
		return categoryOrder(items[i].Category) < categoryOrder(items[j].Category)
	})

	return PickerModel{
		items: items,
	}
}

func categoryOrder(c sweep.Category) int {
	switch c {
	case sweep.CategorySuggested:
		return 0
	case sweep.CategoryGone:
		return 1
	case sweep.CategoryOrphan:
		return 2
	case sweep.CategoryActive:
		return 3
	case sweep.CategoryProtected:
		return 4
	default:
		return 5
	}
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.confirmed = true
			return m, tea.Quit

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}

		case " ":
			// Toggle selection
			if !m.items[m.cursor].Disabled {
				m.items[m.cursor].Selected = !m.items[m.cursor].Selected
			}

		case "a":
			// Select all
			for i := range m.items {
				if !m.items[i].Disabled {
					m.items[i].Selected = true
				}
			}

		case "n":
			// Select none
			for i := range m.items {
				m.items[i].Selected = false
			}

		case "s":
			// Select only suggested
			for i := range m.items {
				if !m.items[i].Disabled {
					m.items[i].Selected = m.items[i].Category == sweep.CategorySuggested ||
						m.items[i].Category == sweep.CategoryGone
				}
			}
		}
	}

	return m, nil
}

func (m PickerModel) View() string {
	var b strings.Builder

	b.WriteString(RenderNukeHeader())
	b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Select branches to delete:")))

	currentCategory := sweep.Category("")

	for i, item := range m.items {
		// Print category header when it changes
		if item.Category != currentCategory {
			currentCategory = item.Category
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("  %s\n", categoryHeader(currentCategory)))
		}

		cursor := "  "
		if i == m.cursor {
			cursor = CursorStyle.Render() + " "
		}

		var checkbox string
		if item.Disabled {
			checkbox = MutedStyle.Render("▢")
		} else if item.Selected {
			checkbox = SuccessStyle.Render("▣")
		} else {
			checkbox = "▢"
		}

		name := item.Name
		if i == m.cursor && !item.Disabled {
			name = SelectedStyle.Render(name)
		} else if item.Disabled {
			name = MutedStyle.Render(name)
		}

		// Build the line
		line := fmt.Sprintf("%s%s %s", cursor, checkbox, name)

		// Add age or reason
		if item.Disabled && item.DisableReason != "" {
			line += fmt.Sprintf("  %s", MutedStyle.Render(item.DisableReason))
		} else if !item.LastCommit.IsZero() {
			line += fmt.Sprintf("  %s", MutedStyle.Render(formatAge(item.LastCommit)))
		}

		b.WriteString(line + "\n")
	}

	// Help
	b.WriteString(fmt.Sprintf("\n  %s\n", Divider(55)))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		HelpStyle.Render("space select · a all · s suggested · n none · enter confirm · q quit")))

	return b.String()
}

func categoryHeader(c sweep.Category) string {
	switch c {
	case sweep.CategorySuggested:
		return WarningStyle.Render("Suggested") + MutedStyle.Render(" (stale, no upstream)")
	case sweep.CategoryGone:
		return WarningStyle.Render("Gone") + MutedStyle.Render(" (upstream deleted, no PR found)")
	case sweep.CategoryOrphan:
		return MutedStyle.Render("Orphan") + MutedStyle.Render(" (no upstream, recent)")
	case sweep.CategoryActive:
		return MutedStyle.Render("Active") + MutedStyle.Render(" (has upstream)")
	case sweep.CategoryProtected:
		return MutedStyle.Render("Protected")
	default:
		return MutedStyle.Render("Other")
	}
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)

	switch {
	case d < time.Hour*24:
		return "today"
	case d < time.Hour*24*2:
		return "yesterday"
	case d < time.Hour*24*7:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < time.Hour*24*30:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case d < time.Hour*24*365:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// Cancelled returns true if user quit without confirming
func (m PickerModel) Cancelled() bool {
	return m.quitting
}

// SelectedBranches returns the names of selected branches
func (m PickerModel) SelectedBranches() []string {
	var selected []string
	for _, item := range m.items {
		if item.Selected && !item.Disabled {
			selected = append(selected, item.Name)
		}
	}
	return selected
}

// RunPicker runs the interactive picker and returns selected branch names
func RunPicker(result *sweep.Result) ([]string, error) {
	m := NewPicker(result)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(PickerModel)
	if fm.Cancelled() {
		return nil, nil // User cancelled
	}

	return fm.SelectedBranches(), nil
}
