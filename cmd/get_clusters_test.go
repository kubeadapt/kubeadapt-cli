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

func TestGetClusters_TableOutput(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	cfg = &config.Config{APIURL: server.URL, APIKey: "test-key"}
	log = zap.NewNop()
	outputFmt = "table"
	noColor = true
	t.Cleanup(func() {
		cfg = nil
		log = nil
		outputFmt = "table"
		noColor = false
	})

	ctx := context.Background()
	getClustersCmd.SetContext(ctx)
	err := getClustersCmd.RunE(getClustersCmd, nil)
	if err != nil {
		t.Fatalf("getClustersCmd.RunE() = %v, want nil", err)
	}
}
func TestGetClusters_JSONOutput(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	cfg = &config.Config{APIURL: server.URL, APIKey: "test-key"}
	log = zap.NewNop()
	outputFmt = "json"
	noColor = true
	t.Cleanup(func() {
		cfg = nil
		log = nil
		outputFmt = "table"
		noColor = false
	})

	// Capture stdout — restore in t.Cleanup so it runs even on panic
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() {
		w.Close()
		os.Stdout = oldStdout
	})

	ctx := context.Background()
	getClustersCmd.SetContext(ctx)
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
	cfg = &config.Config{APIURL: "http://localhost:9999", APIKey: ""}
	log = zap.NewNop()
	t.Cleanup(func() {
		cfg = nil
		log = nil
	})

	ctx := context.Background()
	getClustersCmd.SetContext(ctx)
	err := getClustersCmd.RunE(getClustersCmd, nil)
	if err == nil {
		t.Fatalf("getClustersCmd.RunE() = nil, want error")
	}

	if !strings.Contains(err.Error(), "no API key") {
		t.Errorf("error message = %q, want to contain 'no API key'", err.Error())
	}
}
