package cmd

import (
	"fmt"

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

	if len(worktrees) <= 1 {
		// Only main worktree exists
		fmt.Print(ui.RenderNoWorktrees())
		return nil
	}

	// Show worktree list
	fmt.Print(ui.RenderWorktreeList(worktrees))

	// Show stats
	stats := worktree.GetStats(worktrees)
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

// worktreeItem implements ui.MultiPickerItem for worktrees.
type worktreeItem struct {
	wt worktree.Worktree
}

func (w worktreeItem) Key() string {
	return w.wt.Path
}

func (w worktreeItem) Label() string {
	return worktree.ShortenPath(w.wt.Path)
}

func (w worktreeItem) Details() string {
	return w.wt.Branch
}

func runWorktreePicker(broken []worktree.Worktree) ([]string, error) {
	// Build items
	var items []ui.MultiPickerItem
	for _, w := range broken {
		items = append(items, worktreeItem{wt: w})
	}

	// Pre-select all broken worktrees
	preSelected := make(map[string]bool)
	for _, w := range broken {
		preSelected[w.Path] = true
	}

	return ui.RunMultiPicker(ui.MultiPickerConfig{
		Title: "Select broken worktrees to remove:",
		Groups: []ui.MultiPickerGroup{
			{Items: items},
		},
		PreSelected: preSelected,
	})
}
