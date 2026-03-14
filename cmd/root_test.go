package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// resetGlobals resets all package-level globals to their zero/default values.
// Must be called in t.Cleanup() for every test that touches PersistentPreRunE.
func resetGlobals(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		cfg = nil
		log = nil
		apiURL = ""
		apiKey = ""
		cfgFile = ""
		outputFmt = "table"
		noColor = false
		verbose = false
	})
}

// dummyCmd returns a cobra.Command whose name is NOT skipped by PersistentPreRunE.
// PersistentPreRunE skips "login", "version", and "completion".
func dummyCmd() *cobra.Command {
	return &cobra.Command{Use: "status"}
}

// writeConfigFile writes a minimal YAML config file to path and returns path.
func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeConfigFile: %v", err)
	}
}

// TestConfigOverrideChain_FlagOverridesEnv verifies that a CLI flag value
// takes priority over both environment variables and the config file value.
func TestConfigOverrideChain_FlagOverridesEnv(t *testing.T) {
	resetGlobals(t)

	// Prepare config file: lowest priority.
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	// Environment variable: middle priority.
	t.Setenv("KUBEADAPT_API_URL", "http://from-env.com")
	t.Setenv("KUBEADAPT_API_KEY", "env-key")

	// Flag values: highest priority.
	cfgFile = cfgPath
	apiURL = "http://from-flag.com"
	apiKey = "flag-key"

	if err := rootCmd.PersistentPreRunE(dummyCmd(), nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	if cfg == nil {
		t.Fatal("cfg is nil after PersistentPreRunE")
	}
	if cfg.APIURL != "http://from-flag.com" {
		t.Errorf("expected APIURL %q (flag wins), got %q", "http://from-flag.com", cfg.APIURL)
	}
	if cfg.APIKey != "flag-key" {
		t.Errorf("expected APIKey %q (flag wins), got %q", "flag-key", cfg.APIKey)
	}
}

// TestConfigOverrideChain_EnvOverridesConfig verifies that an environment
// variable overrides the config file when no flag is set.
func TestConfigOverrideChain_EnvOverridesConfig(t *testing.T) {
	resetGlobals(t)

	// Prepare config file: lowest priority.
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	// Environment variable: overrides config.
	t.Setenv("KUBEADAPT_API_URL", "http://from-env.com")
	t.Setenv("KUBEADAPT_API_KEY", "env-key")

	// No flag set, so apiURL and apiKey remain "".
	cfgFile = cfgPath

	if err := rootCmd.PersistentPreRunE(dummyCmd(), nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	if cfg == nil {
		t.Fatal("cfg is nil after PersistentPreRunE")
	}
	if cfg.APIURL != "http://from-env.com" {
		t.Errorf("expected APIURL %q (env wins), got %q", "http://from-env.com", cfg.APIURL)
	}
	if cfg.APIKey != "env-key" {
		t.Errorf("expected APIKey %q (env wins), got %q", "env-key", cfg.APIKey)
	}
}

// TestConfigOverrideChain_ConfigFallback verifies that the config file value
// is used when neither a flag nor an environment variable is set.
func TestConfigOverrideChain_ConfigFallback(t *testing.T) {
	resetGlobals(t)

	// Prepare config file.
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	// No env vars, no flags.
	cfgFile = cfgPath

	if err := rootCmd.PersistentPreRunE(dummyCmd(), nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	if cfg == nil {
		t.Fatal("cfg is nil after PersistentPreRunE")
	}
	if cfg.APIURL != "http://from-config.com" {
		t.Errorf("expected APIURL %q (config wins), got %q", "http://from-config.com", cfg.APIURL)
	}
	if cfg.APIKey != "config-key" {
		t.Errorf("expected APIKey %q (config wins), got %q", "config-key", cfg.APIKey)
	}
}

// TestConfigMissing_UsesDefault verifies that when no config file exists and
// no env vars or flags are set, the built-in default URL is used.
func TestConfigMissing_UsesDefault(t *testing.T) {
	resetGlobals(t)

	// Point cfgFile at a path that does not exist. Load() will error and
	// PersistentPreRunE will fall back to config.Default().
	cfgFile = filepath.Join(t.TempDir(), "nonexistent.yaml")

	// No env vars, no flags.

	if err := rootCmd.PersistentPreRunE(dummyCmd(), nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	if cfg == nil {
		t.Fatal("cfg is nil after PersistentPreRunE")
	}
	const wantURL = "http://localhost:8002"
	if cfg.APIURL != wantURL {
		t.Errorf("expected default APIURL %q, got %q", wantURL, cfg.APIURL)
	}
}
