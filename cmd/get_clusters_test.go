package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
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
	ctx := context.WithValue(context.Background(), runContextKey{}, rc)
	return ctx
}

func TestGetClusters_TableOutput(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	ctx := setupTestContext(t, server.URL, "test-key", "table")
	getClustersCmd.SetContext(ctx)
	err := getClustersCmd.RunE(getClustersCmd, nil)
	if err != nil {
		t.Fatalf("getClustersCmd.RunE() = %v, want nil", err)
	}
}

func TestGetClusters_JSONOutput(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

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

	if err != nil {
		t.Fatalf("getClustersCmd.RunE() = %v, want nil", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("output is not valid JSON: %s", buf.String())
	}
}

func TestGetClusters_NoAPIKey(t *testing.T) {
	ctx := setupTestContext(t, "http://localhost:9999", "", "table")
	getClustersCmd.SetContext(ctx)

	err := getClustersCmd.RunE(getClustersCmd, nil)
	if err == nil {
		t.Fatalf("getClustersCmd.RunE() = nil, want error")
	}

	if !strings.Contains(err.Error(), "no API key") {
		t.Errorf("error message = %q, want to contain 'no API key'", err.Error())
	}
}
