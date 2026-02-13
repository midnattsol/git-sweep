package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/midnattsol/git-sweep/internal/sweep"
)

// PickerItem represents an item in the picker
type PickerItem struct {
	Name          string
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
			Name: br.Branch.Name,
		}

		if br.Skip != "" {
			item.Disabled = true
			item.DisableReason = string(br.Skip)
		} else {
			// Pre-select eligible branches
			item.Selected = true
		}

		items = append(items, item)
	}

	return PickerModel{
		items: items,
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
		}
	}

	return m, nil
}

func (m PickerModel) View() string {
	var b strings.Builder

	b.WriteString(RenderNukeHeader())
	b.WriteString(fmt.Sprintf("\n  %s\n\n", MutedStyle.Render("Select branches to delete:")))

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = CursorStyle.Render() + " "
		}

		var checkbox string
		if item.Disabled {
			checkbox = MutedStyle.Render("[ ]")
		} else if item.Selected {
			checkbox = SuccessStyle.Render("[x]")
		} else {
			checkbox = "[ ]"
		}

		name := item.Name
		if i == m.cursor && !item.Disabled {
			name = SelectedStyle.Render(name)
		} else if item.Disabled {
			name = MutedStyle.Render(name)
		}

		line := fmt.Sprintf("%s%s %s", cursor, checkbox, name)

		if item.Disabled && item.DisableReason != "" {
			line += fmt.Sprintf("  %s", MutedStyle.Render(item.DisableReason))
		}

		b.WriteString(line + "\n")
	}

	// Help
	b.WriteString(fmt.Sprintf("\n  %s\n", Divider(50)))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		HelpStyle.Render("space select · a all · n none · enter confirm · q quit")))

	return b.String()
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
