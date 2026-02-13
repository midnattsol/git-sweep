package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/ui"
	"github.com/midnattsol/git-sweep/internal/worktree"
)

var (
	flagWorktreeClean bool
	flagWorktreeForce bool
)

func NewWorktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Clean up broken git worktrees",
		Long: `List and clean up broken git worktrees.

Shows all worktrees and highlights broken ones (paths that no longer exist).

Examples:
  git sweep worktree              # List all worktrees, show status
  git sweep worktree --clean      # Interactive picker to remove broken worktrees
  git sweep worktree --force      # Remove all broken worktrees without asking`,
		RunE: runWorktree,
	}

	cmd.Flags().BoolVar(&flagWorktreeClean, "clean", false, "Interactive picker to remove broken worktrees")
	cmd.Flags().BoolVar(&flagWorktreeForce, "force", false, "Remove all broken worktrees without confirmation")

	return cmd
}

func runWorktree(cmd *cobra.Command, args []string) error {
	fmt.Print(ui.RenderWorktreeHeader())

	// List worktrees
	worktrees, err := worktree.List()
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Filter out main worktree for display purposes (can't delete main)
	var displayWorktrees []worktree.Worktree
	for _, w := range worktrees {
		displayWorktrees = append(displayWorktrees, w)
	}

	if len(displayWorktrees) <= 1 {
		// Only main worktree exists
		fmt.Print(ui.RenderNoWorktrees())
		return nil
	}

	// Show worktree list
	fmt.Print(ui.RenderWorktreeList(displayWorktrees))

	// Show stats
	stats := worktree.GetStats(displayWorktrees)
	fmt.Print(ui.RenderWorktreeStats(stats))

	if !flagWorktreeClean && !flagWorktreeForce {
		// Just listing
		if stats.Broken > 0 {
			fmt.Print(ui.RenderWorktreeTip())
		} else {
			fmt.Print(ui.RenderNoBrokenWorktrees())
		}
		return nil
	}

	// Get broken worktrees (excluding main)
	var broken []worktree.Worktree
	for _, w := range worktrees {
		if w.IsBroken && !w.IsMain {
			broken = append(broken, w)
		}
	}

	if len(broken) == 0 {
		fmt.Print(ui.RenderNoBrokenWorktrees())
		return nil
	}

	if flagWorktreeForce {
		// Prune worktrees without asking
		if err := worktree.Prune(); err != nil {
			fmt.Print(ui.RenderError(err.Error()))
			return err
		}
		fmt.Print(ui.RenderWorktreePruned())
		return nil
	}

	// Interactive picker
	toRemove, err := runWorktreePicker(broken)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	if toRemove == nil {
		// User cancelled
		return nil
	}

	if len(toRemove) == 0 {
		fmt.Printf("\n  %s No worktrees selected.\n\n", ui.MutedStyle.Render("●"))
		return nil
	}

	// For broken worktrees, use prune instead of individual removal
	if err := worktree.Prune(); err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	fmt.Print(ui.RenderWorktreeRemoved(len(toRemove), len(toRemove), nil))

	return nil
}

// Worktree picker model

type worktreePickerModel struct {
	worktrees []worktree.Worktree
	selected  map[string]bool
	cursor    int
	quitting  bool
	confirmed bool
}

func newWorktreePicker(worktrees []worktree.Worktree) worktreePickerModel {
	selected := make(map[string]bool)
	// Pre-select all broken worktrees
	for _, w := range worktrees {
		selected[w.Path] = true
	}

	return worktreePickerModel{
		worktrees: worktrees,
		selected:  selected,
	}
}

func (m worktreePickerModel) Init() tea.Cmd {
	return nil
}

func (m worktreePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.cursor = len(m.worktrees) - 1
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.worktrees) {
				m.cursor = 0
			}

		case " ":
			// Toggle selection
			path := m.worktrees[m.cursor].Path
			m.selected[path] = !m.selected[path]

		case "a":
			// Select all
			for _, w := range m.worktrees {
				m.selected[w.Path] = true
			}

		case "n":
			// Select none
			m.selected = make(map[string]bool)
		}
	}

	return m, nil
}

func (m worktreePickerModel) View() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n  %s\n\n", ui.MutedStyle.Render("Select broken worktrees to remove:")))

	for i, w := range m.worktrees {
		cursor := "  "
		if i == m.cursor {
			cursor = ui.CursorStyle.Render() + " "
		}

		var checkbox string
		if m.selected[w.Path] {
			checkbox = ui.SuccessStyle.Render("▣")
		} else {
			checkbox = "▢"
		}

		path := worktree.ShortenPath(w.Path)
		if i == m.cursor {
			path = ui.SelectedStyle.Render(path)
		}

		info := ""
		if w.Branch != "" {
			info = fmt.Sprintf("  %s", ui.MutedStyle.Render(w.Branch))
		}

		b.WriteString(fmt.Sprintf("%s%s %s%s\n", cursor, checkbox, path, info))
	}

	// Help
	b.WriteString(fmt.Sprintf("\n  %s\n", ui.Divider(45)))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		ui.HelpStyle.Render("␣ toggle · a all · n none · ↵ confirm · q quit")))

	return b.String()
}

func (m worktreePickerModel) Cancelled() bool {
	return m.quitting
}

func (m worktreePickerModel) SelectedPaths() []string {
	var paths []string
	for path, selected := range m.selected {
		if selected {
			paths = append(paths, path)
		}
	}
	return paths
}

func runWorktreePicker(broken []worktree.Worktree) ([]string, error) {
	m := newWorktreePicker(broken)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(worktreePickerModel)
	if fm.Cancelled() {
		return nil, nil
	}

	return fm.SelectedPaths(), nil
}
