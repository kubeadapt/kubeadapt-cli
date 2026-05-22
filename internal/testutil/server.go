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

// Common literal values reused across the route table. Centralized to satisfy
// the goconst linter and to make wholesale renames a single-edit operation.
const (
	costModeDefault    = "fully_loaded"
	missingPathValue   = "missing"
	cursorPage2        = "page2"
	cursorPage3        = "page3"
	statusKey          = "status"
	statusOK           = "ok"
)

// MockServer is an in-process httptest.Server preconfigured with every
// kubeadapt-cli endpoint. Tests can mutate the behavior-control fields
// (RequireAPIKey, ForceStatus, ForceError, RateLimitFails, RateLimitHeaders)
// between requests to exercise error paths without re-creating the server.
//
// MockServer instances are safe for concurrent use by HTTP handlers because
// every field read/write goes through the embedded mutex.
type MockServer struct {
	*httptest.Server

	// URL is a convenience alias for Server.URL.
	URL string

	mu sync.Mutex

	// RequireAPIKey, when set, causes the server to return 401 UNAUTHORIZED
	// unless the client sends `Authorization: Bearer <RequireAPIKey>`.
	RequireAPIKey string

	// RateLimitHeaders are echoed on every response — useful for testing the
	// client's RateLimit snapshot machinery.
	RateLimitHeaders map[string]string

	// ForceStatus, when non-zero, makes the next request return this HTTP
	// status. The next-request flag is consumed after one use.
	ForceStatus int

	// ForceError, when non-empty, makes the next request return an envelope
	// error with this code (and a matching default HTTP status when
	// ForceStatus is zero). Consumed after one use.
	ForceError api.ErrorCode

	// RateLimitFails, when > 0, makes the next N requests return 429
	// RATE_LIMITED with Retry-After: 1. Decremented per request.
	RateLimitFails int

	// RequestLog captures every URL + headers seen, for assertion in tests.
	RequestLog []RequestRecord
}

// RequestRecord captures the salient fields of every request the mock server
// has seen, in the order they arrived.
type RequestRecord struct {
	Method        string
	Path          string
	Query         url.Values
	Authorization string
}

// Option configures a MockServer at construction time.
type Option func(*MockServer)

// WithAPIKey requires every incoming request to present the supplied bearer
// token in the Authorization header. Requests without a matching token are
// rejected with 401 UNAUTHORIZED.
func WithAPIKey(key string) Option {
	return func(ms *MockServer) { ms.RequireAPIKey = key }
}

// WithRateLimitHeaders attaches the supplied headers to every response. The
// map is copied so subsequent mutations of the caller's map do not affect
// the server.
func WithRateLimitHeaders(m map[string]string) Option {
	return func(ms *MockServer) {
		ms.RateLimitHeaders = make(map[string]string, len(m))
		for k, v := range m {
			ms.RateLimitHeaders[k] = v
		}
	}
}

// NewMockServer starts an in-process HTTP server that responds to every
// kubeadapt-cli endpoint with realistic envelope-shaped responses. The
// server is registered with t.Cleanup so it is torn down automatically.
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

// Requests returns a copy of the request log. Safe for concurrent use.
func (ms *MockServer) Requests() []RequestRecord {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	out := make([]RequestRecord, len(ms.RequestLog))
	copy(out, ms.RequestLog)
	return out
}

// wrap returns an http.HandlerFunc that applies the global behavior controls
// (request logging, RateLimitHeaders echo, ForceStatus/ForceError/RateLimitFails,
// RequireAPIKey) before dispatching to the supplied per-route handler.
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

// mux constructs the ServeMux with every supported route registered. Health
// and OpenAPI routes are also wrapped so RateLimitHeaders + RequestLog still
// apply to them (useful for tests that count traffic).
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

// resolveCostMode returns the cost_mode supplied via query string, or the
// default ("fully_loaded") when absent.
func resolveCostMode(q url.Values) string {
	if cm := q.Get("cost_mode"); cm != "" {
		return cm
	}
	return costModeDefault
}

// rejectCostMode writes a 422 INVALID_COST_MODE response and returns true
// when the request includes a cost_mode= query parameter (including empty
// values). The endpoint argument is interpolated into the error message.
func rejectCostMode(w http.ResponseWriter, r *http.Request, endpoint string) bool {
	if r.URL.Query().Has("cost_mode") {
		WriteError(w, http.StatusUnprocessableEntity, api.CodeInvalidCostMode,
			endpoint+" does not accept cost_mode",
			map[string]any{"field": "cost_mode", "allowed": []string{}})
		return true
	}
	return false
}

// metaWithCostMode returns a Meta block with cost_mode echoed from the request
// query (or the default).
func metaWithCostMode(r *http.Request) types.Meta {
	m := defaultMeta()
	m.CostMode = resolveCostMode(r.URL.Query())
	return m
}

// paginate slices a fixture list according to the limit/cursor query and
// returns the page plus the corresponding Pagination block. Cursors are the
// deterministic strings "page2"/"page3" so tests can assert against them.
//
// The total_count field on Pagination is populated only when the caller
// passes ?include_total=true.
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
