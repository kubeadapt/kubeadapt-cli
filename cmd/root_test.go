package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func resetGlobals(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		apiURL = ""
		apiKey = ""
		cfgFile = ""
		outputFmt = "table"
		noColor = false
		verbose = false
		quiet = false
	})
}

func dummyCmd() *cobra.Command {
	c := &cobra.Command{Use: "status"}
	c.SetContext(context.Background())
	return c
}

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeConfigFile: %v", err)
	}
}

func TestConfigOverrideChain_FlagOverridesEnv(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	t.Setenv("KUBEADAPT_API_URL", "http://from-env.com")
	t.Setenv("KUBEADAPT_API_KEY", "env-key")

	cfgFile = cfgPath
	apiURL = "http://from-flag.com"
	apiKey = "flag-key"

	cmd := dummyCmd()
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	rc := getRunContext(cmd)
	if rc == nil {
		t.Fatal("RunContext is nil after PersistentPreRunE")
	}
	if rc.Config.APIURL != "http://from-flag.com" {
		t.Errorf("expected APIURL %q (flag wins), got %q", "http://from-flag.com", rc.Config.APIURL)
	}
	if rc.Config.APIKey != "flag-key" {
		t.Errorf("expected APIKey %q (flag wins), got %q", "flag-key", rc.Config.APIKey)
	}
}

func TestConfigOverrideChain_EnvOverridesConfig(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	t.Setenv("KUBEADAPT_API_URL", "http://from-env.com")
	t.Setenv("KUBEADAPT_API_KEY", "env-key")

	cfgFile = cfgPath

	cmd := dummyCmd()
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	rc := getRunContext(cmd)
	if rc == nil {
		t.Fatal("RunContext is nil after PersistentPreRunE")
	}
	if rc.Config.APIURL != "http://from-env.com" {
		t.Errorf("expected APIURL %q (env wins), got %q", "http://from-env.com", rc.Config.APIURL)
	}
	if rc.Config.APIKey != "env-key" {
		t.Errorf("expected APIKey %q (env wins), got %q", "env-key", rc.Config.APIKey)
	}
}

func TestConfigOverrideChain_ConfigFallback(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	cfgFile = cfgPath

	cmd := dummyCmd()
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	rc := getRunContext(cmd)
	if rc == nil {
		t.Fatal("RunContext is nil after PersistentPreRunE")
	}
	if rc.Config.APIURL != "http://from-config.com" {
		t.Errorf("expected APIURL %q (config wins), got %q", "http://from-config.com", rc.Config.APIURL)
	}
	if rc.Config.APIKey != "config-key" {
		t.Errorf("expected APIKey %q (config wins), got %q", "config-key", rc.Config.APIKey)
	}
}

func TestConfigMissing_UsesDefault(t *testing.T) {
	resetGlobals(t)

	cfgFile = filepath.Join(t.TempDir(), "nonexistent.yaml")

	cmd := dummyCmd()
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE error: %v", err)
	}

	rc := getRunContext(cmd)
	if rc == nil {
		t.Fatal("RunContext is nil after PersistentPreRunE")
	}
	const wantURL = "https://public-api.kubeadapt.io"
	if rc.Config.APIURL != wantURL {
		t.Errorf("expected default APIURL %q, got %q", wantURL, rc.Config.APIURL)
	}
}
