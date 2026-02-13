package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/midnattsol/git-sweep/internal/worktree"
)

// RenderWorktreeHeader renders the header for worktree command
func RenderWorktreeHeader() string {
	title := TitleStyle.Render("git-sweep worktree")
	return fmt.Sprintf("\n  %s\n", title)
}

// RenderNoWorktrees renders message when no worktrees found
func RenderNoWorktrees() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("No additional worktrees found."))
}

// RenderNoBrokenWorktrees renders message when no broken worktrees found
func RenderNoBrokenWorktrees() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("All worktrees are healthy."))
}

// RenderWorktreeList renders the list of worktrees grouped by status
func RenderWorktreeList(worktrees []worktree.Worktree) string {
	var b strings.Builder

	// Group by status
	var broken, locked, healthy []worktree.Worktree
	for _, w := range worktrees {
		if w.IsBroken {
			broken = append(broken, w)
		} else if w.IsLocked {
			locked = append(locked, w)
		} else {
			healthy = append(healthy, w)
		}
	}

	if len(broken) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", WarningStyle.Render("Broken worktrees")))
		for _, w := range broken {
			b.WriteString(renderWorktreeLine(w, true))
		}
	}

	if len(locked) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Locked worktrees")))
		for _, w := range locked {
			b.WriteString(renderWorktreeLine(w, false))
		}
	}

	if len(healthy) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Healthy worktrees")))
		for _, w := range healthy {
			b.WriteString(renderWorktreeLine(w, false))
		}
	}

	return b.String()
}

func renderWorktreeLine(w worktree.Worktree, isBroken bool) string {
	var icon string
	if isBroken {
		icon = WarningStyle.Render("○")
	} else if w.IsMain {
		icon = SuccessStyle.Render("●")
	} else {
		icon = MutedStyle.Render("○")
	}

	// Build info parts
	var info []string

	if w.Branch != "" {
		info = append(info, BranchStyle.Render(w.Branch))
	}

	if w.IsMain {
		info = append(info, MutedStyle.Render("(main)"))
	}

	if w.IsLocked {
		info = append(info, WarningStyle.Render("(locked)"))
	}

	if w.Reason != "" {
		info = append(info, ErrorStyle.Render("("+w.Reason+")"))
	}

	path := worktree.ShortenPath(w.Path)

	return fmt.Sprintf("  %s %s  %s\n",
		icon,
		path,
		strings.Join(info, " "))
}

// RenderWorktreeStats renders worktree statistics
func RenderWorktreeStats(stats worktree.Stats) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Total %s", BoldStyle.Render(fmt.Sprintf("%d", stats.Total))))

	if stats.Broken > 0 {
		parts = append(parts, fmt.Sprintf("Broken %s", WarningStyle.Render(fmt.Sprintf("%d", stats.Broken))))
	}

	if stats.Locked > 0 {
		parts = append(parts, fmt.Sprintf("Locked %s", MutedStyle.Render(fmt.Sprintf("%d", stats.Locked))))
	}

	content := strings.Join(parts, MutedStyle.Render(" · "))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1).
		Render(content)

	return fmt.Sprintf("\n%s\n", indent(box, 2))
}

// RenderWorktreeTip renders the tip for worktree command
func RenderWorktreeTip() string {
	return fmt.Sprintf("\n  %s Run %s to clean up broken worktrees.\n\n",
		MutedStyle.Render("Tip:"),
		BoldStyle.Render("git sweep worktree --clean"))
}

// RenderWorktreeRemoved renders the result of removing worktrees
func RenderWorktreeRemoved(removed int, total int, errors []error) string {
	var b strings.Builder

	if removed > 0 {
		b.WriteString(fmt.Sprintf("\n  %s Removed %d worktree(s)\n",
			CheckStyle.Render(),
			removed))
	}

	for _, err := range errors {
		b.WriteString(fmt.Sprintf("  %s %s\n", CrossStyle.Render(), ErrorStyle.Render(err.Error())))
	}

	b.WriteString("\n")
	return b.String()
}

// RenderWorktreePruned renders the result of pruning worktrees
func RenderWorktreePruned() string {
	return fmt.Sprintf("\n  %s Pruned worktree admin files\n\n", CheckStyle.Render())
}
