package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListClusters(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, meta, err := c.ListClusters(t.Context(), api.ClusterFilter{
		PagedOpts: api.PagedOpts{Limit: 100},
		Provider:  "aws",
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one cluster")
	require.NotNil(t, meta, "expected pagination meta on list endpoint")
	assert.NotNil(t, meta.Pagination, "expected pagination meta on list endpoint")
}

func TestGetCluster(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, meta, err := c.GetCluster(t.Context(), "cluster-prod-us-east")
	require.NoError(t, err)
	require.NotNil(t, got, "expected cluster payload")
	assert.NotEmpty(t, got.ID, "expected cluster payload")
	require.NotNil(t, meta, "expected meta")
}
