package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRecommendations(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.ListRecommendations(t.Context(), api.RecommendationFilter{
		RecommendationType: "workload_rightsizing",
		RiskLevel:          "low",
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one recommendation")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "workload_rightsizing", last.Query.Get("recommendation_type"))
	assert.False(t, last.Query.Has("cost_mode"), "cost_mode must not be sent to /recommendations")
}

func TestGetRecommendation(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetRecommendation(t.Context(), "rec-1")
	require.NoError(t, err)
	require.NotNil(t, got, "expected recommendation payload")
	assert.NotEmpty(t, got.ID, "expected recommendation payload")
}
