package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	// cursorPastEnd drives the mock server's paginate() helper past the end of
	// every fixture list, which is the only way to get a genuinely zero-row
	// page out of it without editing internal/testutil.
	cursorPastEnd = "page3"

	// Fixture IDs mirroring internal/testutil; the mock server only special-cases
	// the literal "missing", so any well-formed value reaches the happy path.
	workloadUIDForPods    = "11111111-1111-1111-1111-000000000001"
	teamIDForAssignments  = "55555555-5555-5555-5555-000000000001"
	clusterIDForNamespace = "00000000-0000-0000-0000-000000000001"
	clusterIDSecond       = "00000000-0000-0000-0000-000000000002"
)

// listHarness runs one real `get` subcommand's RunE against a mock server.
//
// The subcommands are package-level singletons whose flag sets are shared with
// getCmd.PersistentFlags() by pointer, and cobra 1.9.1 only propagates a parent
// context into a subcommand whose ctx is still nil. Both leak across in-process
// runs, so every field this harness touches is restored on cleanup.
type listHarness struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func newListHarness(t *testing.T, c *cobra.Command, serverURL, outFmt string, flags map[string]string) *listHarness {
	t.Helper()

	if c.Flags().Lookup(flagCostMode) == nil {
		c.Flags().AddFlagSet(getCmd.PersistentFlags())
	}

	h := &listHarness{}
	c.SetOut(&h.stdout)
	c.SetErr(&h.stderr)

	rc := &RunContext{
		Config:    &config.Config{APIURL: serverURL, APIKey: "test-key"},
		Logger:    zap.NewNop(),
		OutputFmt: outFmt,
		NoColor:   true,
	}
	c.SetContext(context.WithValue(t.Context(), runContextKey{}, rc))

	for name, val := range flags {
		require.NoError(t, c.Flags().Set(name, val), "set --%s", name)
	}

	t.Cleanup(func() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed {
				return
			}
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
		c.SetContext(nil)
		c.SetOut(nil)
		c.SetErr(nil)
	})
	return h
}

// dataField returns the raw JSON of the envelope's `data` key so a test can
// tell `[]` from `null` - json.Unmarshal into []T erases that difference.
func dataField(t *testing.T, raw []byte) string {
	t.Helper()
	var envelope map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &envelope), "stdout was not JSON: %s", raw)
	got, ok := envelope["data"]
	require.True(t, ok, "envelope has no data key: %s", raw)
	return string(got)
}

func TestGetList_EmptyResultEmitsEmptyArrayNotNull(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "clusters", cmd: getClustersCmd},
		{name: "namespaces", cmd: getNamespacesCmd},
		{name: "nodes", cmd: getNodesCmd},
		{name: "recommendations", cmd: getRecommendationsCmd},
		{name: "teams", cmd: getTeamsCmd},
		{name: "departments", cmd: getDepartmentsCmd},
		{name: "workloads", cmd: getWorkloadsCmd},
		{name: "pods", cmd: getPodsCmd, args: []string{workloadUIDForPods}},
		{name: "team-assignments", cmd: getTeamAssignmentsCmd, args: []string{teamIDForAssignments}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := testutil.NewMockServer(t)
			h := newListHarness(t, tc.cmd, server.URL, formatJSON, map[string]string{
				flagCursor: cursorPastEnd,
			})

			require.NoError(t, tc.cmd.RunE(tc.cmd, tc.args))
			assert.Equal(t, "[]", dataField(t, h.stdout.Bytes()),
				"a zero-row result must stay iterable for jq '.data[]'")
		})
	}
}

func TestGetList_EmptyResultEmitsEmptyArrayInYAML(t *testing.T) {
	server := testutil.NewMockServer(t)
	h := newListHarness(t, getClustersCmd, server.URL, formatYAML, map[string]string{
		flagCursor: cursorPastEnd,
	})

	require.NoError(t, getClustersCmd.RunE(getClustersCmd, nil))
	assert.Contains(t, h.stdout.String(), "data: []")
}

