package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNodeGroups(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.ListNodeGroups(t.Context(), api.NodeGroupFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one node group")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "/v1/node-groups", last.Path)
}

func TestGetNodeGroup(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetNodeGroup(t.Context(), "cluster-prod-us-east", "general-purpose")
	require.NoError(t, err)
	require.NotNil(t, got, "expected node group payload")
	assert.NotEmpty(t, got.ID, "expected node group payload")
}
