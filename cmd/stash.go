package cmd

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/stash"
	"github.com/midnattsol/git-sweep/internal/ui"
)

var (
	flagStashClean bool
	flagStashDays  int
	flagStashForce bool
)

func NewStashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stash",
		Short: "Clean up old git stashes",
		Long: `List and clean up old git stashes.

Shows stashes with useful context: date, branch, files affected.

Examples:
  git sweep stash              # List all stashes with details
  git sweep stash --clean      # Interactive picker to delete stashes
  git sweep stash --clean --force   # Delete stashes older than 30 days without asking
  git sweep stash --days 60    # Consider stashes older than 60 days as "old"`,
		RunE: runStash,
	}

	cmd.Flags().BoolVar(&flagStashClean, "clean", false, "Interactive picker to delete stashes")
	cmd.Flags().BoolVar(&flagStashForce, "force", false, "Delete old stashes without confirmation (use with --clean)")
	cmd.Flags().IntVar(&flagStashDays, "days", 30, "Days threshold for considering a stash 'old'")

	return cmd
}

func runStash(cmd *cobra.Command, args []string) error {
	fmt.Print(ui.RenderStashHeader())

	// List stashes
	stashes, err := stash.List()
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	if len(stashes) == 0 {
		fmt.Print(ui.RenderNoStashes())
		return nil
	}

	// Show stash list
	fmt.Print(ui.RenderStashList(stashes, flagStashDays))

	// Show stats
	stats := stash.GetStats(stashes, flagStashDays)
	fmt.Print(ui.RenderStashStats(stats))

	if !flagStashClean {
		// Just listing, show tip
		if stats.OldCount > 0 {
			fmt.Print(ui.RenderStashTip())
		} else {
			fmt.Println()
		}
		return nil
	}

	// Clean mode
	if flagStashForce {
		// Delete all old stashes without asking
		var toDelete []int
		for _, s := range stashes {
			if s.IsOld(flagStashDays) {
				toDelete = append(toDelete, s.Index)
			}
		}

		if len(toDelete) == 0 {
			fmt.Printf("\n  %s No stashes older than %d days.\n\n", ui.CheckStyle.Render(), flagStashDays)
			return nil
		}

		dropped, errors := stash.DropMultiple(toDelete)
		fmt.Print(ui.RenderStashDropped(dropped, len(toDelete), errors))
		return nil
	}

	// Interactive picker
	toDelete, err := runStashPicker(stashes, flagStashDays)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	if toDelete == nil {
		// User cancelled
		return nil
	}

	if len(toDelete) == 0 {
		fmt.Printf("\n  %s No stashes selected.\n\n", ui.MutedStyle.Render("●"))
		return nil
	}

	dropped, errors := stash.DropMultiple(toDelete)
	fmt.Print(ui.RenderStashDropped(dropped, len(toDelete), errors))

	return nil
}

// Stash picker model

type stashPickerModel struct {
	stashes   []stash.Stash
	selected  map[int]bool
	cursor    int
	oldDays   int
	quitting  bool
	confirmed bool
}

func newStashPicker(stashes []stash.Stash, oldDays int) stashPickerModel {
	selected := make(map[int]bool)
	// Pre-select old stashes
	for _, s := range stashes {
		if s.IsOld(oldDays) {
			selected[s.Index] = true
		}
	}

	return stashPickerModel{
		stashes:  stashes,
		selected: selected,
		oldDays:  oldDays,
	}
}

func (m stashPickerModel) Init() tea.Cmd {
	return nil
}

func (m stashPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.cursor = len(m.stashes) - 1
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.stashes) {
				m.cursor = 0
			}

		case " ":
			// Toggle selection
			idx := m.stashes[m.cursor].Index
			m.selected[idx] = !m.selected[idx]

		case "a":
			// Select all
			for _, s := range m.stashes {
				m.selected[s.Index] = true
			}

		case "n":
			// Select none
			m.selected = make(map[int]bool)

		case "o":
			// Select only old
			m.selected = make(map[int]bool)
			for _, s := range m.stashes {
				if s.IsOld(m.oldDays) {
					m.selected[s.Index] = true
				}
			}
		}
	}

	return m, nil
}

func (m stashPickerModel) View() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n  %s\n", ui.MutedStyle.Render("Select stashes to delete:")))

	// Group by age
	var oldStashes, recentStashes []stash.Stash
	for _, s := range m.stashes {
		if s.IsOld(m.oldDays) {
			oldStashes = append(oldStashes, s)
		} else {
			recentStashes = append(recentStashes, s)
		}
	}

	// Sort by date (oldest first for old, newest first for recent)
	sort.Slice(oldStashes, func(i, j int) bool {
		return oldStashes[i].Date.Before(oldStashes[j].Date)
	})
	sort.Slice(recentStashes, func(i, j int) bool {
		return recentStashes[i].Date.After(recentStashes[j].Date)
	})

	cursorIdx := 0

	if len(oldStashes) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", ui.WarningStyle.Render(fmt.Sprintf("Old (>%d days)", m.oldDays))))
		for _, s := range oldStashes {
			b.WriteString(m.renderStashLine(s, cursorIdx))
			cursorIdx++
		}
	}

	if len(recentStashes) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", ui.MutedStyle.Render("Recent")))
		for _, s := range recentStashes {
			b.WriteString(m.renderStashLine(s, cursorIdx))
			cursorIdx++
		}
	}

	// Help
	b.WriteString(fmt.Sprintf("\n  %s\n", ui.Divider(55)))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		ui.HelpStyle.Render("␣ toggle · a all · o old · n none · ↵ confirm · q quit")))

	return b.String()
}

func (m stashPickerModel) renderStashLine(s stash.Stash, displayIdx int) string {
	cursor := "  "
	if displayIdx == m.cursor {
		cursor = ui.CursorStyle.Render() + " "
	}

	var checkbox string
	if m.selected[s.Index] {
		checkbox = ui.SuccessStyle.Render("▣")
	} else {
		checkbox = "▢"
	}

	name := s.Branch
	if displayIdx == m.cursor {
		name = ui.SelectedStyle.Render(name)
	}

	age := stash.FormatAge(s.Date)
	files := fmt.Sprintf("%d files", s.FileCount)

	return fmt.Sprintf("%s%s %s  %s  %s\n",
		cursor,
		checkbox,
		name,
		ui.MutedStyle.Render(files),
		ui.MutedStyle.Render(age))
}

func (m stashPickerModel) Cancelled() bool {
	return m.quitting
}

func (m stashPickerModel) SelectedIndices() []int {
	var indices []int
	for idx, selected := range m.selected {
		if selected {
			indices = append(indices, idx)
		}
	}
	return indices
}

func runStashPicker(stashes []stash.Stash, oldDays int) ([]int, error) {
	m := newStashPicker(stashes, oldDays)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(stashPickerModel)
	if fm.Cancelled() {
		return nil, nil
	}

	return fm.SelectedIndices(), nil
}
