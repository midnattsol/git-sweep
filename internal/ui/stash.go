package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/midnattsol/git-sweep/internal/stash"
	"github.com/midnattsol/git-sweep/internal/timeutil"
)

// RenderStashHeader renders the header for stash command
func RenderStashHeader() string {
	title := TitleStyle.Render("git-sweep stash")
	return fmt.Sprintf("\n  %s\n", title)
}

// RenderNoStashes renders message when no stashes found
func RenderNoStashes() string {
	return fmt.Sprintf("\n  %s %s\n\n", CheckStyle.Render(), MutedStyle.Render("No stashes found."))
}

// RenderStashList renders the list of stashes grouped by age
func RenderStashList(stashes []stash.Stash, oldDays int) string {
	var b strings.Builder

	// Group by age category
	type ageGroup struct {
		label   string
		stashes []stash.Stash
		isOld   bool
	}

	// Create time buckets
	now := time.Now()
	oldThreshold := now.AddDate(0, 0, -oldDays)
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	var old, lastMonth, lastWeek, recent []stash.Stash

	for _, s := range stashes {
		if s.Date.IsZero() {
			old = append(old, s)
		} else if s.Date.Before(oldThreshold) {
			old = append(old, s)
		} else if s.Date.Before(monthAgo) {
			lastMonth = append(lastMonth, s)
		} else if s.Date.Before(weekAgo) {
			lastWeek = append(lastWeek, s)
		} else {
			recent = append(recent, s)
		}
	}

	groups := []ageGroup{
		{label: fmt.Sprintf("Old (>%d days)", oldDays), stashes: old, isOld: true},
		{label: "Last month", stashes: lastMonth, isOld: false},
		{label: "Last week", stashes: lastWeek, isOld: false},
		{label: "Recent", stashes: recent, isOld: false},
	}

	for _, g := range groups {
		if len(g.stashes) == 0 {
			continue
		}

		// Sort by date (oldest first for old, newest first for others)
		sort.Slice(g.stashes, func(i, j int) bool {
			if g.isOld {
				return g.stashes[i].Date.Before(g.stashes[j].Date)
			}
			return g.stashes[i].Date.After(g.stashes[j].Date)
		})

		if g.isOld {
			b.WriteString(fmt.Sprintf("\n  %s\n", WarningStyle.Render(g.label)))
		} else {
			b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render(g.label)))
		}

		for _, s := range g.stashes {
			icon := MutedStyle.Render("○")
			if g.isOld {
				icon = WarningStyle.Render("○")
			}

			age := timeutil.FormatAge(s.Date)
			files := formatStashFiles(s.Files, 2)

			b.WriteString(fmt.Sprintf("  %s %s  %s  %s",
				icon,
				s.Branch,
				MutedStyle.Render(fmt.Sprintf("%d files", s.FileCount)),
				MutedStyle.Render(age)))

			if files != "" {
				b.WriteString(fmt.Sprintf("  %s", MutedStyle.Render("("+files+")")))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func formatStashFiles(files []string, max int) string {
	if len(files) == 0 {
		return ""
	}

	// Get just filenames, not full paths
	short := make([]string, 0, len(files))
	for _, f := range files {
		parts := strings.Split(f, "/")
		short = append(short, parts[len(parts)-1])
	}

	if len(short) <= max {
		return strings.Join(short, ", ")
	}

	return fmt.Sprintf("%s +%d", strings.Join(short[:max], ", "), len(short)-max)
}

// RenderStashStats renders stash statistics
func RenderStashStats(stats stash.Stats) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Total %s", BoldStyle.Render(fmt.Sprintf("%d", stats.Total))))
	if stats.OldCount > 0 {
		parts = append(parts, fmt.Sprintf("Old %s", WarningStyle.Render(fmt.Sprintf("%d", stats.OldCount))))
	}
	return RenderStatsBox(parts)
}

// RenderStashTip renders the tip for stash command
func RenderStashTip() string {
	return fmt.Sprintf("\n  %s Run %s to clean up old stashes.\n\n",
		MutedStyle.Render("Tip:"),
		BoldStyle.Render("git sweep stash --clean"))
}

// RenderStashDropped renders the result of dropping stashes
func RenderStashDropped(dropped int, total int, errors []error) string {
	var b strings.Builder

	if dropped > 0 {
		b.WriteString(fmt.Sprintf("\n  %s Dropped %d stash(es)\n",
			CheckStyle.Render(),
			dropped))
	}

	for _, err := range errors {
		b.WriteString(fmt.Sprintf("  %s %s\n", CrossStyle.Render(), ErrorStyle.Render(err.Error())))
	}

	b.WriteString("\n")
	return b.String()
}
