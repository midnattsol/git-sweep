package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration from environment variables
type Config struct {
	Remote           string
	ProtectedPattern []string
	Limit            int
	StaleDays        int
	NoColor          bool
}

// Default protected branch patterns
var defaultProtected = []string{
	"master",
	"main",
	"develop",
	"development",
	"staging",
	"release/**",
	"hotfix/**",
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Remote:           getEnv("GIT_SWEEP_REMOTE", "origin"),
		ProtectedPattern: getEnvList("GIT_SWEEP_PROTECTED", defaultProtected),
		Limit:            getEnvInt("GIT_SWEEP_LIMIT", 50),
		StaleDays:        getEnvInt("GIT_SWEEP_STALE_DAYS", 30),
		NoColor:          getEnvBool("GIT_SWEEP_NO_COLOR", false),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		v = strings.ToLower(v)
		return v == "true" || v == "1" || v == "yes"
	}
	return fallback
}

// IsProtected checks if a branch name matches any protected pattern
func (c *Config) IsProtected(branch string) bool {
	for _, pattern := range c.ProtectedPattern {
		if matchPattern(pattern, branch) {
			return true
		}
	}
	return false
}

// matchPattern matches a branch against a glob-like pattern
// Supports * (single level) and ** (any depth)
func matchPattern(pattern, branch string) bool {
	// Exact match
	if pattern == branch {
		return true
	}

	// Handle ** (any depth)
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(branch, prefix+"/")
	}

	// Handle * (single level)
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if !strings.HasPrefix(branch, prefix+"/") {
			return false
		}
		rest := strings.TrimPrefix(branch, prefix+"/")
		// Should not contain another /
		return !strings.Contains(rest, "/")
	}

	return false
}
