package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrganization(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, meta, err := c.GetOrganization(t.Context())
	require.NoError(t, err)
	require.NotNil(t, got, "expected organization payload")
	assert.NotEmpty(t, got.ID, "expected organization id")
	require.NotNil(t, meta, "expected meta with request_id")
	assert.NotEmpty(t, meta.RequestID, "expected meta with request_id")
}

func TestGetOrganizationDashboard(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, meta, err := c.GetOrganizationDashboard(t.Context(), api.OrganizationDashboardOpts{
		CostModeOpt:      api.CostModeOpt{CostMode: "fully_loaded"},
		TopClustersLimit: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, got, "expected dashboard payload")
	assert.LessOrEqual(t, len(got.TopClusters), 3, "top_clusters_limit=3 not honored")
	require.NotNil(t, meta, "expected meta")
	assert.Equal(t, "fully_loaded", meta.CostMode)
}
