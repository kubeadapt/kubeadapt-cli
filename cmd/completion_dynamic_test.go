package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Fixture identities mirrored from internal/testutil. Completions are asserted
// against the literal "<id>\t<name>" a shell would render.
const (
	compClusterProdID    = "00000000-0000-0000-0000-000000000001"
	compClusterStagingID = "00000000-0000-0000-0000-000000000002"
	compClusterDevID     = "00000000-0000-0000-0000-000000000003"
	compClusterProdName  = "prod-cluster"
)

// newCompletionHarness points cmd at serverURL for one completion call.
//
// The subcommands are package-level singletons and cobra 1.9.1 only propagates
// a parent context into a subcommand whose ctx is still nil, so both the
// context and any flag mutation leak across in-process runs unless restored.
func newCompletionHarness(t *testing.T, cmd *cobra.Command, serverURL string) {
	t.Helper()

	cmd.SetContext(context.WithValue(t.Context(), runContextKey{}, &RunContext{
		Config:    &config.Config{APIURL: serverURL, APIKey: "test-key"},
		Logger:    zap.NewNop(),
		OutputFmt: formatJSON,
		NoColor:   true,
	}))

	t.Cleanup(func() {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed {
				return
			}
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
		cmd.SetContext(nil)
	})
}

// isolateCredentials removes every ambient source of an API key so a test that
// expects "unauthenticated" cannot silently reach the real API instead.
func isolateCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("KUBEADAPT_API_KEY", "")
	t.Setenv("KUBEADAPT_API_URL", "")

	prevCfg, prevKey, prevURL := cfgFile, apiKey, apiURL
	cfgFile = filepath.Join(t.TempDir(), "absent.yaml")
	apiKey, apiURL = "", ""
	t.Cleanup(func() { cfgFile, apiKey, apiURL = prevCfg, prevKey, prevURL })
}

func TestCompletion_ClusterIDsFromAPI(t *testing.T) {
	server := testutil.NewMockServer(t)
	newCompletionHarness(t, getClusterCmd, server.URL)

	require.NotNil(t, getClusterCmd.ValidArgsFunction,
		"`get cluster` takes a UUID; without completion the only workflow is copy-paste from `get clusters`")

	got, directive := getClusterCmd.ValidArgsFunction(getClusterCmd, nil, "")

	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive,
		"falling back to filenames when an ID is expected is worse than no completion")
	assert.Contains(t, got, compClusterProdID+"\t"+compClusterProdName,
		"a bare UUID list is unusable; the cluster name must ride along as the description")
	assert.Contains(t, got, compClusterStagingID+"\tstaging-cluster")
	assert.Contains(t, got, compClusterDevID+"\tdev-cluster")
}

func TestCompletion_DetailCommandsAllComplete(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *cobra.Command
		wantID  string
		wantDsc string
	}{
		{"cluster", getClusterCmd, compClusterProdID, compClusterProdName},
		{"workload", getWorkloadCmd, "11111111-1111-1111-1111-000000000001", "api-gateway"},
		{"pods", getPodsCmd, "11111111-1111-1111-1111-000000000002", "prometheus"},
		{"node", getNodeCmd, "33333333-3333-3333-3333-000000000001", "ip-10-0-1-42.ec2.internal"},
		{"recommendation", getRecommendationCmd, "44444444-4444-4444-4444-000000000002", "Right-size prometheus memory limit"},
		{"team", getTeamCmd, "55555555-5555-5555-5555-000000000001", "platform"},
		{"team-assignments", getTeamAssignmentsCmd, "55555555-5555-5555-5555-000000000002", "sre"},
		{"department", getDepartmentCmd, "66666666-6666-6666-6666-000000000001", "engineering"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newCompletionHarness(t, tc.cmd, server.URL)

			require.NotNil(t, tc.cmd.ValidArgsFunction, "`get %s` takes an opaque ID", tc.name)
			got, directive := tc.cmd.ValidArgsFunction(tc.cmd, nil, "")

			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
			assert.Contains(t, got, tc.wantID+"\t"+tc.wantDsc)
		})
	}
}

func TestCompletion_FiltersByPrefix(t *testing.T) {
	server := testutil.NewMockServer(t)
	newCompletionHarness(t, getClusterCmd, server.URL)
	require.NotNil(t, getClusterCmd.ValidArgsFunction)

	// The three cluster fixtures diverge only in the final character, so a
	// short prefix would pass without the filter ever running.
	got, _ := getClusterCmd.ValidArgsFunction(getClusterCmd, nil, compClusterStagingID)

	assert.Equal(t, []cobra.Completion{compClusterStagingID + "\tstaging-cluster"}, got,
		"the shell filters on the raw value, so unmatched IDs are pure noise")
}

func TestCompletion_SecondArgOffersNothing(t *testing.T) {
	server := testutil.NewMockServer(t)
	newCompletionHarness(t, getClusterCmd, server.URL)
	require.NotNil(t, getClusterCmd.ValidArgsFunction)

	got, directive := getClusterCmd.ValidArgsFunction(getClusterCmd, []string{compClusterProdID}, "")

	assert.Empty(t, got, "`get cluster` is ExactArgs(1); a second value is never valid")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Zero(t, len(server.Requests()), "completing an impossible argument must not cost a request")
}

