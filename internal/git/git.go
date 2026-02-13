package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Branch represents a local git branch
type Branch struct {
	Name         string
	Upstream     string
	IsCurrent    bool
	UpstreamGone bool
	LastCommit   time.Time
}

// SkipReason explains why a branch was skipped
type SkipReason string

const (
	SkipNone           SkipReason = ""
	SkipCurrent        SkipReason = "current branch"
	SkipProtected      SkipReason = "protected"
	SkipNoUpstream     SkipReason = "no upstream"
	SkipUpstreamExists SkipReason = "upstream exists"
	SkipWrongRemote    SkipReason = "wrong remote"
	SkipNoPR           SkipReason = "no merged PR"
)

// IsInsideWorkTree checks if we're inside a git repository
func IsInsideWorkTree() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// FetchAndPrune fetches from remote and prunes deleted branches
func FetchAndPrune(remote string) error {
	cmd := exec.Command("git", "fetch", remote, "--prune", "--quiet")
	return cmd.Run()
}

// CurrentBranch returns the current branch name
func CurrentBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", nil // detached HEAD
	}
	return strings.TrimSpace(string(out)), nil
}

// ListBranches returns all local branches with their upstream info
func ListBranches(remote string) ([]Branch, error) {
	// Format: refname:short, upstream:short, upstream:track, committerdate:unix
	cmd := exec.Command("git", "for-each-ref",
		"--format=%(refname:short)\t%(upstream:short)\t%(upstream:track)\t%(committerdate:unix)",
		"refs/heads")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	current, _ := CurrentBranch()

	var branches []Branch
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		name := parts[0]
		upstream := ""
		track := ""
		var lastCommit time.Time

		if len(parts) > 1 {
			upstream = parts[1]
		}
		if len(parts) > 2 {
			track = parts[2]
		}
		if len(parts) > 3 {
			if ts, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
				lastCommit = time.Unix(ts, 0)
			}
		}

		b := Branch{
			Name:         name,
			Upstream:     upstream,
			IsCurrent:    name == current,
			UpstreamGone: strings.Contains(track, "gone"),
			LastCommit:   lastCommit,
		}

		branches = append(branches, b)
	}

	return branches, nil
}

// UpstreamExists checks if a remote ref still exists
func UpstreamExists(upstream string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/"+upstream)
	return cmd.Run() == nil
}

// DeleteBranch deletes a local branch forcefully
func DeleteBranch(name string) error {
	cmd := exec.Command("git", "branch", "-D", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete %s: %s", name, stderr.String())
	}
	return nil
}

// IsOnRemote checks if a branch's upstream is on the specified remote
func IsOnRemote(upstream, remote string) bool {
	return strings.HasPrefix(upstream, remote+"/")
}