func TestTeamAssignments_RejectsCostMode(t *testing.T) {
	server := testutil.NewMockServer(t)
	newListHarness(t, getTeamAssignmentsCmd, server.URL, formatJSON, map[string]string{
		flagCostMode: "workload_only",
	})

	err := getTeamAssignmentsCmd.RunE(getTeamAssignmentsCmd, []string{teamIDForAssignments})
	require.Error(t, err, "the assignments route ignores cost_mode, so accepting the flag silently lies")
	assert.Contains(t, err.Error(), "--cost-mode is not accepted")
}

// Asserts on the exact string, not a substring: ten inline checks collapsing
// onto one helper must not reword an error users already scripted against.
func TestCostModeRejection_StillLocalAndPreservesMessage(t *testing.T) {
	tests := []struct {
		cmd      *cobra.Command
		args     []string
		endpoint string
	}{
		{getClustersCmd, nil, "clusters"},
		{getClusterCmd, []string{clusterIDForNamespace}, "cluster"},
		{getNodesCmd, nil, "nodes"},
		{getNodeCmd, []string{"33333333-3333-3333-3333-000000000001"}, "node"},
		{getNodeGroupsCmd, nil, "node-groups"},
		{getNodeGroupCmd, []string{"general-purpose"}, "node-group"},
		{getRecommendationsCmd, nil, "recommendations"},
		{getRecommendationCmd, []string{"44444444-4444-4444-4444-000000000001"}, "recommendation"},
		{getOverviewCmd, nil, "organization"},
		{getTeamAssignmentsCmd, []string{teamIDForAssignments}, "team-assignments"},
	}
	for _, tc := range tests {
		t.Run(tc.endpoint, func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newListHarness(t, tc.cmd, server.URL, formatJSON, map[string]string{
				flagCostMode: "workload_only",
			})

			err := tc.cmd.RunE(tc.cmd, tc.args)
			require.Error(t, err)
			assert.Equal(t, "--cost-mode is not accepted by the "+tc.endpoint+" endpoint", err.Error())
			assert.Empty(t, server.Requests(),
				"the rejection is known locally; spending a round-trip to learn it wastes the rate budget")

			rejection, ok := errors.AsType[*api.CostModeUnsupportedError](err)
			require.True(t, ok, "each site must reuse the API's contract, not restate it inline")
			assert.Equal(t, tc.endpoint, rejection.Endpoint)
		})
	}
}

func TestNodeGroups_RejectsPaginationFlags(t *testing.T) {
	tests := []struct{ flag, value string }{
		{flagLimit, "2"},
		{flagCursor, "page2"},
		{flagPaginate, "true"},
		{flagIncludeTotal, "true"},
	}
	for _, tc := range tests {
		t.Run(tc.flag, func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newListHarness(t, getNodeGroupsCmd, server.URL, formatJSON, map[string]string{
				tc.flag: tc.value,
			})

			err := getNodeGroupsCmd.RunE(getNodeGroupsCmd, nil)
			require.Error(t, err, "--%s is advertised but the node-groups route has no pagination", tc.flag)
			assert.Contains(t, err.Error(), "--"+tc.flag)
			assert.Contains(t, err.Error(), "node-groups")
		})
	}
}

func TestNodeGroups_DefaultPaginationFlagsStillWork(t *testing.T) {
	server := testutil.NewMockServer(t)
	h := newListHarness(t, getNodeGroupsCmd, server.URL, formatJSON, nil)

	require.NoError(t, getNodeGroupsCmd.RunE(getNodeGroupsCmd, nil),
		"untouched defaults must not trip the rejection")
	assert.NotEqual(t, "[]", dataField(t, h.stdout.Bytes()))
}

func TestGetNamespace_ClusterIDAcceptsSliceForm(t *testing.T) {
	server := testutil.NewMockServer(t)
	h := newListHarness(t, getNamespaceCmd, server.URL, formatJSON, nil)
	require.NoError(t, getNamespaceCmd.Flags().Set("cluster-id", clusterIDForNamespace))

	require.NoError(t, getNamespaceCmd.RunE(getNamespaceCmd, []string{"default"}),
		"a single --cluster-id must keep working after the StringSlice switch")
	assert.NotEmpty(t, h.stdout.String())
}

