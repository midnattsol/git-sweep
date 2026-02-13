package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/tags"
	"github.com/midnattsol/git-sweep/internal/ui"
)

var (
	flagTagsClean  bool
	flagTagsForce  bool
	flagTagsRemote string
)

func NewTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Clean up orphan git tags",
		Long: `List and clean up orphan git tags (local tags without remote counterpart).

Shows all local tags and highlights orphans that exist locally but not on remote.

Examples:
  git sweep tags              # List all tags, show orphans
  git sweep tags --clean      # Interactive picker to delete orphan tags
  git sweep tags --force      # Delete all orphan tags without asking`,
		RunE: runTags,
	}

	cmd.Flags().BoolVar(&flagTagsClean, "clean", false, "Interactive picker to delete orphan tags")
	cmd.Flags().BoolVar(&flagTagsForce, "force", false, "Delete all orphan tags without confirmation")
	cmd.Flags().StringVar(&flagTagsRemote, "remote", "origin", "Remote to check tags against")

	return cmd
}

func runTags(cmd *cobra.Command, args []string) error {
	fmt.Print(ui.RenderTagsHeader())

	// List tags
	tagList, err := tags.List(flagTagsRemote)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	if len(tagList) == 0 {
		fmt.Print(ui.RenderNoTags())
		return nil
	}

	// Show tag list
	fmt.Print(ui.RenderTagList(tagList))

	// Show stats
	stats := tags.GetStats(tagList)
	fmt.Print(ui.RenderTagStats(stats))

	if !flagTagsClean && !flagTagsForce {
		// Just listing
		if stats.Orphans > 0 {
			fmt.Print(ui.RenderTagsTip())
		} else {
			fmt.Print(ui.RenderNoOrphanTags())
		}
		return nil
	}

	// Get orphan tags
	orphans := tags.GetOrphans(tagList)
	if len(orphans) == 0 {
		fmt.Print(ui.RenderNoOrphanTags())
		return nil
	}

	if flagTagsForce {
		// Delete all orphans without asking
		var toDelete []string
		for _, t := range orphans {
			toDelete = append(toDelete, t.Name)
		}

		deleted, errors := tags.DeleteMultiple(toDelete)
		fmt.Print(ui.RenderTagsDeleted(deleted, len(toDelete), errors))
		return nil
	}

	// Interactive picker
	toDelete, err := runTagsPicker(orphans)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	if toDelete == nil {
		// User cancelled
		return nil
	}

	if len(toDelete) == 0 {
		fmt.Printf("\n  %s No tags selected.\n\n", ui.MutedStyle.Render("●"))
		return nil
	}

	deleted, errors := tags.DeleteMultiple(toDelete)
	fmt.Print(ui.RenderTagsDeleted(deleted, len(toDelete), errors))

	return nil
}

// tagItem implements ui.MultiPickerItem for tags.
type tagItem struct {
	tag tags.Tag
}

func (t tagItem) Key() string {
	return t.tag.Name
}

func (t tagItem) Label() string {
	return t.tag.Name
}

func (t tagItem) Details() string {
	return ""
}

func runTagsPicker(orphans []tags.Tag) ([]string, error) {
	// Sort alphabetically
	sorted := make([]tags.Tag, len(orphans))
	copy(sorted, orphans)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	// Build items
	var items []ui.MultiPickerItem
	for _, t := range sorted {
		items = append(items, tagItem{tag: t})
	}

	// Pre-select all orphan tags
	preSelected := make(map[string]bool)
	for _, t := range sorted {
		preSelected[t.Name] = true
	}

	return ui.RunMultiPicker(ui.MultiPickerConfig{
		Title: "Select orphan tags to delete:",
		Groups: []ui.MultiPickerGroup{
			{Items: items},
		},
		PreSelected: preSelected,
	})
}
