package config

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	defaultAPIURL = "https://public-api.kubeadapt.io"
	configFile    = "config.yaml"
)

type Config struct {
	Version int    `yaml:"version"`
	APIURL  string `yaml:"api_url"`
	APIKey  string `yaml:"api_key"`
}

func Default() *Config {
	return &Config{
		Version: 1,
		APIURL:  defaultAPIURL,
		APIKey:  "",
	}
}

func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "kubeadapt", configFile)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// The legacy location is probed before os.UserConfigDir so that an existing
	// user's config is not silently orphaned when they upgrade.
	legacyPath := filepath.Join(home, ".kubeadapt", configFile)
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(home, ".config", "kubeadapt", configFile)
	}
	return filepath.Join(configDir, "kubeadapt", configFile)
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.APIURL = cmp.Or(cfg.APIURL, defaultAPIURL)
	cfg.Version = cmp.Or(cfg.Version, 1)

	return cfg, nil
}

func Save(cfg *Config, path string) error {
	if path == "" {
		path = DefaultPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Written to a sibling temp file and renamed into place so that a crash or
	// full disk cannot truncate the existing config, which holds the user's
	// only API key.
	tmp, err := os.CreateTemp(dir, configFile+".*.tmp")
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	if err := tmp.Chmod(0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
