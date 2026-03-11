package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.APIURL != defaultAPIURL {
		t.Errorf("expected API URL %q, got %q", defaultAPIURL, cfg.APIURL)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty API key, got %q", cfg.APIKey)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		APIURL: "https://custom.api.com",
		APIKey: "test-api-key-12345",
	}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.APIURL != cfg.APIURL {
		t.Errorf("APIURL: expected %q, got %q", cfg.APIURL, loaded.APIURL)
	}
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("APIKey: expected %q, got %q", cfg.APIKey, loaded.APIKey)
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent config")
	}
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
		if got != tt.want {
			t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	// Write invalid YAML
	if err := os.WriteFile(path, []byte("{invalid: yaml: content: ["), 0600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing config") {
		t.Errorf("expected error to contain 'parsing config', got: %v", err)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")
	cfg := &Config{APIURL: "https://test.api.com", APIKey: "test-key"}
	if err := Save(cfg, nestedPath); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	// Verify file was created
	if _, err := os.Stat(nestedPath); err != nil {
		t.Errorf("expected file to exist at %s: %v", nestedPath, err)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error for empty file: %v", err)
	}
	if cfg.APIURL != defaultAPIURL {
		t.Errorf("expected default APIURL %q, got %q", defaultAPIURL, cfg.APIURL)
	}
}

func TestSaveAndLoadPreservesEmptyKey(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	cfg := &Config{APIURL: "https://test.api.com", APIKey: ""}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", loaded.APIKey)
	}
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Skip("os.UserHomeDir() unavailable in this environment")
	}
	if !strings.Contains(path, ".kubeadapt") {
		t.Errorf("expected path to contain '.kubeadapt', got %q", path)
	}
	if !strings.Contains(path, "config.yaml") {
		t.Errorf("expected path to contain 'config.yaml', got %q", path)
	}
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
			if got != tt.want {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadReadFileError(t *testing.T) {
	_, err := Load("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "reading config") {
		t.Errorf("expected error to contain 'reading config', got: %v", err)
	}
}

func TestSave_EmptyPath(t *testing.T) {
	defaultPath := DefaultPath()
	if defaultPath == "" {
		t.Skip("home directory not available")
	}
	// Save to default path, then restore original content
	origData, readErr := os.ReadFile(defaultPath)
	cfg := &Config{APIURL: "http://test-empty-path.com", APIKey: "test-key"}
	err := Save(cfg, "")
	if err != nil {
		t.Fatalf("Save(\"\") error: %v", err)
	}
	// Restore original content
	t.Cleanup(func() {
		if readErr == nil {
			_ = os.WriteFile(defaultPath, origData, 0600)
		} else {
			_ = os.Remove(defaultPath)
		}
	})
}

func TestDefaultPath_ReturnsNonEmpty(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Skip("os.UserHomeDir() unavailable in this environment")
	}
	if !strings.Contains(path, ".kubeadapt") {
		t.Errorf("expected path to contain '.kubeadapt', got %q", path)
	}
	if !strings.Contains(path, "config.yaml") {
		t.Errorf("expected path to contain 'config.yaml', got %q", path)
	}
}

func TestSaveWriteError(t *testing.T) {
	// Try to save to a read-only directory (simulate write error)
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0500); err != nil {
		t.Fatalf("Mkdir() error: %v", err)
	}
	cfg := &Config{APIURL: "https://test.api.com", APIKey: "test-key"}
	path := filepath.Join(readOnlyDir, "config.yaml")
	err := Save(cfg, path)
	if err == nil {
		t.Fatal("expected error for write to read-only directory")
	}
	if !strings.Contains(err.Error(), "writing config") {
		t.Errorf("expected error to contain 'writing config', got: %v", err)
	}
}
