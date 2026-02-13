package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/stash"
	"github.com/midnattsol/git-sweep/internal/timeutil"
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

// stashItem implements ui.MultiPickerItem for stashes.
type stashItem struct {
	stash stash.Stash
}

func (s stashItem) Key() string {
	return strconv.Itoa(s.stash.Index)
}

func (s stashItem) Label() string {
	return s.stash.Branch
}

func (s stashItem) Details() string {
	age := timeutil.FormatAge(s.stash.Date)
	return fmt.Sprintf("%d files · %s", s.stash.FileCount, age)
}

func runStashPicker(stashes []stash.Stash, oldDays int) ([]int, error) {
	// Group by age
	var oldStashes, recentStashes []stash.Stash
	for _, s := range stashes {
		if s.IsOld(oldDays) {
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

	// Build groups
	var groups []ui.MultiPickerGroup

	if len(oldStashes) > 0 {
		var items []ui.MultiPickerItem
		for _, s := range oldStashes {
			items = append(items, stashItem{stash: s})
		}
		groups = append(groups, ui.MultiPickerGroup{
			Header: fmt.Sprintf("Old (>%d days)", oldDays),
			Style:  ui.WarningStyle.Render,
			Items:  items,
		})
	}

	if len(recentStashes) > 0 {
		var items []ui.MultiPickerItem
		for _, s := range recentStashes {
			items = append(items, stashItem{stash: s})
		}
		groups = append(groups, ui.MultiPickerGroup{
			Header: "Recent",
			Style:  ui.MutedStyle.Render,
			Items:  items,
		})
	}

	// Pre-select old stashes
	preSelected := make(map[string]bool)
	for _, s := range oldStashes {
		preSelected[strconv.Itoa(s.Index)] = true
	}

	// Extra key handler for "o" (select only old)
	extraKeys := map[string]func(m *ui.MultiPickerModel){
		"o": func(m *ui.MultiPickerModel) {
			m.ClearSelection()
			for _, s := range oldStashes {
				m.SetSelected(strconv.Itoa(s.Index), true)
			}
		},
	}

	selected, err := ui.RunMultiPicker(ui.MultiPickerConfig{
		Title:       "Select stashes to delete:",
		Groups:      groups,
		PreSelected: preSelected,
		ExtraKeys:   extraKeys,
		ExtraHelp:   "o old",
	})
	if err != nil {
		return nil, err
	}
	if selected == nil {
		return nil, nil
	}

	// Convert string keys back to indices
	var indices []int
	for _, key := range selected {
		idx, _ := strconv.Atoi(key)
		indices = append(indices, idx)
	}
	return indices, nil
}
