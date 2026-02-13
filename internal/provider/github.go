package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v60/github"
)

// GitHub implements the Provider interface for GitHub
type GitHub struct {
	client *github.Client
	owner  string
	repo   string
}

// NewGitHub creates a new GitHub provider
func NewGitHub(info *RepoInfo, token string) (*GitHub, error) {
	if token == "" {
		var err error
		token, err = getGitHubToken()
		if err != nil {
			return nil, err
		}
	}

	client := github.NewClient(nil).WithAuthToken(token)

	return &GitHub{
		client: client,
		owner:  info.Owner,
		repo:   info.Repo,
	}, nil
}

// getGitHubToken retrieves a GitHub token from environment
func getGitHubToken() (string, error) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("GITHUB_TOKEN not set. Create one at https://github.com/settings/tokens")
}

// Name returns the provider name
func (g *GitHub) Name() string {
	return "github"
}

// MergedPRBranches returns a set of branch names that have merged PRs
func (g *GitHub) MergedPRBranches(ctx context.Context, limit int) (map[string]bool, error) {
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
		prs, resp, err := g.client.PullRequests.List(ctx, g.owner, g.repo, opts)
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
