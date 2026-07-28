package testutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// Centralized to satisfy goconst and make renames a single edit.
const (
	costModeDefault  = "fully_loaded"
	missingPathValue = "missing"
	cursorPage2      = "page2"
	cursorPage3      = "page3"
	statusKey        = "status"
	statusOK         = "ok"
)

// Behavior-control fields may be mutated between requests to exercise error
// paths without re-creating the server. Safe for concurrent use: every field
// access goes through the embedded mutex.
type MockServer struct {
	*httptest.Server

	URL string

	mu sync.Mutex

	// When set, returns 401 unless the request carries a matching bearer token.
	RequireAPIKey string

	// Echoed on every response.
	RateLimitHeaders map[string]string

	// Applies to the next request only, then is consumed.
	ForceStatus int

	// Applies to the next request only. Picks a default status when ForceStatus
	// is zero.
	ForceError api.ErrorCode

	// Next N requests return 429 with Retry-After: 1. Decremented per request.
	RateLimitFails int

	RequestLog []RequestRecord
}

type RequestRecord struct {
	Method        string
	Path          string
	Query         url.Values
	Authorization string
}

type Option func(*MockServer)

func WithAPIKey(key string) Option {
	return func(ms *MockServer) { ms.RequireAPIKey = key }
}

// The map is copied, so later caller mutations do not affect the server.
func WithRateLimitHeaders(m map[string]string) Option {
	return func(ms *MockServer) {
		ms.RateLimitHeaders = make(map[string]string, len(m))
		for k, v := range m {
			ms.RateLimitHeaders[k] = v
		}
	}
}

// Registered with t.Cleanup, so teardown is automatic.
func NewMockServer(t *testing.T, opts ...Option) *MockServer {
	t.Helper()
	ms := &MockServer{
		RateLimitHeaders: map[string]string{},
		RequestLog:       []RequestRecord{},
	}
	for _, o := range opts {
		o(ms)
	}
	ms.Server = httptest.NewServer(ms.mux())
	ms.URL = ms.Server.URL
	t.Cleanup(ms.Close)
	return ms
}

func (ms *MockServer) Requests() []RequestRecord {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	out := make([]RequestRecord, len(ms.RequestLog))
	copy(out, ms.RequestLog)
	return out
}

// Applies the global behavior controls before dispatching to the route handler.
func (ms *MockServer) wrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ms.mu.Lock()
		ms.RequestLog = append(ms.RequestLog, RequestRecord{
			Method:        r.Method,
			Path:          r.URL.Path,
			Query:         r.URL.Query(),
			Authorization: r.Header.Get("Authorization"),
		})
		forceStatus := ms.ForceStatus
		forceErr := ms.ForceError
		ms.ForceStatus = 0
		ms.ForceError = ""
		rateFailsRemaining := ms.RateLimitFails
		if rateFailsRemaining > 0 {
			ms.RateLimitFails--
		}
		headersCopy := make(map[string]string, len(ms.RateLimitHeaders))
		for k, v := range ms.RateLimitHeaders {
			headersCopy[k] = v
		}
		requireKey := ms.RequireAPIKey
		ms.mu.Unlock()

		for k, v := range headersCopy {
			w.Header().Set(k, v)
		}

		if rateFailsRemaining > 0 {
			WriteError(w, http.StatusTooManyRequests, api.CodeRateLimited, "rate limited")
			return
		}
		if forceErr != "" {
			status := forceStatus
			if status == 0 {
				status = errorStatusFor(forceErr)
			}
			WriteError(w, status, forceErr, "forced error")
			return
		}
		if forceStatus != 0 {
			WriteError(w, forceStatus, api.ErrorCode(""), "forced status")
			return
		}
		if requireKey != "" {
			if r.Header.Get("Authorization") != "Bearer "+requireKey {
				WriteError(w, http.StatusUnauthorized, api.CodeUnauthorized, "missing or invalid bearer token")
				return
			}
		}
		h(w, r)
	}
}

