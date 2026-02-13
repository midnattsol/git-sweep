package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/midnattsol/git-sweep/internal/tags"
)

// RenderTagsHeader renders the header for tags command
func RenderTagsHeader() string {
	title := TitleStyle.Render("git-sweep tags")
	return fmt.Sprintf("\n  %s\n", title)
}

// RenderNoTags renders message when no tags found
func RenderNoTags() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("No local tags found."))
}

// RenderNoOrphanTags renders message when no orphan tags found
func RenderNoOrphanTags() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("All tags have remote counterparts."))
}

// RenderTagList renders the list of tags grouped by orphan status
func RenderTagList(tagList []tags.Tag) string {
	var b strings.Builder

	// Group by orphan status
	var orphans, synced []tags.Tag
	for _, t := range tagList {
		if t.IsOrphan {
			orphans = append(orphans, t)
		} else {
			synced = append(synced, t)
		}
	}

	// Sort alphabetically
	sort.Slice(orphans, func(i, j int) bool {
		return orphans[i].Name < orphans[j].Name
	})
	sort.Slice(synced, func(i, j int) bool {
		return synced[i].Name < synced[j].Name
	})

	if len(orphans) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", WarningStyle.Render("Orphan tags (no remote)")))
		for _, t := range orphans {
			b.WriteString(fmt.Sprintf("  %s %s\n", WarningStyle.Render("○"), t.Name))
		}
	}

	if len(synced) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Synced tags")))
		for _, t := range synced {
			b.WriteString(fmt.Sprintf("  %s %s\n", MutedStyle.Render("○"), t.Name))
		}
	}

	return b.String()
}

// RenderTagStats renders tag statistics
func RenderTagStats(stats tags.Stats) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Total %s", BoldStyle.Render(fmt.Sprintf("%d", stats.Total))))

	if stats.Orphans > 0 {
		parts = append(parts, fmt.Sprintf("Orphans %s", WarningStyle.Render(fmt.Sprintf("%d", stats.Orphans))))
	}

	content := strings.Join(parts, MutedStyle.Render(" · "))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1).
		Render(content)

	return fmt.Sprintf("\n%s\n", indent(box, 2))
}

// RenderTagsTip renders the tip for tags command
func RenderTagsTip() string {
	return fmt.Sprintf("\n  %s Run %s to clean up orphan tags.\n\n",
		MutedStyle.Render("Tip:"),
		BoldStyle.Render("git sweep tags --clean"))
}

// RenderTagsDeleted renders the result of deleting tags
func RenderTagsDeleted(deleted int, total int, errors []error) string {
	var b strings.Builder

	if deleted > 0 {
		b.WriteString(fmt.Sprintf("\n  %s Deleted %d tag(s)\n",
			CheckStyle.Render(),
			deleted))
	}

	for _, err := range errors {
		b.WriteString(fmt.Sprintf("  %s %s\n", CrossStyle.Render(), ErrorStyle.Render(err.Error())))
	}

	b.WriteString("\n")
	return b.String()
}
