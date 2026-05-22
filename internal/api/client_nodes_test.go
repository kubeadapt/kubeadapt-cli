package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNodes(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	spot := true
	got, _, err := c.ListNodes(t.Context(), api.NodeFilter{
		ClusterIDs: []string{"cluster-prod-us-east"},
		IsSpot:     &spot,
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one node")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "true", last.Query.Get("is_spot"))
	assert.False(t, last.Query.Has("cost_mode"), "cost_mode must not be sent to /nodes")
}

func TestGetNode(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetNode(t.Context(), "node-uid-1")
	require.NoError(t, err)
	require.NotNil(t, got, "expected node payload")
	assert.NotEmpty(t, got.ID, "expected node payload")
}
