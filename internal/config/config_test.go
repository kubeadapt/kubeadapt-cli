package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	assert.Equal(t, defaultAPIURL, cfg.APIURL)
	assert.Equal(t, "https://public-api.kubeadapt.io", cfg.APIURL, "default must point at the public API host")
	assert.Empty(t, cfg.APIKey)
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Skip("os.UserHomeDir() unavailable in this environment")
	}
	assert.Contains(t, path, "kubeadapt")
	assert.Contains(t, path, "config.yaml")
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "****"},
		{"short", "short", "****"},
		{"exactly 8 chars", "12345678", "****"},            // <= 8 -> masked
		{"exactly 9 chars", "123456789", "1234...6789"},    // > 8 -> first4...last4
		{"long key", "ka_1234567890abcdef", "ka_1...cdef"}, // existing pattern
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskAPIKey(tt.input), "MaskAPIKey(%q)", tt.input)
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		// path, when set, is used verbatim and no file is written.
		path       string
		content    string
		wantErr    string
		wantAPIURL string
	}{
		{name: "missing file reports read error", path: "/nonexistent/path/to/config.yaml", wantErr: "reading config"},
		{name: "malformed yaml reports parse error", content: "{invalid: yaml: content: [", wantErr: "parsing config"},
		{name: "empty file falls back to default", content: "", wantAPIURL: defaultAPIURL},
		{
			name:       "absent api_url falls back to public host",
			content:    "version: 1\napi_key: ka_test\n",
			wantAPIURL: "https://public-api.kubeadapt.io",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path
			if path == "" {
				path = filepath.Join(t.TempDir(), "config.yaml")
				require.NoError(t, os.WriteFile(path, []byte(tt.content), 0600))
			}

			cfg, err := Load(path)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAPIURL, cfg.APIURL)
		})
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
	}{
		{"populated key", &Config{APIURL: "https://custom.api.com", APIKey: "test-api-key-12345"}},
		{"empty key", &Config{APIURL: "https://test.api.com", APIKey: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, Save(tt.cfg, path))

			loaded, err := Load(path)
			require.NoError(t, err)
			assert.Equal(t, tt.cfg.APIURL, loaded.APIURL)
			assert.Equal(t, tt.cfg.APIKey, loaded.APIKey)
		})
	}
}

// Save writes via temp+fsync+rename. The rename must leave the directory
// holding exactly the final file at 0600 - a leaked temp file means a partial
// write could be picked up, and a wider mode would expose the API key.
func TestSave_AtomicWriteLeavesOnly0600Config(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	require.NoError(t, Save(&Config{APIURL: "https://test.api.com", APIKey: "k"}, path))

	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	assert.Equal(t, []string{"config.yaml"}, names, "atomic write must not leave a temp file behind")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestSaveCreatesDirectory(t *testing.T) {
	nestedPath := filepath.Join(t.TempDir(), "nested", "dir", "config.yaml")
	cfg := &Config{APIURL: "https://test.api.com", APIKey: "test-key"}
	require.NoError(t, Save(cfg, nestedPath))

	_, err := os.Stat(nestedPath)
	assert.NoError(t, err, "expected file to exist at %s", nestedPath)
}

// A failed write must surface an error and must not destroy whatever config
// was already on disk.
func TestSave_WriteFailure(t *testing.T) {
	tests := []struct {
		name     string
		original []byte // when non-nil, seeded before the read-only chmod
	}{
		{name: "reports writing error", original: nil},
		{
			name:     "preserves original on failure",
			original: []byte("version: 1\napi_url: https://original.example\napi_key: ka_original\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os.Geteuid() == 0 {
				t.Skip("running as root bypasses directory permissions")
			}

			dir := filepath.Join(t.TempDir(), "cfg")
			require.NoError(t, os.Mkdir(dir, 0700))
			path := filepath.Join(dir, "config.yaml")

			if tt.original != nil {
				require.NoError(t, os.WriteFile(path, tt.original, 0600))
			}
			require.NoError(t, os.Chmod(dir, 0500))
			t.Cleanup(func() { _ = os.Chmod(dir, 0700) })

			err := Save(&Config{APIURL: "https://replacement.example", APIKey: "ka_replacement"}, path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "writing config")

			if tt.original != nil {
				after, readErr := os.ReadFile(path)
				require.NoError(t, readErr)
				assert.Equal(t, tt.original, after, "a failed save must leave the original byte-identical")
			}
		})
	}
}

// Kept standalone: this is the only case that touches the real home directory,
// so it needs backup/restore cleanup that would pollute a shared table.
func TestSave_EmptyPath(t *testing.T) {
	defaultPath := DefaultPath()
	if defaultPath == "" {
		t.Skip("home directory not available")
	}
	// Read original content before any writes so we can restore it.
	origData, readErr := os.ReadFile(defaultPath)
	// Register cleanup BEFORE calling Save so it runs even on panic.
	t.Cleanup(func() {
		if readErr == nil {
			_ = os.WriteFile(defaultPath, origData, 0600)
		} else {
			_ = os.Remove(defaultPath)
		}
	})
	cfg := &Config{APIURL: "http://test-empty-path.com", APIKey: "test-key"}
	require.NoError(t, Save(cfg, ""))
}
