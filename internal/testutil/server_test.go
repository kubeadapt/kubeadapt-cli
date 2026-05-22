package testutil_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
)

func TestMockServer_OrganizationEnvelope(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/v1/organization")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var env types.Envelope[types.Organization]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))

	assert.Nil(t, env.Error)
	assert.NotEmpty(t, env.Data.Metadata.Name)
	assert.Equal(t, "USD", env.Data.Cost.CurrentRunRateHourly.Currency)
	assert.NotEmpty(t, env.Meta.RequestID)
}

func TestMockServer_ClustersListIsPaginated(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/v1/clusters")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var env types.Envelope[[]types.Cluster]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Len(t, env.Data, 3)
	require.NotNil(t, env.Meta.Pagination)
	assert.False(t, env.Meta.Pagination.HasMore)
}

func TestMockServer_RejectsCostModeOnClusters(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/v1/clusters?cost_mode=workload_only")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	var env types.Envelope[map[string]any]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	require.NotNil(t, env.Error)
	assert.Equal(t, string(api.CodeInvalidCostMode), env.Error.Code)
}

func TestMockServer_ClusterDetailMissingReturns404(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/v1/clusters/missing")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var env types.Envelope[map[string]any]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	require.NotNil(t, env.Error)
	assert.Equal(t, string(api.CodeClusterNotFound), env.Error.Code)
}

func TestMockServer_DashboardEchoesCostMode(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/v1/organization/dashboard?cost_mode=workload_only&top_clusters_limit=2")
	require.NoError(t, err)
	defer resp.Body.Close()

	var env types.Envelope[types.OrganizationDashboard]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Equal(t, "workload_only", env.Meta.CostMode)
	assert.Len(t, env.Data.TopClusters, 2)
}

func TestMockServer_RequireAPIKeyRejectsMissing(t *testing.T) {
	ms := testutil.NewMockServer(t, testutil.WithAPIKey("secret-key"))

	resp, err := http.Get(ms.URL + "/v1/organization")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMockServer_RequireAPIKeyAcceptsBearer(t *testing.T) {
	ms := testutil.NewMockServer(t, testutil.WithAPIKey("secret-key"))

	req, _ := http.NewRequest(http.MethodGet, ms.URL+"/v1/organization", nil)
	req.Header.Set("Authorization", "Bearer secret-key")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_PaginationViaCursor(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/v1/clusters?limit=2")
	require.NoError(t, err)
	defer resp.Body.Close()
	var page1 types.Envelope[[]types.Cluster]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page1))
	assert.Len(t, page1.Data, 2)
	require.NotNil(t, page1.Meta.Pagination)
	assert.True(t, page1.Meta.Pagination.HasMore)
	assert.Equal(t, "page2", page1.Meta.Pagination.NextCursor)

	resp2, err := http.Get(ms.URL + "/v1/clusters?limit=2&cursor=page2")
	require.NoError(t, err)
	defer resp2.Body.Close()
	var page2 types.Envelope[[]types.Cluster]
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&page2))
	assert.Len(t, page2.Data, 1)
	require.NotNil(t, page2.Meta.Pagination)
	assert.False(t, page2.Meta.Pagination.HasMore)
}

func TestMockServer_RateLimitFails(t *testing.T) {
	ms := testutil.NewMockServer(t)
	ms.RateLimitFails = 1

	resp, err := http.Get(ms.URL + "/v1/organization")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "1", resp.Header.Get("Retry-After"))

	resp2, err := http.Get(ms.URL + "/v1/organization")
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestMockServer_HealthIsUnenveloped(t *testing.T) {
	ms := testutil.NewMockServer(t)

	resp, err := http.Get(ms.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var raw map[string]string
	require.NoError(t, json.Unmarshal(body, &raw), "body=%s", body)
	assert.Equal(t, "ok", raw["status"])
	assert.Equal(t, "mock", raw["version"])
}
