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

// getEnvelope issues a GET and decodes into the real production envelope type.
// Decoding into types.Envelope[T] rather than a loose map is the point: it
// proves the mock emits JSON the CLI can actually consume, so a drifted fixture
// fails here instead of silently making every downstream test lie.
func getEnvelope[T any](t *testing.T, url string) (types.Envelope[T], int) {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	var env types.Envelope[T]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	return env, resp.StatusCode
}

func requireErrorEnvelope(t *testing.T, url string, wantStatus int, wantCode string) {
	t.Helper()
	env, status := getEnvelope[map[string]any](t, url)
	assert.Equal(t, wantStatus, status)
	require.NotNil(t, env.Error)
	assert.Equal(t, wantCode, env.Error.Code)
}

func TestMockServer_OrganizationEnvelope(t *testing.T) {
	ms := testutil.NewMockServer(t)

	env, status := getEnvelope[types.Organization](t, ms.URL+"/v1/organization")
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, env.Error)
	assert.NotEmpty(t, env.Data.Metadata.Name)
	assert.Equal(t, "USD", env.Data.Cost.CurrentRunRateHourly.Currency)
	assert.NotEmpty(t, env.Meta.RequestID)
}

func TestMockServer_ClustersListIsPaginated(t *testing.T) {
	ms := testutil.NewMockServer(t)

	env, status := getEnvelope[[]types.Cluster](t, ms.URL+"/v1/clusters")
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, env.Data, 3)
	require.NotNil(t, env.Meta.Pagination)
	assert.False(t, env.Meta.Pagination.HasMore)
}

func TestMockServer_DashboardEchoesCostMode(t *testing.T) {
	ms := testutil.NewMockServer(t)

	env, _ := getEnvelope[types.OrganizationDashboard](t,
		ms.URL+"/v1/organization/dashboard?cost_mode=workload_only&top_clusters_limit=2")
	assert.Equal(t, "workload_only", env.Meta.CostMode)
	assert.Len(t, env.Data.TopClusters, 2)
}

// The CLI's error-handling tests assert on these exact statuses and codes. If
// the mock stops producing them those tests pass against nothing.
func TestMockServer_RejectsCostModeOnClusters(t *testing.T) {
	ms := testutil.NewMockServer(t)
	requireErrorEnvelope(t, ms.URL+"/v1/clusters?cost_mode=workload_only",
		http.StatusUnprocessableEntity, string(api.CodeInvalidCostMode))
}

func TestMockServer_ClusterDetailMissingReturns404(t *testing.T) {
	ms := testutil.NewMockServer(t)
	requireErrorEnvelope(t, ms.URL+"/v1/clusters/missing",
		http.StatusNotFound, string(api.CodeClusterNotFound))
}

func TestMockServer_RequireAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		bearer     string
		wantStatus int
	}{
		{"absent key rejected", "", http.StatusUnauthorized},
		{"bearer key accepted", "secret-key", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := testutil.NewMockServer(t, testutil.WithAPIKey("secret-key"))

			req, err := http.NewRequest(http.MethodGet, ms.URL+"/v1/organization", nil)
			require.NoError(t, err)
			if tt.bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tt.bearer)
			}
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// cmd tests walk this cursor chain by hand (notably "page3" to reach a
// zero-row page), so the endpoint-bound cursor contract has to hold here.
func TestMockServer_PaginationViaCursor(t *testing.T) {
	ms := testutil.NewMockServer(t)

	page1, _ := getEnvelope[[]types.Cluster](t, ms.URL+"/v1/clusters?limit=2")
	assert.Len(t, page1.Data, 2)
	require.NotNil(t, page1.Meta.Pagination)
	assert.True(t, page1.Meta.Pagination.HasMore)
	assert.Equal(t, "page2", page1.Meta.Pagination.NextCursor)

	page2, _ := getEnvelope[[]types.Cluster](t, ms.URL+"/v1/clusters?limit=2&cursor=page2")
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

// /health is deliberately un-enveloped; the health command parses it raw.
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
