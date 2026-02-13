package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/xanzy/go-gitlab"
)

// GitLab implements the Provider interface for GitLab (including self-hosted)
type GitLab struct {
	client    *gitlab.Client
	projectID string // namespace/project format
}

// NewGitLab creates a new GitLab provider
func NewGitLab(info *RepoInfo, token string, baseURL string) (*GitLab, error) {
	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("no GitLab token found. Set GITLAB_TOKEN environment variable")
		}
	}

	if baseURL == "" {
		baseURL = os.Getenv("GITLAB_URL")
		if baseURL == "" {
			baseURL = "https://gitlab.com"
		}
	}

	// Ensure baseURL has proper format for API
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GitLab URL: %w", err)
	}
	apiURL := fmt.Sprintf("%s://%s/api/v4", parsedURL.Scheme, parsedURL.Host)

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(apiURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Project ID in GitLab is namespace/project (URL-encoded for API calls)
	projectID := fmt.Sprintf("%s/%s", info.Owner, info.Repo)

	return &GitLab{
		client:    client,
		projectID: projectID,
	}, nil
}

// Name returns the provider name
func (g *GitLab) Name() string {
	return "gitlab"
}

// MergedPRBranches returns a set of branch names that have merged MRs
func (g *GitLab) MergedPRBranches(ctx context.Context, limit int) (map[string]bool, error) {
	result := make(map[string]bool)

	state := "merged"
	orderBy := "updated_at"
	sort := "desc"

	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:   &state,
		OrderBy: &orderBy,
		Sort:    &sort,
		ListOptions: gitlab.ListOptions{
			PerPage: min(limit, 100),
		},
	}

	// Fetch MRs (may need multiple pages if limit > 100)
	fetched := 0
	for fetched < limit {
		mrs, resp, err := g.client.MergeRequests.ListProjectMergeRequests(g.projectID, opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch MRs: %w", err)
		}

		for _, mr := range mrs {
			if mr.SourceBranch != "" {
				result[mr.SourceBranch] = true
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
