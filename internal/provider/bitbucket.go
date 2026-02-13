package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Bitbucket implements the Provider interface for Bitbucket Cloud
type Bitbucket struct {
	client    *http.Client
	workspace string
	repoSlug  string
	username  string
	token     string
}

// NewBitbucket creates a new Bitbucket provider
func NewBitbucket(info *RepoInfo, username, token string) (*Bitbucket, error) {
	if username == "" {
		username = os.Getenv("BITBUCKET_USERNAME")
	}
	if token == "" {
		token = os.Getenv("BITBUCKET_TOKEN")
	}

	if username == "" || token == "" {
		return nil, fmt.Errorf("BITBUCKET_USERNAME and BITBUCKET_TOKEN not set. Create an API token at Bitbucket > Personal Settings > API tokens")
	}

	return &Bitbucket{
		client:    &http.Client{},
		workspace: info.Owner,
		repoSlug:  info.Repo,
		username:  username,
		token:     token,
	}, nil
}

// Name returns the provider name
func (b *Bitbucket) Name() string {
	return "bitbucket"
}

// bitbucketPRResponse represents the Bitbucket API response for pull requests
type bitbucketPRResponse struct {
	Values []struct {
		Source struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`
		State string `json:"state"`
	} `json:"values"`
	Next string `json:"next"`
}

// MergedPRBranches returns a set of branch names that have merged PRs
func (b *Bitbucket) MergedPRBranches(ctx context.Context, limit int) (map[string]bool, error) {
	result := make(map[string]bool)

	// Bitbucket API endpoint for merged pull requests
	baseURL := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests?state=MERGED&pagelen=%d",
		b.workspace, b.repoSlug, min(limit, 50),
	)

	url := baseURL
	fetched := 0

	for fetched < limit && url != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.SetBasicAuth(b.username, b.token)

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PRs: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("Bitbucket API returned status %d", resp.StatusCode)
		}

		var prResp bitbucketPRResponse
		if err := json.NewDecoder(resp.Body).Decode(&prResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, pr := range prResp.Values {
			if pr.Source.Branch.Name != "" {
				result[pr.Source.Branch.Name] = true
			}
			fetched++
			if fetched >= limit {
				break
			}
		}

		url = prResp.Next
	}

	return result, nil
}
