package provider

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Common patterns for detecting providers from remote URLs
var (
	// GitHub patterns
	githubSSHPattern   = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
	githubHTTPSPattern = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)

	// GitLab patterns (gitlab.com and self-hosted)
	gitlabSSHPattern   = regexp.MustCompile(`^git@([^:]+):([^/]+(?:/[^/]+)*)/([^/]+?)(?:\.git)?$`)
	gitlabHTTPSPattern = regexp.MustCompile(`^https://([^/]+)/([^/]+(?:/[^/]+)*)/([^/]+?)(?:\.git)?$`)

	// Bitbucket patterns
	bitbucketSSHPattern   = regexp.MustCompile(`^git@bitbucket\.org:([^/]+)/([^/]+?)(?:\.git)?$`)
	bitbucketHTTPSPattern = regexp.MustCompile(`^https://[^@]*@?bitbucket\.org/([^/]+)/([^/]+?)(?:\.git)?$`)
)

// DetectFromRemote detects the provider and extracts repo info from a git remote
func DetectFromRemote(remoteName string) (*RepoInfo, error) {
	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL for '%s': %w", remoteName, err)
	}

	url := strings.TrimSpace(string(out))
	return DetectFromURL(url)
}

// DetectFromURL detects the provider and extracts repo info from a URL
func DetectFromURL(url string) (*RepoInfo, error) {
	info := &RepoInfo{RemoteURL: url}

	// Try GitHub
	if matches := githubSSHPattern.FindStringSubmatch(url); matches != nil {
		info.Provider = "github"
		info.Owner = matches[1]
		info.Repo = matches[2]
		info.BaseURL = "https://github.com"
		return info, nil
	}
	if matches := githubHTTPSPattern.FindStringSubmatch(url); matches != nil {
		info.Provider = "github"
		info.Owner = matches[1]
		info.Repo = matches[2]
		info.BaseURL = "https://github.com"
		return info, nil
	}

	// Try Bitbucket (before GitLab since patterns overlap)
	if matches := bitbucketSSHPattern.FindStringSubmatch(url); matches != nil {
		info.Provider = "bitbucket"
		info.Owner = matches[1]
		info.Repo = matches[2]
		info.BaseURL = "https://api.bitbucket.org/2.0"
		return info, nil
	}
	if matches := bitbucketHTTPSPattern.FindStringSubmatch(url); matches != nil {
		info.Provider = "bitbucket"
		info.Owner = matches[1]
		info.Repo = matches[2]
		info.BaseURL = "https://api.bitbucket.org/2.0"
		return info, nil
	}

	// Try GitLab (including self-hosted)
	// SSH format: git@gitlab.example.com:group/subgroup/repo.git
	if matches := gitlabSSHPattern.FindStringSubmatch(url); matches != nil {
		host := matches[1]
		// Skip if it's GitHub or Bitbucket
		if host == "github.com" || host == "bitbucket.org" {
			return nil, fmt.Errorf("cannot parse remote URL: %s", url)
		}
		info.Provider = "gitlab"
		// For GitLab, the path can include subgroups
		fullPath := matches[2] + "/" + matches[3]
		parts := strings.Split(fullPath, "/")
		info.Repo = parts[len(parts)-1]
		info.Owner = strings.Join(parts[:len(parts)-1], "/")
		info.BaseURL = fmt.Sprintf("https://%s", host)
		return info, nil
	}

	// HTTPS format for GitLab
	if matches := gitlabHTTPSPattern.FindStringSubmatch(url); matches != nil {
		host := matches[1]
		// Skip if it's GitHub or Bitbucket
		if host == "github.com" || strings.Contains(host, "bitbucket") {
			return nil, fmt.Errorf("cannot parse remote URL: %s", url)
		}
		info.Provider = "gitlab"
		fullPath := matches[2] + "/" + matches[3]
		parts := strings.Split(fullPath, "/")
		info.Repo = parts[len(parts)-1]
		info.Owner = strings.Join(parts[:len(parts)-1], "/")
		info.BaseURL = fmt.Sprintf("https://%s", host)
		return info, nil
	}

	return nil, fmt.Errorf("cannot detect provider from URL: %s", url)
}

// New creates a new Provider based on the detected repo info and config
func New(info *RepoInfo, cfg *Config) (Provider, error) {
	switch info.Provider {
	case "github":
		return NewGitHub(info, cfg.GitHubToken)
	case "gitlab":
		baseURL := cfg.GitLabURL
		if baseURL == "" {
			baseURL = info.BaseURL
		}
		return NewGitLab(info, cfg.GitLabToken, baseURL)
	case "bitbucket":
		return NewBitbucket(info, cfg.BitbucketUsername, cfg.BitbucketAppPassword)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", info.Provider)
	}
}
