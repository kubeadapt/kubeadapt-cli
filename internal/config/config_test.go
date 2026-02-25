package config

import (
	"os"
	"path/filepath"
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
