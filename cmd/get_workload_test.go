package cmd

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const clusterIDStaging = "00000000-0000-0000-0000-000000000002"

func TestClusterIDFlag_SentOnDetailAndPods(t *testing.T) {
	tests := []struct {
		cmd *cobra.Command
		uid string
	}{
		{getWorkloadCmd, testutil.AmbiguousWorkloadUID},
		{getNodeCmd, testutil.AmbiguousNodeUID},
		{getPodsCmd, workloadUIDForPods},
	}
	for _, tt := range tests {
		t.Run(tt.cmd.Name(), func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newListHarness(t, tt.cmd, server.URL, formatJSON, map[string]string{"cluster-id": clusterIDStaging})

			require.NoError(t, tt.cmd.RunE(tt.cmd, []string{tt.uid}))

			reqs := server.Requests()
			require.Len(t, reqs, 1)
			assert.Equal(t, clusterIDStaging, reqs[0].Query.Get("cluster_id"))
		})
	}
}

func TestAmbiguousUID_ListsClusterIDFlags(t *testing.T) {
	tests := []struct {
		cmd      *cobra.Command
		uid      string
		resource string
	}{
		{getWorkloadCmd, testutil.AmbiguousWorkloadUID, "workload"},
		{getNodeCmd, testutil.AmbiguousNodeUID, "node"},
	}
	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			server := testutil.NewMockServer(t)
			newListHarness(t, tt.cmd, server.URL, formatJSON, nil)

			err := tt.cmd.RunE(tt.cmd, []string{tt.uid})

			require.Error(t, err)
			assert.Equal(t, exitError, exitCodeFor(err))
			msg := friendlyError(err)
			assert.Contains(t, msg, "This "+tt.resource+" UID exists in more than one cluster")
			assert.Contains(t, msg, "--cluster-id "+clusterIDStaging)
		})
	}
}