// Health and OpenAPI routes are wrapped too, so RateLimitHeaders and RequestLog
// still apply to tests that count traffic.
func (ms *MockServer) mux() http.Handler {
	m := http.NewServeMux()

	m.HandleFunc("GET /health", ms.wrap(handleHealth))
	m.HandleFunc("GET /health/live", ms.wrap(handleHealthLive))
	m.HandleFunc("GET /health/ready", ms.wrap(handleHealthReady))
	m.HandleFunc("GET /v1/openapi.json", ms.wrap(handleOpenAPIJSON))
	m.HandleFunc("GET /v1/openapi.yaml", ms.wrap(handleOpenAPIYAML))
	m.HandleFunc("GET /v1/docs", ms.wrap(handleDocs))

	m.HandleFunc("GET /v1/organization", ms.wrap(handleOrganization))
	m.HandleFunc("GET /v1/organization/dashboard", ms.wrap(handleOrganizationDashboard))

	m.HandleFunc("GET /v1/clusters", ms.wrap(handleClustersList))
	m.HandleFunc("GET /v1/clusters/{cluster_id}", ms.wrap(handleClusterDetail))
	m.HandleFunc("GET /v1/clusters/{cluster_id}/namespaces", ms.wrap(handleClusterNamespaces))
	m.HandleFunc("GET /v1/clusters/{cluster_id}/namespaces/{namespace}", ms.wrap(handleClusterNamespaceDetail))
	m.HandleFunc("GET /v1/clusters/{cluster_id}/workloads", ms.wrap(handleClusterWorkloads))
	m.HandleFunc("GET /v1/clusters/{cluster_id}/nodes", ms.wrap(handleClusterNodes))
	m.HandleFunc("GET /v1/clusters/{cluster_id}/node-groups", ms.wrap(handleClusterNodeGroups))
	m.HandleFunc("GET /v1/clusters/{cluster_id}/node-groups/{name}", ms.wrap(handleClusterNodeGroupDetail))

	m.HandleFunc("GET /v1/namespaces", ms.wrap(handleNamespacesList))
	m.HandleFunc("GET /v1/workloads", ms.wrap(handleWorkloadsList))
	m.HandleFunc("GET /v1/workloads/{workload_uid}", ms.wrap(handleWorkloadDetail))
	m.HandleFunc("GET /v1/workloads/{workload_uid}/pods", ms.wrap(handleWorkloadPods))
	m.HandleFunc("GET /v1/nodes", ms.wrap(handleNodesList))
	m.HandleFunc("GET /v1/nodes/{node_uid}", ms.wrap(handleNodeDetail))
	m.HandleFunc("GET /v1/node-groups", ms.wrap(handleNodeGroupsList))

	m.HandleFunc("GET /v1/recommendations", ms.wrap(handleRecommendationsList))
	m.HandleFunc("GET /v1/recommendations/{rec_id}", ms.wrap(handleRecommendationDetail))

	m.HandleFunc("GET /v1/teams", ms.wrap(handleTeamsList))
	m.HandleFunc("GET /v1/teams/{team_id}", ms.wrap(handleTeamDetail))
	m.HandleFunc("GET /v1/teams/{team_id}/assignments", ms.wrap(handleTeamAssignments))

	m.HandleFunc("GET /v1/departments", ms.wrap(handleDepartmentsList))
	m.HandleFunc("GET /v1/departments/{dept_id}", ms.wrap(handleDepartmentDetail))

	return m
}

func resolveCostMode(q url.Values) string {
	if cm := q.Get("cost_mode"); cm != "" {
		return cm
	}
	return costModeDefault
}

// Triggers on any cost_mode= parameter, including an empty value.
func rejectCostMode(w http.ResponseWriter, r *http.Request, endpoint string) bool {
	if r.URL.Query().Has("cost_mode") {
		WriteError(w, http.StatusUnprocessableEntity, api.CodeInvalidCostMode,
			endpoint+" does not accept cost_mode",
			map[string]any{"field": "cost_mode", "allowed": []string{}})
		return true
	}
	return false
}

func metaWithCostMode(r *http.Request) types.Meta {
	m := defaultMeta()
	m.CostMode = resolveCostMode(r.URL.Query())
	return m
}

// Cursors are the fixed strings "page2"/"page3" so tests can assert on them.
// total_count is populated only when ?include_total=true is passed.
func paginate[T any](items []T, q url.Values) ([]T, types.Pagination) {
	limit := 100
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n >= 1 && n <= 500 {
			limit = n
		}
	}
	start := 0
	switch q.Get("cursor") {
	case "":
		start = 0
	case cursorPage2:
		start = limit
	case cursorPage3:
		start = limit * 2
	}
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	page := items[start:end]

	pag := types.Pagination{
		Limit:   limit,
		HasMore: end < len(items),
	}
	if pag.HasMore {
		pag.NextCursor = nextCursorAfter(q.Get("cursor"))
	}
	if q.Get("include_total") == "true" {
		total := len(items)
		pag.TotalCount = &total
	}
	return page, pag
}

