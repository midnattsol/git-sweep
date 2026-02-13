// Package provider defines interfaces and implementations for git hosting providers
// (GitHub, GitLab, Bitbucket) to check merged PR/MR status.
package provider

import (
	"context"
)

// Provider is the interface that git hosting providers must implement
type Provider interface {
	// Name returns the provider name (e.g., "github", "gitlab", "bitbucket")
	Name() string

	// MergedPRBranches returns a map of branch names that have merged PRs/MRs
	// The limit parameter controls how many recent PRs to check
	MergedPRBranches(ctx context.Context, limit int) (map[string]bool, error)
}

// Config holds provider-specific configuration
type Config struct {
	// GitHub
	GitHubToken string

	// GitLab
	GitLabToken string
	GitLabURL   string // Default: https://gitlab.com

	// Bitbucket
	BitbucketUsername string
	BitbucketToken    string
}

// RepoInfo contains repository information extracted from remote URL
type RepoInfo struct {
	Provider  string // "github", "gitlab", "bitbucket"
	Owner     string // Owner/namespace/workspace
	Repo      string // Repository name
	BaseURL   string // For self-hosted instances
	RemoteURL string // Original remote URL
}
