package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// isNewer reports whether latest is a strictly newer semver than current.
// Pre-release and build-metadata suffixes are ignored. Returns false when
// either version is empty or unparseable so the upgrade prompt fails closed
// and never nags the user about an unverifiable "upgrade".
func isNewer(latest, current string) bool {
	l, ok := parseSemver(latest)
	if !ok {
		return false
	}
	c, ok := parseSemver(current)
	if !ok {
		return false
	}
	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func parseSemver(s string) ([3]int, bool) {
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}

func CheckForUpdate() string {
	if version.Version == "dev" {
		return ""
	}

	cachePath := filepath.Join(cacheDir(), cacheFile)
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached cachedCheck
		if json.Unmarshal(data, &cached) == nil && time.Since(cached.CheckedAt) < cacheDuration {
			if isNewer(cached.LatestVersion, version.Version) {
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
	_ = os.MkdirAll(cacheDir(), 0700)
	cacheData, _ := json.Marshal(cachedCheck{CheckedAt: time.Now(), LatestVersion: latest})
	_ = os.WriteFile(cachePath, cacheData, 0600)

	if isNewer(latest, version.Version) {
		return fmt.Sprintf("A new version of kubeadapt is available: %s → %s\n  Update with: brew upgrade kubeadapt", version.Version, release.TagName)
	}
	return ""
}