func nextCursorAfter(current string) string {
	switch current {
	case "":
		return cursorPage2
	case cursorPage2:
		return cursorPage3
	default:
		return ""
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{statusKey: statusOK, "version": "mock"})
}

func handleHealthLive(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{statusKey: statusOK})
}

func handleHealthReady(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]any{
		statusKey: "ready",
		"checks":  map[string]string{"postgres": statusOK},
	})
}

func handleOpenAPIJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]any{
		"openapi": "3.1.0",
		"info":    map[string]string{"title": "mock"},
		"paths":   map[string]any{},
	})
}

func handleOpenAPIYAML(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write([]byte("openapi: 3.1.0\ninfo:\n  title: mock\npaths: {}\n"))
}

func handleDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<html><body>mock</body></html>"))
}

func handleOrganization(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "organization") {
		return
	}
	WriteEnvelope(w, SampleOrganization(), defaultMeta())
}

func handleOrganizationDashboard(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	dashboard := SampleOrganizationDashboard()

	limit := 5
	if v := q.Get("top_clusters_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 20 {
			limit = n
		}
	}
	if limit < len(dashboard.TopClusters) {
		dashboard.TopClusters = dashboard.TopClusters[:limit]
	}

	WriteEnvelope(w, dashboard, metaWithCostMode(r))
}

func handleClustersList(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "clusters") {
		return
	}
	page, pag := paginate(SampleClusters(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleClusterDetail(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "cluster") {
		return
	}
	if r.PathValue("cluster_id") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeClusterNotFound, "cluster not found")
		return
	}
	WriteEnvelope(w, SampleCluster(), defaultMeta())
}

func handleClusterNamespaces(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleNamespaces(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleClusterNamespaceDetail(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("namespace") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeNamespaceNotFound, "namespace not found")
		return
	}
	WriteEnvelope(w, SampleNamespace(), metaWithCostMode(r))
}

func handleClusterWorkloads(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleWorkloads(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleClusterNodes(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "nodes") {
		return
	}
	page, pag := paginate(SampleNodes(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleClusterNodeGroups(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "node-groups") {
		return
	}
	page, pag := paginate(SampleNodeGroups(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleClusterNodeGroupDetail(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "node-group") {
		return
	}
	if r.PathValue("name") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeNodeGroupNotFound, "node group not found")
		return
	}
	WriteEnvelope(w, SampleNodeGroup(), defaultMeta())
}

func handleNamespacesList(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleNamespaces(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleWorkloadsList(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleWorkloads(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleWorkloadDetail(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("workload_uid") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeWorkloadNotFound, "workload not found")
		return
	}
	WriteEnvelope(w, SampleWorkload(), metaWithCostMode(r))
}

func handleWorkloadPods(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SamplePods(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleNodesList(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "nodes") {
		return
	}
	page, pag := paginate(SampleNodes(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleNodeDetail(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "node") {
		return
	}
	if r.PathValue("node_uid") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeNodeNotFound, "node not found")
		return
	}
	WriteEnvelope(w, SampleNode(), defaultMeta())
}

func handleNodeGroupsList(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "node-groups") {
		return
	}
	page, pag := paginate(SampleNodeGroups(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleRecommendationsList(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "recommendations") {
		return
	}
	page, pag := paginate(SampleRecommendations(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleRecommendationDetail(w http.ResponseWriter, r *http.Request) {
	if rejectCostMode(w, r, "recommendation") {
		return
	}
	if r.PathValue("rec_id") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeRecommendationNotFound, "recommendation not found")
		return
	}
	WriteEnvelope(w, SampleRecommendation(), defaultMeta())
}

func handleTeamsList(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleTeams(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleTeamDetail(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("team_id") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeTeamNotFound, "team not found")
		return
	}
	WriteEnvelope(w, SampleTeam(), metaWithCostMode(r))
}

func handleTeamAssignments(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleTeamAssignments(), r.URL.Query())
	WritePaginated(w, page, defaultMeta(), pag)
}

func handleDepartmentsList(w http.ResponseWriter, r *http.Request) {
	page, pag := paginate(SampleDepartments(), r.URL.Query())
	WritePaginated(w, page, metaWithCostMode(r), pag)
}

func handleDepartmentDetail(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("dept_id") == missingPathValue {
		WriteError(w, http.StatusNotFound, api.CodeDepartmentNotFound, "department not found")
		return
	}
	WriteEnvelope(w, SampleDepartment(), metaWithCostMode(r))
}
