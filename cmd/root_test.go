package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
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
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	assert.Equal(t, "http://from-flag.com", rc.Config.APIURL, "flag wins")
	assert.Equal(t, "flag-key", rc.Config.APIKey, "flag wins")
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
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	assert.Equal(t, "http://from-env.com", rc.Config.APIURL, "env wins")
	assert.Equal(t, "env-key", rc.Config.APIKey, "env wins")
}

func TestConfigOverrideChain_ConfigFallback(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	cfgFile = cfgPath

	cmd := dummyCmd()
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	assert.Equal(t, "http://from-config.com", rc.Config.APIURL, "config wins")
	assert.Equal(t, "config-key", rc.Config.APIKey, "config wins")
}

func TestConfigMissing_UsesDefault(t *testing.T) {
	resetGlobals(t)

	cfgFile = filepath.Join(t.TempDir(), "nonexistent.yaml")

	cmd := dummyCmd()
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	const wantURL = "https://api.kubeadapt.io"
	assert.Equal(t, wantURL, rc.Config.APIURL)
}
