package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestContext(t *testing.T, serverURL, key, format string) context.Context {
	t.Helper()
	rc := &RunContext{
		Config:    &config.Config{APIURL: serverURL, APIKey: key},
		Logger:    zap.NewNop(),
		OutputFmt: format,
		NoColor:   true,
	}
	ctx := context.WithValue(t.Context(), runContextKey{}, rc)

	// The new persistent flags (cost-mode/cursor/limit/paginate/include-total)
	// live on getCmd. When tests call getClustersCmd.RunE directly they bypass
	// the parent's flag inheritance, so parsePagedFlags fails with "flag not
	// defined". Bind the parent's persistent flag set here so direct RunE calls
	// see the defaults. Safe to call multiple times — cobra dedupes by name.
	if getClustersCmd.Flags().Lookup(flagCostMode) == nil {
		getClustersCmd.Flags().AddFlagSet(getCmd.PersistentFlags())
	}
	return ctx
}

func TestGetClusters_TableOutput(t *testing.T) {
	server := testutil.NewMockServer(t)

	ctx := setupTestContext(t, server.URL, "test-key", "table")
	getClustersCmd.SetContext(ctx)
	require.NoError(t, getClustersCmd.RunE(getClustersCmd, nil))
}

func TestGetClusters_JSONOutput(t *testing.T) {
	server := testutil.NewMockServer(t)

	ctx := setupTestContext(t, server.URL, "test-key", "json")
	getClustersCmd.SetContext(ctx)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() {
		w.Close()
		os.Stdout = oldStdout
	})

	err := getClustersCmd.RunE(getClustersCmd, nil)

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = oldStdout

	require.NoError(t, err)
	assert.True(t, json.Valid(buf.Bytes()), "output is not valid JSON: %s", buf.String())
}

func TestGetClusters_NoAPIKey(t *testing.T) {
	ctx := setupTestContext(t, "http://localhost:9999", "", "table")
	getClustersCmd.SetContext(ctx)

	err := getClustersCmd.RunE(getClustersCmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no API key")
}
