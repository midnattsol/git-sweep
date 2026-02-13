// Package worktree provides functionality to list and clean git worktrees
package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path     string
	Branch   string
	Commit   string // HEAD commit SHA
	IsBroken bool   // Path doesn't exist or is invalid
	IsMain   bool   // Main worktree (the original repo)
	IsBare   bool   // Bare worktree
	IsLocked bool   // Locked worktree
	Reason   string // Reason for broken status
}

// Stats holds worktree statistics
type Stats struct {
	Total  int
	Broken int
	Locked int
}

// List returns all worktrees with their status
func List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktrees(out), nil
}

func parseWorktrees(data []byte) []Worktree {
	var worktrees []Worktree
	var current *Worktree

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "worktree "):
			// New worktree entry
			if current != nil {
				worktrees = append(worktrees, *current)
			}
			path := strings.TrimPrefix(line, "worktree ")
			current = &Worktree{Path: path}

			// Check if path exists
			if !pathExists(path) {
				current.IsBroken = true
				current.Reason = "path does not exist"
			}

		case strings.HasPrefix(line, "HEAD "):
			if current != nil {
				current.Commit = strings.TrimPrefix(line, "HEAD ")
			}

		case strings.HasPrefix(line, "branch "):
			if current != nil {
				ref := strings.TrimPrefix(line, "branch ")
				// Extract branch name from refs/heads/branch-name
				current.Branch = strings.TrimPrefix(ref, "refs/heads/")
			}

		case line == "bare":
			if current != nil {
				current.IsBare = true
				current.IsMain = true
			}

		case line == "detached":
			if current != nil {
				current.Branch = "(detached)"
			}

		case line == "locked":
			if current != nil {
				current.IsLocked = true
			}

		case line == "":
			// Empty line separates entries
			continue
		}
	}

	// Don't forget the last entry
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	// Mark the first worktree as main (if not bare)
	if len(worktrees) > 0 && !worktrees[0].IsBare {
		worktrees[0].IsMain = true
	}

	return worktrees
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetBroken returns only broken worktrees
func GetBroken(worktrees []Worktree) []Worktree {
	var broken []Worktree
	for _, w := range worktrees {
		if w.IsBroken {
			broken = append(broken, w)
		}
	}
	return broken
}

// Remove removes a worktree by path
func Remove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree %s: %w", path, err)
	}
	return nil
}

// RemoveMultiple removes multiple worktrees
func RemoveMultiple(paths []string, force bool) (removed int, errors []error) {
	for _, path := range paths {
		if err := Remove(path, force); err != nil {
			errors = append(errors, err)
		} else {
			removed++
		}
	}
	return removed, errors
}

// Prune removes worktree information for worktrees whose paths no longer exist
func Prune() error {
	cmd := exec.Command("git", "worktree", "prune")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}
	return nil
}

// GetStats returns statistics about worktrees
func GetStats(worktrees []Worktree) Stats {
	stats := Stats{
		Total: len(worktrees),
	}

	for _, w := range worktrees {
		if w.IsBroken {
			stats.Broken++
		}
		if w.IsLocked {
			stats.Locked++
		}
	}

	return stats
}

// ShortenPath returns a shortened version of the path for display
func ShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}

	return path
}

// GetBasename returns the last component of the path
func GetBasename(path string) string {
	return filepath.Base(path)
}
