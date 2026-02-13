package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v60/github"
)

const (
	owner = "midnattsol"
	repo  = "git-sweep"
)

// Release represents a GitHub release
type Release struct {
	TagName string
	Body    string
	Assets  []Asset
}

// Asset represents a release asset
type Asset struct {
	Name        string
	DownloadURL string
}

// CurrentVersion returns the current binary version
var CurrentVersion = "dev"

// CheckForUpdate checks if a newer version is available
func CheckForUpdate(ctx context.Context) (*Release, bool, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.GetTagName(), "v")
	currentVersion := strings.TrimPrefix(CurrentVersion, "v")

	if latestVersion == currentVersion || currentVersion == "dev" {
		return nil, false, nil
	}

	// Convert to our Release struct
	r := &Release{
		TagName: release.GetTagName(),
		Body:    release.GetBody(),
		Assets:  make([]Asset, 0, len(release.Assets)),
	}

	for _, asset := range release.Assets {
		r.Assets = append(r.Assets, Asset{
			Name:        asset.GetName(),
			DownloadURL: asset.GetBrowserDownloadURL(),
		})
	}

	return r, true, nil
}

// GetAssetForPlatform returns the download URL for the current platform
func (r *Release) GetAssetForPlatform() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to our release naming
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	expectedName := fmt.Sprintf("git-sweep-%s-%s.tar.gz", osName, arch)

	for _, asset := range r.Assets {
		if asset.Name == expectedName {
			return asset.DownloadURL, nil
		}
	}

	return "", fmt.Errorf("no release found for %s/%s", osName, arch)
}

// DownloadAndInstall downloads and installs the new version
func DownloadAndInstall(ctx context.Context, downloadURL string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-sweep-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the archive
	archivePath := filepath.Join(tmpDir, "git-sweep.tar.gz")
	if err := downloadFile(ctx, downloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Extract the binary
	binaryPath := filepath.Join(tmpDir, "git-sweep")
	if err := extractBinary(archivePath, binaryPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Replace the current binary
	// First, rename old binary as backup
	backupPath := execPath + ".old"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old binary: %w", err)
	}

	// Move new binary into place
	if err := copyFile(binaryPath, execPath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(execPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the binary
		if header.Name == "git-sweep" && header.Typeflag == tar.TypeReg {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, tr)
			return err
		}
	}

	return fmt.Errorf("binary not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// Config file handling for auto-update preference

// ConfigPath returns the path to the config file
func ConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("HOME")
	}
	return filepath.Join(configDir, "git-sweep", "config.json")
}

// Config holds user preferences
type Config struct {
	AutoUpdate bool `json:"auto_update"`
}

// LoadConfig loads the config file
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SaveConfig saves the config file
func SaveConfig(cfg *Config) error {
	path := ConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// SetAutoUpdate sets the auto-update preference
func SetAutoUpdate(enabled bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		cfg = &Config{}
	}
	cfg.AutoUpdate = enabled
	return SaveConfig(cfg)
}

// IsAutoUpdateEnabled returns true if auto-update is enabled
func IsAutoUpdateEnabled() bool {
	// Check env var first
	if v := os.Getenv("GIT_SWEEP_AUTO_UPDATE"); v != "" {
		v = strings.ToLower(v)
		return v == "true" || v == "1" || v == "yes"
	}

	cfg, err := LoadConfig()
	if err != nil {
		return false
	}
	return cfg.AutoUpdate
}
