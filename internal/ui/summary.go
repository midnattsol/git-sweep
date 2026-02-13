package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/midnattsol/git-sweep/internal/git"
	"github.com/midnattsol/git-sweep/internal/sweep"
)

const defaultWidth = 50

// RenderHeader renders the app header
func RenderHeader() string {
	title := TitleStyle.Render("git-sweep")
	return fmt.Sprintf("\n  %s\n", title)
}

// RenderHeaderWithMode renders the header with mode integrated
func RenderHeaderWithMode(dryRun bool) string {
	title := TitleStyle.Render("git-sweep")
	sep := MutedStyle.Render(" · ")
	var mode string
	if dryRun {
		mode = DryRunStyle.Render("dry run")
	} else {
		mode = ExecuteStyle.Render("execute")
	}
	return fmt.Sprintf("\n  %s%s%s\n", title, sep, mode)
}

// RenderMode renders the current mode (dry run or execute)
func RenderMode(dryRun bool) string {
	var mode string
	if dryRun {
		mode = DryRunStyle.Render("DRY RUN")
	} else {
		mode = ExecuteStyle.Render("EXECUTE")
	}
	return fmt.Sprintf("\n  %s\n", mode)
}

// RenderBranchList renders the list of branches with their status
func RenderBranchList(result *sweep.Result, includeCandidates, showSkipped bool) string {
	var b strings.Builder

	// Eligible branches (merged PR)
	eligible := filterEligible(result.Branches)
	if len(eligible) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Would delete")))
		for _, br := range eligible {
			age := formatBranchAge(br.Branch.LastCommit)
			if age != "" {
				b.WriteString(fmt.Sprintf("  %s %s  %s  %s\n",
					CheckStyle.Render(),
					BranchStyle.Render(br.Branch.Name),
					MutedStyle.Render("merged PR"),
					MutedStyle.Render(age)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s  %s\n",
					CheckStyle.Render(),
					BranchStyle.Render(br.Branch.Name),
					MutedStyle.Render("merged PR")))
			}
		}
	}

	// Candidates (upstream gone, no merged PR)
	candidates := filterCandidates(result.Branches)
	if len(candidates) > 0 {
		if includeCandidates {
			b.WriteString(fmt.Sprintf("\n  %s\n", WarningStyle.Render("Candidates (--candidates)")))
		} else {
			b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Candidates (use --candidates to include)")))
		}
		for _, br := range candidates {
			age := formatBranchAge(br.Branch.LastCommit)
			icon := WarningStyle.Render("○")
			if !includeCandidates {
				icon = MutedStyle.Render("○")
			}
			if age != "" {
				b.WriteString(fmt.Sprintf("  %s %s  %s  %s\n",
					icon,
					br.Branch.Name,
					MutedStyle.Render("upstream gone"),
					MutedStyle.Render(age)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s  %s\n",
					icon,
					br.Branch.Name,
					MutedStyle.Render("upstream gone")))
			}
		}
	}

	// Skipped branches
	if showSkipped {
		skipped := filterSkipped(result.Branches)
		if len(skipped) > 0 {
			b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Skipped")))
			for _, br := range skipped {
				reason := formatSkipReason(br.Skip)
				age := formatBranchAge(br.Branch.LastCommit)
				if age != "" {
					b.WriteString(fmt.Sprintf("  %s %s  %s  %s\n",
						CircleStyle.Render(),
						br.Branch.Name,
						MutedStyle.Render(reason),
						MutedStyle.Render(age)))
				} else {
					b.WriteString(fmt.Sprintf("  %s %s  %s\n",
						CircleStyle.Render(),
						br.Branch.Name,
						MutedStyle.Render(reason)))
				}
			}
		}
	}

	return b.String()
}

