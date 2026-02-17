package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderNukeHeader renders the header for branch cleanup.
func RenderNukeHeader() string {
	title := TitleStyle.Render("git-sweep")
	return fmt.Sprintf("\n  %s\n", title)
}

// RenderNukeSummary renders summary after deletion.
func RenderNukeSummary(deleted int, total int) string {
	content := fmt.Sprintf("Deleted %s of %s branches",
		SuccessStyle.Render(fmt.Sprintf("%d", deleted)),
		BoldStyle.Render(fmt.Sprintf("%d", total)))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1).
		Render(content)

	return fmt.Sprintf("\n%s\n\n", Indent(box, 2))
}

// RenderError renders an error message.
func RenderError(msg string) string {
	return fmt.Sprintf("\n  %s %s\n\n", CrossStyle.Render(), ErrorStyle.Render(msg))
}

// RenderNoBranches renders message when no branches are available for deletion.
func RenderNoBranches() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("No branches to delete."))
}
