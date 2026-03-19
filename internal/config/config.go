package config

import (
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

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}

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

	if err := os.WriteFile(path, data, 0600); err != nil {
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
