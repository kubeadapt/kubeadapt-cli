package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/version"
)

func TestShouldSkip(t *testing.T) {
	cases := []struct {
		name          string
		noUpdateCheck string
		ci            string
		stdoutIsTTY   bool
		want          bool
	}{
		{"interactive terminal, no opt-out", "", "", true, false},
		{"opt-out set to 1", "1", "", true, true},
		{"opt-out set to 0 still counts", "0", "", true, true},
		{"opt-out set to false still counts", "false", "", true, true},
		{"CI set", "", "true", true, true},
		{"CI set to 0 still counts", "", "0", true, true},
		{"stdout piped", "", "", false, true},
		{"piped inside CI", "", "true", false, true},
		{"empty env vars do not skip", "", "", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkip(tc.noUpdateCheck, tc.ci, tc.stdoutIsTTY)
			if got != tc.want {
				t.Errorf("shouldSkip(%q, %q, %v) = %v, want %v",
					tc.noUpdateCheck, tc.ci, tc.stdoutIsTTY, got, tc.want)
			}
		})
	}
}

// go test runs without a terminal, so ShouldSkip must always report true here.
// A regression that dropped the TTY guard would surface as a live network call
// on every developer's test run.
func TestShouldSkip_TrueUnderGoTest(t *testing.T) {
	if !ShouldSkip() {
		t.Error("ShouldSkip() = false with a non-terminal stdout, want true")
	}
}

// Primes the 24h cache with a newer release so checkForUpdate can be exercised
// without reaching GitHub.
func primeNewerRelease(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	prev := version.Version
	version.Version = "0.1.0"
	t.Cleanup(func() { version.Version = prev })

	data, err := json.Marshal(cachedCheck{CheckedAt: time.Now(), LatestVersion: "9.9.9"})
	if err != nil {
		t.Fatalf("marshaling cache: %v", err)
	}
	cacheDir := filepath.Join(dir, "kubeadapt")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("creating cache dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, cacheFile), data, 0600); err != nil {
		t.Fatalf("writing cache: %v", err)
	}
}

func TestCheckForUpdate_ReportsUpgradeWhenNotSkipped(t *testing.T) {
	primeNewerRelease(t)

	got := checkForUpdate(false)
	if got == "" {
		t.Fatal("checkForUpdate(false) = \"\", want an upgrade message; " +
			"the skip assertion below would otherwise prove nothing")
	}
}

func TestCheckForUpdate_SilentWhenSkipped(t *testing.T) {
	primeNewerRelease(t)

	if got := checkForUpdate(true); got != "" {
		t.Errorf("checkForUpdate(true) = %q, want \"\"", got)
	}
}
