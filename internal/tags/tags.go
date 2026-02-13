// Package tags provides functionality to list and clean orphan git tags
package tags

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Tag represents a git tag
type Tag struct {
	Name     string
	IsOrphan bool // Local tag without remote counterpart
}

// Stats holds tag statistics
type Stats struct {
	Total   int
	Orphans int
}

// List returns all local tags with orphan status
func List(remote string) ([]Tag, error) {
	if remote == "" {
		remote = "origin"
	}

	// Get local tags
	localTags, err := getLocalTags()
	if err != nil {
		return nil, err
	}

	if len(localTags) == 0 {
		return nil, nil
	}

	// Get remote tags
	remoteTags, err := getRemoteTags(remote)
	if err != nil {
		// If we can't get remote tags, treat all as non-orphan
		var tags []Tag
		for _, name := range localTags {
			tags = append(tags, Tag{Name: name, IsOrphan: false})
		}
		return tags, nil
	}

	// Build lookup set for remote tags
	remoteSet := make(map[string]bool)
	for _, name := range remoteTags {
		remoteSet[name] = true
	}

	// Mark orphan tags
	var tags []Tag
	for _, name := range localTags {
		tags = append(tags, Tag{
			Name:     name,
			IsOrphan: !remoteSet[name],
		})
	}

	return tags, nil
}

// getLocalTags returns all local tag names
func getLocalTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list local tags: %w", err)
	}

	return parseLines(out), nil
}

// getRemoteTags returns all remote tag names
func getRemoteTags(remote string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", remote)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote tags: %w", err)
	}

	var tags []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		// Format: SHA\trefs/tags/name or refs/tags/name^{}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		ref := parts[1]
		// Skip dereferenced tags (^{})
		if strings.HasSuffix(ref, "^{}") {
			continue
		}

		// Extract tag name from refs/tags/name
		name := strings.TrimPrefix(ref, "refs/tags/")
		if name != "" {
			tags = append(tags, name)
		}
	}

	return tags, nil
}

func parseLines(data []byte) []string {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// GetOrphans returns only orphan tags
func GetOrphans(tags []Tag) []Tag {
	var orphans []Tag
	for _, t := range tags {
		if t.IsOrphan {
			orphans = append(orphans, t)
		}
	}
	return orphans
}

// Delete removes a local tag
func Delete(name string) error {
	cmd := exec.Command("git", "tag", "-d", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete tag %s: %w", name, err)
	}
	return nil
}

// DeleteMultiple removes multiple tags
func DeleteMultiple(names []string) (deleted int, errors []error) {
	for _, name := range names {
		if err := Delete(name); err != nil {
			errors = append(errors, err)
		} else {
			deleted++
		}
	}
	return deleted, errors
}

// GetStats returns statistics about tags
func GetStats(tags []Tag) Stats {
	stats := Stats{
		Total: len(tags),
	}

	for _, t := range tags {
		if t.IsOrphan {
			stats.Orphans++
		}
	}

	return stats
}
