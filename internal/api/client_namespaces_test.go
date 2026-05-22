package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNamespaces_Flat(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, meta, err := c.ListNamespaces(t.Context(), api.NamespaceFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one namespace")
	require.NotNil(t, meta, "expected meta")
}

func TestListNamespaces_Scoped(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	_, _, err := c.ListNamespaces(t.Context(), api.NamespaceFilter{
		ClusterIDs:  []string{"cluster-prod-us-east"},
		CostModeOpt: api.CostModeOpt{CostMode: "workload_only"},
	})
	require.NoError(t, err)
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "/v1/clusters/cluster-prod-us-east/namespaces", last.Path)
	assert.Empty(t, last.Query.Get("cluster_id"), "scoped path should not carry cluster_id query")
}

func TestGetNamespace(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetNamespace(t.Context(), "cluster-prod-us-east", "default", api.NamespaceGetOpts{})
	require.NoError(t, err)
	require.NotNil(t, got, "expected namespace payload")
	assert.NotEmpty(t, got.ID, "expected namespace payload")
}
