package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/version"
)

const (
	releaseURL    = "https://api.github.com/repos/kubeadapt/kubeadapt-cli/releases/latest"
	cacheDuration = 24 * time.Hour
	cacheFile     = "update-check.json"
)

type cachedCheck struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

func cacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "kubeadapt")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "kubeadapt")
}

func CheckForUpdate() string {
	if version.Version == "dev" {
		return ""
	}

	cachePath := filepath.Join(cacheDir(), cacheFile)
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached cachedCheck
		if json.Unmarshal(data, &cached) == nil && time.Since(cached.CheckedAt) < cacheDuration {
			if cached.LatestVersion != "" && cached.LatestVersion != version.Version {
				return fmt.Sprintf("A new version of kubeadapt is available: %s → %s\n  Update with: brew upgrade kubeadapt", version.Version, cached.LatestVersion)
			}
			return ""
		}
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(releaseURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if json.NewDecoder(resp.Body).Decode(&release) != nil {
		return ""
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version.Version, "v")
	_ = os.MkdirAll(cacheDir(), 0700)
	cacheData, _ := json.Marshal(cachedCheck{CheckedAt: time.Now(), LatestVersion: latest})
	_ = os.WriteFile(cachePath, cacheData, 0600)

	if latest != current && release.TagName != "" {
		return fmt.Sprintf("A new version of kubeadapt is available: %s → %s\n  Update with: brew upgrade kubeadapt", version.Version, release.TagName)
	}
	return ""
}
