# CLI Test Suite Overhaul — Learnings

## [2026-03-11] Task 1: Error Path Tests for internal/api

### Patterns & Conventions
- Package declaration: `package api` (same-package tests, not `_test` suffix)
- No testify — stdlib `testing` only: `t.Fatal`, `t.Fatalf`, `t.Errorf`
- No `t.Parallel()` in this package
- No helper functions — raw assertions inline
- `errors.AsType[*APIError](err)` — Go 1.26 generic, available for type-asserting wrapped errors
- Per-test `httptest.NewServer` pattern for client error tests (not shared testutil mock)
- NetworkError test: create server, call `server.Close()` BEFORE request
- ContextCancelled test: `cancel()` BEFORE making the request

### Import Pattern for Per-Test Servers
```go
import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/kubeadapt/kubeadapt-cli/internal/testutil"
)
```

### client.go Error Handling (lines 107-118)
- Status >= 400: creates `&APIError{StatusCode: resp.StatusCode}`
- Tries `json.Unmarshal(respBody, apiErr)` — if fails, sets `apiErr.Message = string(respBody)`
- Non-JSON error body (plain text) → `apiErr.Message` = raw string

### Coverage Results
- errors.go: **100%** (all 6 functions covered)
- NetworkError produces a non-*APIError (wrapped transport error)
- MalformedJSON on 200 response produces unmarshal error, not *APIError

## [2026-03-11] Task 7: doRequest Edge Cases
- 204 NoContent: client.go:126 skips unmarshal when StatusNoContent — resp pointer is non-nil but zero-valued
- WithTimeout(1ms) + server sleeping 50ms reliably triggers timeout
- Empty body on 200: json.Unmarshal("", &struct{}) returns EOF error — err is non-nil
- Need "time" and "go.uber.org/zap" imports in client_test.go
- All 4 tests pass with -race flag: TestClient_204NoContent, TestClient_WithTimeout, TestClient_WithLogger, TestClient_EmptyResponseBody

## [2026-03-11] Task 6: Representative Untested Methods
- testutil.NewMockServer() has routes for /v1/costs/teams, /v1/node-groups, /v1/persistent-volumes
- All return empty arrays — resp non-nil check is sufficient
- GetPersistentVolumes signature: (ctx, clusterID, namespace, storageClass string)
- GetNodeGroups signature: (ctx, clusterID string)
- GetCostsTeams signature: (ctx, clusterID string)
- Removed unused types import from client_test.go (was causing build failure)
- All 3 tests follow TestGetClusters pattern exactly: setup server, create client, call method, assert err nil and resp non-nil

## [2026-03-11] Task 8: Render Smoke Tests
- All Render*() functions accept noColor bool as last param
- testutil.Sample*() functions provide test data
- Empty slice tests: pass []types.XxxResponse{} — no panic expected
- noColor=true avoids ANSI codes in test output
- output/ coverage after Task 8: 90.6% (exceeds 75% requirement)
- Enhanced TestRenderOverview with noColor=false test
- Enhanced TestRenderClusters with empty slice test
- All 7 render smoke tests pass with -race flag

## [2026-03-11] Task 5: Query Parameter Validation
- capturedParams pattern: declare `var capturedParams url.Values` outside handler, assign inside
- Per-test servers for param tests do NOT need auth check — simplifies handler
- offset=0 is NOT sent (conditional `if offset > 0`) — TestGetNodes_WithFilters verifies this
- Need to add "net/url" and "strings" imports to client_test.go

## [2026-03-11] Task 10: cmd/get_clusters_test.go
- Call getClustersCmd.RunE(cmd, nil) with a proper cobra.Command that has context set
- Create cmd with `cmd := &cobra.Command{}` and set context with `cmd.SetContext(context.Background())`
- Set cfg, log, outputFmt, noColor globals before calling RunE
- Reset ALL globals in t.Cleanup() — critical for test isolation
- Capture stdout with os.Pipe() for JSON output verification
- No t.Parallel() — global state is shared across tests
- newAPIClient() checks cfg.APIKey != "" — empty key returns error
- All 3 tests pass: TableOutput, JSONOutput, NoAPIKey

## [2026-03-11] Task 9: cmd/root_test.go Config Override Chain
- Call rootCmd.PersistentPreRunE(dummyCmd, nil) directly — avoids os.Exit from Execute()
- dummyCmd must NOT be named "login"/"version"/"completion" — use "status" or "clusters"
- Reset ALL globals in t.Cleanup(): cfg, log, apiURL, apiKey, cfgFile, outputFmt, noColor, verbose
- t.Setenv() auto-restores env vars after test
- defaultAPIURL in config.go is "http://localhost:8002"
- No t.Parallel() — global state is shared
- Use a non-existent path in TempDir for "missing config" tests — avoids reading real ~/.kubeadapt/config.yaml
- Go test build cache can mask pre-existing syntax errors; running with -run filter forces recompile
