// Package stash provides functionality to list and clean git stashes
package stash

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Stash represents a single git stash entry
type Stash struct {
	Index     int
	Ref       string    // stash@{0}
	Branch    string    // Branch where stash was created
	Message   string    // Stash message
	Date      time.Time // When the stash was created
	Files     []string  // Files in the stash
	FileCount int
}

// Stats holds stash statistics
type Stats struct {
	Total    int
	OldCount int // Stashes older than threshold
	OldDays  int // Threshold in days
}

// List returns all stashes with detailed information
func List() ([]Stash, error) {
	// Get stash list with date and message
	// Format: %gd = reflog selector, %gs = reflog subject, %ci = committer date
	cmd := exec.Command("git", "stash", "list", "--format=%gd|%gs|%ci")
	out, err := cmd.Output()
	if err != nil {
		// No stashes is not an error
		if len(out) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list stashes: %w", err)
	}

	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}

	var stashes []Stash
	scanner := bufio.NewScanner(bytes.NewReader(out))

	// Regex to extract branch from "WIP on branch: ..." or "On branch: ..."
	branchRegex := regexp.MustCompile(`(?:WIP on|On) ([^:]+):`)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}

		ref := parts[0]     // stash@{0}
		message := parts[1] // WIP on branch: commit msg
		dateStr := strings.TrimSpace(parts[2])

		// Parse index from ref
		index := 0
		if _, err := fmt.Sscanf(ref, "stash@{%d}", &index); err != nil {
			continue
		}

		// Parse date
		date, _ := time.Parse("2006-01-02 15:04:05 -0700", dateStr)

		// Extract branch from message
		branch := "unknown"
		if matches := branchRegex.FindStringSubmatch(message); len(matches) > 1 {
			branch = matches[1]
		}

		// Get files in this stash
		files, _ := getStashFiles(ref)

		stashes = append(stashes, Stash{
			Index:     index,
			Ref:       ref,
			Branch:    branch,
			Message:   message,
			Date:      date,
			Files:     files,
			FileCount: len(files),
		})
	}

	return stashes, nil
}

// getStashFiles returns the list of files modified in a stash
func getStashFiles(ref string) ([]string, error) {
	cmd := exec.Command("git", "stash", "show", ref, "--name-only")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" {
			files = append(files, file)
		}
	}

	return files, nil
}

// Drop removes a stash by its index
func Drop(index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	cmd := exec.Command("git", "stash", "drop", ref)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to drop %s: %w", ref, err)
	}
	return nil
}

// DropMultiple removes multiple stashes by their indices (in reverse order to maintain indices)
func DropMultiple(indices []int) (dropped int, errors []error) {
	// Sort indices in descending order to avoid index shifting
	sortedIndices := make([]int, len(indices))
	copy(sortedIndices, indices)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedIndices)))

	for _, idx := range sortedIndices {
		if err := Drop(idx); err != nil {
			errors = append(errors, err)
		} else {
			dropped++
		}
	}

	return dropped, errors
}

// GetStats returns statistics about stashes
func GetStats(stashes []Stash, oldDays int) Stats {
	stats := Stats{
		Total:   len(stashes),
		OldDays: oldDays,
	}

	threshold := time.Now().AddDate(0, 0, -oldDays)
	for _, s := range stashes {
		if !s.Date.IsZero() && s.Date.Before(threshold) {
			stats.OldCount++
		}
	}

	return stats
}

// IsOld returns true if the stash is older than the given number of days
func (s *Stash) IsOld(days int) bool {
	if s.Date.IsZero() {
		return false
	}
	threshold := time.Now().AddDate(0, 0, -days)
	return s.Date.Before(threshold)
}

// FormatFiles returns a truncated list of files for display
func (s *Stash) FormatFiles(max int) string {
	if len(s.Files) == 0 {
		return ""
	}

	if len(s.Files) <= max {
		return strings.Join(s.Files, ", ")
	}

	shown := strings.Join(s.Files[:max], ", ")
	return fmt.Sprintf("%s, +%d more", shown, len(s.Files)-max)
}

// ParseStashIndex extracts the index from a stash ref like "stash@{2}"
func ParseStashIndex(ref string) (int, error) {
	var index int
	_, err := fmt.Sscanf(ref, "stash@{%d}", &index)
	if err != nil {
		// Try just the number
		index, err = strconv.Atoi(ref)
		if err != nil {
			return 0, fmt.Errorf("invalid stash reference: %s", ref)
		}
	}
	return index, nil
}
