package cmd

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// Tags picker model

type tagsPickerModel struct {
	tags      []tags.Tag
	selected  map[string]bool
	cursor    int
	quitting  bool
	confirmed bool
}

func newTagsPicker(tagList []tags.Tag) tagsPickerModel {
	selected := make(map[string]bool)
	// Pre-select all orphan tags
	for _, t := range tagList {
		selected[t.Name] = true
	}

	// Sort alphabetically
	sorted := make([]tags.Tag, len(tagList))
	copy(sorted, tagList)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return tagsPickerModel{
		tags:     sorted,
		selected: selected,
	}
}

func (m tagsPickerModel) Init() tea.Cmd {
	return nil
}

func (m tagsPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.cursor = len(m.tags) - 1
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.tags) {
				m.cursor = 0
			}

		case " ":
			// Toggle selection
			name := m.tags[m.cursor].Name
			m.selected[name] = !m.selected[name]

		case "a":
			// Select all
			for _, t := range m.tags {
				m.selected[t.Name] = true
			}

		case "n":
			// Select none
			m.selected = make(map[string]bool)
		}
	}

	return m, nil
}

func (m tagsPickerModel) View() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n  %s\n\n", ui.MutedStyle.Render("Select orphan tags to delete:")))

	for i, t := range m.tags {
		cursor := "  "
		if i == m.cursor {
			cursor = ui.CursorStyle.Render() + " "
		}

		var checkbox string
		if m.selected[t.Name] {
			checkbox = ui.SuccessStyle.Render("▣")
		} else {
			checkbox = "▢"
		}

		name := t.Name
		if i == m.cursor {
			name = ui.SelectedStyle.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, name))
	}

	// Help
	b.WriteString(fmt.Sprintf("\n  %s\n", ui.Divider(45)))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		ui.HelpStyle.Render("␣ toggle · a all · n none · ↵ confirm · q quit")))

	return b.String()
}

func (m tagsPickerModel) Cancelled() bool {
	return m.quitting
}

func (m tagsPickerModel) SelectedTags() []string {
	var names []string
	for name, selected := range m.selected {
		if selected {
			names = append(names, name)
		}
	}
	return names
}

func runTagsPicker(orphans []tags.Tag) ([]string, error) {
	m := newTagsPicker(orphans)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(tagsPickerModel)
	if fm.Cancelled() {
		return nil, nil
	}

	return fm.SelectedTags(), nil
}