// RenderSummary renders the summary stats
func RenderSummary(stats sweep.Stats) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Eligible %s", BoldStyle.Render(fmt.Sprintf("%d", stats.Eligible))))
	candidatesOnly := stats.Candidates - stats.Eligible
	parts = append(parts, fmt.Sprintf("Candidates %s", BoldStyle.Render(fmt.Sprintf("%d", candidatesOnly))))
	parts = append(parts, fmt.Sprintf("Skipped %s", BoldStyle.Render(fmt.Sprintf("%d", stats.Skipped))))

	content := strings.Join(parts, MutedStyle.Render(" · "))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1).
		Render(content)

	return fmt.Sprintf("\n%s\n", indent(box, 2))
}

// RenderDryRunTip renders the tip for dry run mode
func RenderDryRunTip() string {
	return fmt.Sprintf("\n  %s Remove %s to delete.\n\n",
		MutedStyle.Render("Tip:"),
		BoldStyle.Render("--dry-run"))
}

// RenderDeletedBranches renders the list of deleted branches
func RenderDeletedBranches(result *sweep.Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render("Deleted")))

	for _, br := range result.Branches {
		if br.Deleted {
			b.WriteString(fmt.Sprintf("  %s %s\n", CheckStyle.Render(), BranchStyle.Render(br.Branch.Name)))
		} else if br.DeleteErr != nil {
			b.WriteString(fmt.Sprintf("  %s %s  %s\n",
				CrossStyle.Render(),
				br.Branch.Name,
				ErrorStyle.Render(br.DeleteErr.Error())))
		}
	}

	return b.String()
}

// RenderDeleteSummary renders summary after deletion
func RenderDeleteSummary(deleted int, total int) string {
	content := fmt.Sprintf("Deleted %s of %s branches",
		SuccessStyle.Render(fmt.Sprintf("%d", deleted)),
		BoldStyle.Render(fmt.Sprintf("%d", total)))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1).
		Render(content)

	return fmt.Sprintf("\n%s\n\n", indent(box, 2))
}

// RenderNukeHeader renders the header for nuke mode
func RenderNukeHeader() string {
	title := TitleStyle.Render("git-sweep --nuke")
	return fmt.Sprintf("\n  %s\n", title)
}

// RenderNukeSummary renders summary after nuke deletion
func RenderNukeSummary(deleted int, total int) string {
	content := fmt.Sprintf("Deleted %s of %s branches",
		SuccessStyle.Render(fmt.Sprintf("%d", deleted)),
		BoldStyle.Render(fmt.Sprintf("%d", total)))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1).
		Render(content)

	return fmt.Sprintf("\n%s\n\n", indent(box, 2))
}

// RenderError renders an error message
func RenderError(msg string) string {
	return fmt.Sprintf("\n  %s %s\n\n", CrossStyle.Render(), ErrorStyle.Render(msg))
}

// RenderNoBranches renders message when no branches to delete
func RenderNoBranches() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("No branches to delete."))
}

// Helper functions

func filterEligible(branches []sweep.BranchResult) []sweep.BranchResult {
	var result []sweep.BranchResult
	for _, b := range branches {
		if b.Eligible {
			result = append(result, b)
		}
	}
	return result
}

func filterCandidates(branches []sweep.BranchResult) []sweep.BranchResult {
	var result []sweep.BranchResult
	for _, b := range branches {
		if b.Candidate {
			result = append(result, b)
		}
	}
	return result
}

func filterSkipped(branches []sweep.BranchResult) []sweep.BranchResult {
	var result []sweep.BranchResult
	for _, b := range branches {
		if b.Skip != git.SkipNone {
			result = append(result, b)
		}
	}
	return result
}

func formatSkipReason(reason git.SkipReason) string {
	return string(reason)
}

func formatBranchAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)

	switch {
	case d < time.Hour*24:
		return "today"
	case d < time.Hour*24*2:
		return "yesterday"
	case d < time.Hour*24*7:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < time.Hour*24*30:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case d < time.Hour*24*365:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}
