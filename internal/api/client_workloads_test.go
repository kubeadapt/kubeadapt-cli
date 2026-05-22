package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWorkloads(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	hpa := true
	got, _, err := c.ListWorkloads(t.Context(), api.WorkloadFilter{
		ClusterIDs: []string{"cluster-prod-us-east"},
		Namespaces: []string{"default", "monitoring"},
		Kinds:      []string{"Deployment"},
		HasHPA:     &hpa,
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one workload")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "/v1/clusters/cluster-prod-us-east/workloads", last.Path)
	assert.Equal(t, "default,monitoring", last.Query.Get("namespace"))
	assert.Equal(t, "true", last.Query.Get("has_hpa"))
}

func TestGetWorkload(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetWorkload(t.Context(), "wl-uid-1", api.WorkloadGetOpts{})
	require.NoError(t, err)
	require.NotNil(t, got, "expected workload payload")
	assert.NotEmpty(t, got.ID, "expected workload payload")
}

func TestListWorkloadPods(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.ListWorkloadPods(t.Context(), "wl-uid-1", api.PodFilter{
		Phase:    "Running",
		QoSClass: "Burstable",
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one pod")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "Running", last.Query.Get("phase"))
}