func TestScopeCommands_ClusterIDRejectsMultipleValues(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "namespace", cmd: getNamespaceCmd, args: []string{"default"}},
		{name: "node-group", cmd: getNodeGroupCmd, args: []string{"general-purpose"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newListHarness(t, tt.cmd, server.URL, formatJSON, nil)
			require.NoError(t, tt.cmd.Flags().Set("cluster-id", clusterIDForNamespace))
			require.NoError(t, tt.cmd.Flags().Set("cluster-id", clusterIDSecond))

			err := tt.cmd.RunE(tt.cmd, tt.args)
			require.Error(t, err, "a %s lives in exactly one cluster", tt.name)
			assert.Contains(t, err.Error(), "exactly one")
		})
	}
}

func TestClusterIDFlag_UsageIsIdenticalAcrossFilterCommands(t *testing.T) {
	filters := map[string]*cobra.Command{
		"namespaces":       getNamespacesCmd,
		"node-groups":      getNodeGroupsCmd,
		"nodes":            getNodesCmd,
		"recommendations":  getRecommendationsCmd,
		"team-assignments": getTeamAssignmentsCmd,
		"workloads":        getWorkloadsCmd,
	}
	for name, c := range filters {
		f := c.Flags().Lookup("cluster-id")
		require.NotNil(t, f, "%s should define --cluster-id", name)
		assert.Equal(t, clusterIDFilterUsage, f.Usage, "%s documents --cluster-id differently", name)
		assert.NotContains(t, f.Usage, "scoped path",
			"%s leaks an internal routing detail into user-facing help", name)
	}
}

func TestClusterIDFlag_ScopeCommandsDocumentTheSingleValueRule(t *testing.T) {
	scopes := map[string]*cobra.Command{
		"namespace":  getNamespaceCmd,
		"node-group": getNodeGroupCmd,
	}
	for name, c := range scopes {
		f := c.Flags().Lookup("cluster-id")
		require.NotNil(t, f, "%s should define --cluster-id", name)
		assert.Equal(t, "stringSlice", f.Value.Type(),
			"%s must parse --cluster-id the same way the filter commands do", name)
		assert.Contains(t, f.Usage, "exactly one", "%s should say only one value is allowed", name)
	}
}

func TestRecommendationsFlags_UsageDescribesTheFilter(t *testing.T) {
	f := getRecommendationsCmd.Flags()
	for _, name := range []string{"recommendation-type", "status", "risk-level", "priority", "resource-type"} {
		flag := f.Lookup(name)
		require.NotNil(t, flag, "--%s should exist", name)
		assert.Contains(t, flag.Usage, "Filter by",
			"--%s usage is a bare enum; every sibling flag says what it filters", name)
	}
}

func TestSeverityFlags_ShareOneOrdering(t *testing.T) {
	f := getRecommendationsCmd.Flags()
	risk := f.Lookup("risk-level")
	priority := f.Lookup("priority")
	require.NotNil(t, risk)
	require.NotNil(t, priority)

	order := "(low|medium|high)"
	assert.Contains(t, risk.Usage, order)
	assert.Contains(t, priority.Usage, order,
		"risk-level and priority describe the same severity scale and must not invert it")
}

func TestTriStateBoolFlags_AreDocumentedAsTriState(t *testing.T) {
	tests := []struct {
		cmd  *cobra.Command
		flag string
	}{
		{getNodesCmd, "is-spot"},
		{getNodesCmd, "is-ready"},
		{getPodsCmd, "has-hostpath"},
		{getPodsCmd, "has-emptydir"},
		{getPodsCmd, "host-network"},
		{getWorkloadsCmd, "has-hpa"},
	}
	for _, tc := range tests {
		t.Run(tc.flag, func(t *testing.T) {
			f := tc.cmd.Flags().Lookup(tc.flag)
			require.NotNil(t, f)
			assert.Contains(t, f.Usage, "tri-state",
				"--%s=false differs from omitting it; that has to be discoverable", tc.flag)
		})
	}
}
