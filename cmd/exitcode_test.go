package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// Cobra keeps parsed values on the shared command tree, so a --limit from one
// case would otherwise leak into the next. Slice flags are replaced rather than
// Set(DefValue), whose "[]" spelling would parse as a literal element.
func resetFlags(c *cobra.Command) {
	for _, sub := range c.Commands() {
		resetFlags(sub)
	}
	reset := func(f *pflag.Flag) {
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	}
	c.Flags().VisitAll(reset)
	c.PersistentFlags().VisitAll(reset)
}

// Goes through run(), the same entry point Execute() uses, so the wiring it
// performs is covered too; a regression that stopped classifying usage errors
// there would otherwise stay invisible. Returns the process exit code and
// whatever was reported to the user.
func runRoot(t *testing.T, args ...string) (int, string) {
	t.Helper()
	resetGlobals(t)
	resetFlags(rootCmd)
	t.Cleanup(func() { resetFlags(rootCmd) })

	t.Setenv("KUBEADAPT_API_KEY", "")
	t.Setenv("KUBEADAPT_API_URL", "")

	cfgFile = filepath.Join(t.TempDir(), "nonexistent.yaml")
	args = append(args, "--config", cfgFile)

	rootCmd.SetArgs(args)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	var reported bytes.Buffer
	return run(&reported), reported.String()
}

func TestExitCode_UsageErrorsExitTwo(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"invalid --output", []string{"get", "clusters", "-o", "xml"}},
		{"unknown command", []string{"frobnicate"}},
		{"unknown flag", []string{"get", "clusters", "--offset", "1"}},
		{"unknown subcommand", []string{"get", "frobnicate"}},
		{"invalid --limit", []string{"get", "clusters", "--limit", "999"}},
		{"unparseable --limit", []string{"get", "clusters", "--limit", "abc"}},
		{"invalid --cost-mode", []string{"get", "clusters", "--cost-mode", "bogus"}},
		{"invalid --max-wait", []string{"get", "clusters", "--max-wait", "-5s"}},
		{"invalid --top-clusters-limit", []string{"get", "dashboard", "--top-clusters-limit", "99"}},
		{"missing required flag", []string{"get", "namespace", "ns"}},
		{"wrong arg count", []string{"get", "cluster"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, reported := runRoot(t, tc.args...)
			assert.Equal(t, exitUsage, code,
				"a pre-flight rejection must exit 2; reported: %s", reported)
		})
	}
}

func TestExitCode_AuthFailureExitsOne(t *testing.T) {
	code, reported := runRoot(t, "get", "clusters")
	assert.Equal(t, exitError, code,
		"a missing API key is a runtime failure, not a usage error; reported: %s", reported)
}

// A 400 means the request was well-formed enough to send. Classifying it as a
// usage error would tell scripts the invocation was wrong when it was not.
func TestExitCode_APIRejectionExitsOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"VALIDATION_ERROR","message":"bad cursor"}}`))
	}))
	defer srv.Close()

	code, reported := runRoot(t, "get", "clusters", "--api-key", "k", "--api-url", srv.URL)
	assert.Equal(t, exitError, code,
		"an API-layer rejection must stay on 1; reported: %s", reported)
}

func TestExitCode_SuccessIsZero(t *testing.T) {
	assert.Equal(t, exitOK, exitCodeFor(nil))
}

func TestExitCode_ReservesThreeAndAbove(t *testing.T) {
	for _, code := range []int{exitOK, exitError, exitUsage} {
		assert.Less(t, code, 3,
			"exit codes 3-125 are reserved; callers must not depend on them")
	}
}
