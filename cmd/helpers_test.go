package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGetCmd() *cobra.Command {
	root := &cobra.Command{Use: "kubeadapt"}
	get := &cobra.Command{
		Use: "get",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if _, err := parsePagedFlags(cmd); err != nil {
				return err
			}
			return nil
		},
	}
	get.PersistentFlags().String(flagCostMode, "fully_loaded", "")
	get.PersistentFlags().String(flagCursor, "", "")
	get.PersistentFlags().Int(flagLimit, 100, "")
	get.PersistentFlags().Bool(flagPaginate, false, "")
	get.PersistentFlags().Bool(flagIncludeTotal, false, "")

	dummy := &cobra.Command{
		Use:  "dummy",
		RunE: func(_ *cobra.Command, _ []string) error { return nil },
	}
	get.AddCommand(dummy)
	root.AddCommand(get)
	return root
}

func executeGet(t *testing.T, args ...string) (*cobra.Command, error) {
	t.Helper()
	root := newTestGetCmd()
	root.SetArgs(append([]string{"get"}, args...))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	get, _, _ := root.Find([]string{"get"})
	return get, err
}

func TestParsePagedFlags_Defaults(t *testing.T) {
	get, err := executeGet(t, "dummy")
	require.NoError(t, err)
	got, err := parsePagedFlags(get)
	require.NoError(t, err)
	assert.Equal(t, "fully_loaded", got.CostMode)
	assert.Empty(t, got.Cursor)
	assert.Equal(t, 100, got.Limit)
	assert.False(t, got.Paginate)
	assert.False(t, got.IncludeTotal)
}

func TestParsePagedFlags_RejectsInvalidCostMode(t *testing.T) {
	_, err := executeGet(t, "--cost-mode=bogus", "dummy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --cost-mode")
}

func TestParsePagedFlags_RejectsLimitZero(t *testing.T) {
	_, err := executeGet(t, "--limit=0", "dummy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --limit")
}

func TestParsePagedFlags_RejectsLimit501(t *testing.T) {
	_, err := executeGet(t, "--limit=501", "dummy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --limit")
}

func TestParsePagedFlags_AcceptsValid(t *testing.T) {
	get, err := executeGet(t,
		"--cost-mode=workload_only",
		"--cursor=abc123",
		"--limit=250",
		"--paginate=true",
		"--include-total=true",
		"dummy",
	)
	require.NoError(t, err)
	got, err := parsePagedFlags(get)
	require.NoError(t, err)
	want := PagedFlags{
		CostMode:     "workload_only",
		Cursor:       "abc123",
		Limit:        250,
		Paginate:     true,
		IncludeTotal: true,
		MaxWait:      defaultMaxWait,
	}
	assert.Equal(t, want, got)
}

func TestParsePagedFlags_MaxWaitDefaultsWhenFlagAbsent(t *testing.T) {
	get, err := executeGet(t, "dummy")
	require.NoError(t, err)
	got, err := parsePagedFlags(get)
	require.NoError(t, err)
	assert.Equal(t, defaultMaxWait, got.MaxWait)
}

func TestParsePagedFlags_MaxWait(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    time.Duration
		wantErr bool
	}{
		{name: "explicit duration", arg: "--max-wait=90s", want: 90 * time.Second},
		{name: "zero means unbounded", arg: "--max-wait=0", want: 0},
		{name: "negative rejected", arg: "--max-wait=-1s", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := newTestGetCmd()
			get, _, err := root.Find([]string{"get"})
			require.NoError(t, err)
			get.PersistentFlags().Duration(flagMaxWait, defaultMaxWait, "")

			root.SetArgs([]string{"get", tc.arg, "dummy"})
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})

			if tc.wantErr {
				execErr := root.Execute()
				require.Error(t, execErr, "a negative budget must be rejected before any request")
				assert.Contains(t, execErr.Error(), "--max-wait")
				return
			}
			require.NoError(t, root.Execute())

			got, err := parsePagedFlags(get)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.MaxWait)
		})
	}
}

func TestGetCmdRejectsInvalidCostModeAtPreRun(t *testing.T) {
	_, err := executeGet(t, "--cost-mode=nope", "dummy")
	require.Error(t, err, "expected PersistentPreRunE to reject invalid cost-mode")
}
