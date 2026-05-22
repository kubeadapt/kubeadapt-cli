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
	assert.Empty(t, cfg.APIKey)
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		APIURL: "https://custom.api.com",
		APIKey: "test-api-key-12345",
	}

	require.NoError(t, Save(cfg, path))

	// Verify file permissions
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	loaded, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, cfg.APIURL, loaded.APIURL)
	assert.Equal(t, cfg.APIKey, loaded.APIKey)
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"short", "****"},
		{"12345678", "****"},
		{"ka_1234567890abcdef", "ka_1...cdef"},
	}

	for _, tt := range tests {
		got := MaskAPIKey(tt.input)
		assert.Equal(t, tt.want, got, "MaskAPIKey(%q)", tt.input)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	// Write invalid YAML
	require.NoError(t, os.WriteFile(path, []byte("{invalid: yaml: content: ["), 0600))
	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config")
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")
	cfg := &Config{APIURL: "https://test.api.com", APIKey: "test-key"}
	require.NoError(t, Save(cfg, nestedPath))
	// Verify file was created
	_, err := os.Stat(nestedPath)
	assert.NoError(t, err, "expected file to exist at %s", nestedPath)
}

func TestLoadEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(""), 0600))
	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, defaultAPIURL, cfg.APIURL)
}

func TestSaveAndLoadPreservesEmptyKey(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	cfg := &Config{APIURL: "https://test.api.com", APIKey: ""}
	require.NoError(t, Save(cfg, path))
	loaded, err := Load(path)
	require.NoError(t, err)
	assert.Empty(t, loaded.APIKey)
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Skip("os.UserHomeDir() unavailable in this environment")
	}
	assert.Contains(t, path, "kubeadapt")
	assert.Contains(t, path, "config.yaml")
}

func TestMaskAPIKey_Boundary(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"exactly 8 chars", "12345678", "****"},            // ≤ 8 → masked
		{"exactly 9 chars", "123456789", "1234...6789"},    // > 8 → first4...last4
		{"long key", "ka_1234567890abcdef", "ka_1...cdef"}, // existing pattern
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAPIKey(tt.input)
			assert.Equal(t, tt.want, got, "MaskAPIKey(%q)", tt.input)
		})
	}
}

func TestLoadReadFileError(t *testing.T) {
	_, err := Load("/nonexistent/path/to/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading config")
}

func TestSave_EmptyPath(t *testing.T) {
	defaultPath := DefaultPath()
	if defaultPath == "" {
		t.Skip("home directory not available")
	}
	// Read original content before any writes so we can restore it
	origData, readErr := os.ReadFile(defaultPath)
	// Register cleanup BEFORE calling Save so it runs even on panic
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

func TestDefaultPath_ReturnsNonEmpty(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Skip("os.UserHomeDir() unavailable in this environment")
	}
	assert.Contains(t, path, "kubeadapt")
	assert.Contains(t, path, "config.yaml")
}

func TestSaveWriteError(t *testing.T) {
	// Try to save to a read-only directory (simulate write error)
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0500))
	cfg := &Config{APIURL: "https://test.api.com", APIKey: "test-key"}
	path := filepath.Join(readOnlyDir, "config.yaml")
	err := Save(cfg, path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing config")
}