func TestCompletion_FailsSilentlyWhenUnauthenticated(t *testing.T) {
	isolateCredentials(t)
	getClusterCmd.SetContext(nil)
	require.NotNil(t, getClusterCmd.ValidArgsFunction)

	stdout, stderr := captureOSStreams(t, func() {
		got, directive := getClusterCmd.ValidArgsFunction(getClusterCmd, nil, "")
		assert.Empty(t, got)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive,
			"an auth error must not leak into the completion stream as a directive")
	})

	assert.Empty(t, stdout, "anything printed here is injected straight into the user's completion list")
	assert.Empty(t, stderr, "a TAB press must never surface an auth error")
}

func TestCompletion_FailsSilentlyOnTimeout(t *testing.T) {
	blocked := make(chan struct{})
	hung := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-blocked
	}))
	t.Cleanup(func() { close(blocked); hung.Close() })

	prev := completionTimeout
	completionTimeout = 100 * time.Millisecond
	t.Cleanup(func() { completionTimeout = prev })

	newCompletionHarness(t, getClusterCmd, hung.URL)
	require.NotNil(t, getClusterCmd.ValidArgsFunction)

	done := make(chan struct{})
	var got []cobra.Completion
	go func() {
		defer close(done)
		got, _ = getClusterCmd.ValidArgsFunction(getClusterCmd, nil, "")
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("completion blocked the shell waiting on an unresponsive API")
	}
	assert.Empty(t, got)
}

func TestCompletion_CapsResultCount(t *testing.T) {
	require.LessOrEqual(t, completionMaxResults, 200,
		"an org with 3891 workloads makes an uncapped completion list unusable")
}

func TestCompletion_ClusterIDFlagCompletesFromAPI(t *testing.T) {
	for _, c := range []*cobra.Command{
		getNodesCmd, getNodeGroupsCmd, getNamespacesCmd, getNamespaceCmd,
		getNodeGroupCmd, getRecommendationsCmd, getTeamAssignmentsCmd, getWorkloadsCmd,
	} {
		t.Run(c.Name(), func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newCompletionHarness(t, c, server.URL)

			fn, ok := c.GetFlagCompletionFunc("cluster-id")
			require.True(t, ok, "--cluster-id takes a UUID and must complete from live clusters")

			got, directive := fn(c, nil, "")
			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
			assert.Contains(t, got, compClusterProdID+"\t"+compClusterProdName)
		})
	}
}

func TestCompletion_StaticEnumFlags(t *testing.T) {
	tests := []struct {
		cmd  *cobra.Command
		flag string
		want []string
	}{
		{getCmd, flagCostMode, []string{"fully_loaded", "workload_only"}},
		{getClustersCmd, "provider", []string{"aws", "gcp", "azure", "on-prem"}},
		{getClustersCmd, "environment", []string{"production", "non-production", "staging", "dev"}},
		{getClustersCmd, "status", []string{"pending", "active", "disconnected", "error", "discovered"}},
		{getRecommendationsCmd, "status", []string{"pending", "applied", "dismissed", "archived"}},
		{getRecommendationsCmd, "priority", []string{"low", "medium", "high"}},
		{getRecommendationsCmd, "risk-level", []string{"low", "medium", "high"}},
		{getRecommendationsCmd, "recommendation-type", []string{"workload_rightsizing"}},
		{getRecommendationsCmd, "resource-type", []string{"Deployment", "StatefulSet", "DaemonSet", "Pod", "Node"}},
		{getNodesCmd, "architecture", []string{"amd64", "arm64"}},
		{getNodesCmd, "capacity-type", []string{"on-demand", "spot"}},
		{getPodsCmd, "phase", []string{"Pending", "Running", "Succeeded", "Failed", "Unknown"}},
		{getPodsCmd, "qos-class", []string{"Guaranteed", "Burstable", "BestEffort"}},
		{getWorkloadsCmd, "kind", []string{"Deployment", "StatefulSet", "DaemonSet"}},
		{getTeamsCmd, "origin", []string{"k8s", "kubeadapt"}},
		{getDepartmentsCmd, "origin", []string{"k8s", "kubeadapt"}},
		{getTeamAssignmentsCmd, "entity-type", []string{"namespace", "workload", "cluster"}},
	}
	for _, tc := range tests {
		t.Run(tc.cmd.Name()+"/"+tc.flag, func(t *testing.T) {
			fn, ok := tc.cmd.GetFlagCompletionFunc(tc.flag)
			require.True(t, ok, "--%s is a closed enum; the shell should offer the values", tc.flag)

			got, directive := fn(tc.cmd, nil, "")
			assert.Equal(t, tc.want, got)
			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
		})
	}
}

func TestCompletion_StaticEnumFlagsFilterByPrefix(t *testing.T) {
	fn, ok := getPodsCmd.GetFlagCompletionFunc("phase")
	require.True(t, ok)

	got, _ := fn(getPodsCmd, nil, "F")
	assert.Equal(t, []string{"Failed"}, got)
}

// captureOSStreams swaps the process stdout/stderr for the duration of fn. A
// completion function writes into the shell's candidate list, so silence has
// to be proven at the file-descriptor level, not on cmd.OutOrStdout().
func captureOSStreams(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	outR, outW, err := os.Pipe()
	require.NoError(t, err)
	errR, errW, err := os.Pipe()
	require.NoError(t, err)

	prevOut, prevErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	fn()
	os.Stdout, os.Stderr = prevOut, prevErr

	require.NoError(t, outW.Close())
	require.NoError(t, errW.Close())
	return readAllString(t, outR), readAllString(t, errR)
}

func readAllString(t *testing.T, f *os.File) string {
	t.Helper()
	defer func() { _ = f.Close() }()
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	return string(buf[:n])
}
