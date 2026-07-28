package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeadapt/kubeadapt-cli/internal/version"
)

// executeRoot drives the real command tree the way a shell would, so the
// assertions cover flag parsing, arg validation, and PersistentPreRunE rather
// than a hand-built RunE call that skips all three.
//
// cobra 1.9 copies the root context into a subcommand only while that
// subcommand's own context is nil, so a context set during one in-process
// Execute survives into the next one. Clearing it per call keeps cases
// independent.
func executeRoot(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs(args)
	for _, sub := range rootCmd.Commands() {
		sub.SetContext(nil)
	}

	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs([]string{})
		for _, sub := range rootCmd.Commands() {
			sub.SetContext(nil)
		}
	})

	err = rootCmd.Execute()
	return out.String(), errOut.String(), err
}

func TestVersion_WritesToCmdOut(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	out, _, err := executeRoot(t, "version")

	require.NoError(t, err)
	assert.NotEmpty(t, out,
		"version must write through cmd.OutOrStdout so output can be captured and redirected")
	assert.Contains(t, out, version.Version)
}

func TestVersion_JSONOutput(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	out, _, err := executeRoot(t, "version", "-o", "json")

	require.NoError(t, err)
	require.True(t, json.Valid([]byte(out)), "version -o json must emit valid JSON, got: %q", out)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, version.Version, payload["version"])
	assert.Contains(t, payload, "go_version")
}

func TestVersion_YAMLOutput(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	out, _, err := executeRoot(t, "version", "-o", "yaml")

	require.NoError(t, err)
	assert.Contains(t, out, "version: "+version.Version)
	assert.False(t, json.Valid([]byte(out)), "yaml output must not be the JSON document")
}

func TestVersion_TableOutputUnchanged(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	out, _, err := executeRoot(t, "version")

	require.NoError(t, err)
	for _, want := range []string{"kubeadapt " + version.Version, "Commit:", "Built:", "Go version:", "OS/Arch:"} {
		assert.Contains(t, out, want)
	}
}

func TestVersion_RejectsExtraArgs(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	_, _, err := executeRoot(t, "version", "foo")

	require.Error(t, err, "unexpected positional args must be rejected, not silently ignored")
}
