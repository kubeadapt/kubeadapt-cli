package cmd

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
)

// health builds its own http.Client on the shared default transport, which
// parks an idle connection that outlives the test and trips goleak.
func closeIdleConns(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		if tr, ok := http.DefaultTransport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}
	})
}

func healthArgs(t *testing.T, serverURL string, extra ...string) []string {
	t.Helper()
	args := []string{"health", "--api-url", serverURL, "--config", filepath.Join(t.TempDir(), "absent.yaml")}
	return append(args, extra...)
}

func TestHealth_JSONOutput(t *testing.T) {
	resetGlobals(t)
	closeIdleConns(t)
	server := testutil.NewMockServer(t)

	out, _, err := executeRoot(t, healthArgs(t, server.URL, "-o", "json")...)

	require.NoError(t, err)
	require.True(t, json.Valid([]byte(out)), "health -o json must emit valid JSON, got: %q", out)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, server.URL+"/health", payload["url"])
}

func TestHealth_YAMLOutput(t *testing.T) {
	resetGlobals(t)
	closeIdleConns(t)
	server := testutil.NewMockServer(t)

	out, _, err := executeRoot(t, healthArgs(t, server.URL, "-o", "yaml")...)

	require.NoError(t, err)
	assert.Contains(t, out, "status: ok")
	assert.False(t, json.Valid([]byte(out)), "yaml output must not be the JSON document")
}

func TestHealth_TableOutputUnchanged(t *testing.T) {
	resetGlobals(t)
	closeIdleConns(t)
	server := testutil.NewMockServer(t)

	out, _, err := executeRoot(t, healthArgs(t, server.URL)...)

	require.NoError(t, err)
	assert.Contains(t, out, "Status:  ok")
	assert.Contains(t, out, "URL:     "+server.URL+"/health")
}

func TestHealth_RejectsExtraArgs(t *testing.T) {
	resetGlobals(t)
	closeIdleConns(t)
	server := testutil.NewMockServer(t)

	_, _, err := executeRoot(t, healthArgs(t, server.URL, "foo")...)

	require.Error(t, err, "unexpected positional args must be rejected, not silently ignored")
}

func TestHealth_HasUtilityGroupID(t *testing.T) {
	assert.Equal(t, groupUtility, newHealthCmd().GroupID,
		"health must be listed under Utility, not cobra's Additional Commands catch-all")
}

func TestHealth_QuietKeepsStatusOutput(t *testing.T) {
	resetGlobals(t)
	closeIdleConns(t)
	server := testutil.NewMockServer(t)

	out, _, err := executeRoot(t, healthArgs(t, server.URL, "--quiet")...)

	require.NoError(t, err)
	assert.Contains(t, out, "ok", "--quiet must not suppress the status the command exists to report")
	assert.NotContains(t, out, "URL:", "--quiet must drop the informational URL line")
}
