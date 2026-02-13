package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v60/github"
)

// GetToken returns a GitHub token from environment or gh CLI
func GetToken() (string, error) {
	// First try GITHUB_TOKEN env var
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Fallback to gh auth token
	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err == nil {
		token := strings.TrimSpace(string(out))
		if token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no GitHub token found. Set GITHUB_TOKEN or run: gh auth login")
}

// GetRepoInfo extracts owner and repo from git remote
func GetRepoInfo(remote string) (owner, repo string, err error) {
	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", remote)
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(out))
	return parseGitHubURL(url)
}

// parseGitHubURL extracts owner/repo from various GitHub URL formats
func parseGitHubURL(url string) (owner, repo string, err error) {
	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		path := url[idx+len("github.com/"):]
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("cannot parse GitHub URL: %s", url)
}

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
	owner  string
	repo   string
}

// NewClient creates a new GitHub client
func NewClient(remote string) (*Client, error) {
	token, err := GetToken()
	if err != nil {
		return nil, err
	}

	owner, repo, err := GetRepoInfo(remote)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(token)

	return &Client{
		client: client,
		owner:  owner,
		repo:   repo,
	}, nil
}

// MergedPRBranches returns a set of branch names that have merged PRs
func (c *Client) MergedPRBranches(ctx context.Context, limit int) (map[string]bool, error) {
	result := make(map[string]bool)

	opts := &github.PullRequestListOptions{
		State:     "closed",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: min(limit, 100),
		},
	}

	// Fetch PRs (may need multiple pages if limit > 100)
	fetched := 0
	for fetched < limit {
		prs, resp, err := c.client.PullRequests.List(ctx, c.owner, c.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PRs: %w", err)
		}

		for _, pr := range prs {
			if pr.GetMerged() && pr.Head != nil && pr.Head.Ref != nil {
				result[*pr.Head.Ref] = true
			}
			fetched++
			if fetched >= limit {
				break
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return result, nil
}

// HasMergedPR checks if a specific branch has a merged PR
func HasMergedPR(branch string, mergedBranches map[string]bool) bool {
	return mergedBranches[branch]
}
