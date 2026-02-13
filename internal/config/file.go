package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the persistent configuration stored in a file
type FileConfig struct {
	Protected []string `yaml:"protected,omitempty"`
}

// configFileName is the name of the config file
const configFileName = "config.yaml"

// getConfigDir returns the config directory path
func getConfigDir() (string, error) {
	// Use XDG_CONFIG_HOME if set, otherwise ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "git-sweep"), nil
}

// getConfigPath returns the full path to the config file
func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// LoadFileConfig loads the config from the file
func LoadFileConfig() (*FileConfig, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileConfig{}, nil
		}
		return nil, err
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveFileConfig saves the config to the file
func SaveFileConfig(cfg *FileConfig) error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddProtected adds a pattern to the protected list
func (fc *FileConfig) AddProtected(pattern string) bool {
	// Check if already exists
	for _, p := range fc.Protected {
		if p == pattern {
			return false
		}
	}
	fc.Protected = append(fc.Protected, pattern)
	return true
}

// RemoveProtected removes a pattern from the protected list
func (fc *FileConfig) RemoveProtected(pattern string) bool {
	for i, p := range fc.Protected {
		if p == pattern {
			fc.Protected = append(fc.Protected[:i], fc.Protected[i+1:]...)
			return true
		}
	}
	return false
}

// GetProtectedPatterns returns the effective protected patterns
// combining defaults with user-configured ones
func GetProtectedPatterns() ([]string, error) {
	// Start with defaults
	patterns := make([]string, len(defaultProtected))
	copy(patterns, defaultProtected)

	// Check env var override first
	if v := os.Getenv("GIT_SWEEP_PROTECTED"); v != "" {
		// Env var completely overrides
		return getEnvList("GIT_SWEEP_PROTECTED", defaultProtected), nil
	}

	// Load file config and merge
	fc, err := LoadFileConfig()
	if err != nil {
		return patterns, nil // Return defaults on error
	}

	// Add user-configured patterns (avoiding duplicates)
	seen := make(map[string]bool)
	for _, p := range patterns {
		seen[p] = true
	}
	for _, p := range fc.Protected {
		if !seen[p] {
			patterns = append(patterns, p)
			seen[p] = true
		}
	}

	return patterns, nil
}
