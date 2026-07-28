package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type pageItem struct {
	ID string `json:"id"`
}

// pageSpec is one scripted response in a fake paginated endpoint.
type pageSpec struct {
	status     int
	items      []pageItem
	nextCursor string
	hasMore    bool
	headers    map[string]string
	errCode    api.ErrorCode
	errMessage string
	total      *int
	ignored    []string
}

// fakeAPI serves a scripted sequence of pages and records the cursor seen on
// every request so tests can prove a page was not re-fetched.
type fakeAPI struct {
	t       *testing.T
	mu      sync.Mutex
	pages   []pageSpec
	fallth  *pageSpec
	cursors []string
	calls   int
}

func (f *fakeAPI) seenCursors() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.cursors...)
}

func (f *fakeAPI) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	idx := f.calls
	f.calls++
	f.cursors = append(f.cursors, r.URL.Query().Get("cursor"))
	var spec pageSpec
	switch {
	case idx < len(f.pages):
		spec = f.pages[idx]
	case f.fallth != nil:
		spec = *f.fallth
	default:
		f.mu.Unlock()
		http.Error(w, "unexpected extra request", http.StatusInternalServerError)
		return
	}
	f.mu.Unlock()

	for k, v := range spec.headers {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")

	status := spec.status
	if status == 0 {
		status = http.StatusOK
	}
	env := types.Envelope[[]pageItem]{
		Data: spec.items,
		Meta: types.Meta{
			RequestID:     fmt.Sprintf("req-%d", idx),
			IgnoredParams: spec.ignored,
			Pagination: &types.Pagination{
				NextCursor: spec.nextCursor,
				HasMore:    spec.hasMore,
				Limit:      5,
				TotalCount: spec.total,
			},
		},
	}
	if spec.errCode != "" {
		env.Error = &types.APIErrorBody{Code: string(spec.errCode), Message: spec.errMessage}
	}
	w.WriteHeader(status)
	require.NoError(f.t, json.NewEncoder(w).Encode(env))
}

func okPage(ids []string, next string) pageSpec {
	items := make([]pageItem, 0, len(ids))
	for _, id := range ids {
		items = append(items, pageItem{ID: id})
	}
	return pageSpec{items: items, nextCursor: next, hasMore: next != ""}
}

func ratePage(remaining int, reset time.Time, next string) pageSpec {
	p := okPage([]string{fmt.Sprintf("r%d", remaining)}, next)
	p.headers = map[string]string{
		"X-RateLimit-Limit":     "100",
		"X-RateLimit-Remaining": fmt.Sprintf("%d", remaining),
		"X-RateLimit-Reset":     fmt.Sprintf("%d", reset.Unix()),
	}
	return p
}

func tooManyPage(retryAfter string) pageSpec {
	return pageSpec{
		status:     http.StatusTooManyRequests,
		errCode:    api.CodeRateLimited,
		errMessage: "slow down",
		headers:    map[string]string{"Retry-After": retryAfter},
	}
}

// paginateHarness wires a cobra command tree, a RunContext, and an api.Client
// pointed at a scripted httptest server. Nothing here touches the network.
type paginateHarness struct {
	cmd    *cobra.Command
	client *api.Client
	api    *fakeAPI
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	slept  []time.Duration
}

func newPaginateHarness(t *testing.T, outFmt string, pages []pageSpec, fallthroughPage *pageSpec, args ...string) *paginateHarness {
	t.Helper()

	fa := &fakeAPI{t: t, pages: pages, fallth: fallthroughPage}
	srv := httptest.NewServer(fa)
	t.Cleanup(srv.Close)

	h := &paginateHarness{api: fa, stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}}

	root := &cobra.Command{Use: "kubeadapt"}
	root.PersistentFlags().BoolP("quiet", "q", false, "")
	get := &cobra.Command{Use: "get"}
	get.PersistentFlags().String(flagCostMode, "fully_loaded", "")
	get.PersistentFlags().String(flagCursor, "", "")
	get.PersistentFlags().Int(flagLimit, 5, "")
	get.PersistentFlags().Bool(flagPaginate, false, "")
	get.PersistentFlags().Bool(flagIncludeTotal, false, "")
	get.PersistentFlags().Duration(flagMaxWait, defaultMaxWait, "")

	leaf := &cobra.Command{Use: "items", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	get.AddCommand(leaf)
	root.AddCommand(get)
	root.SetOut(h.stdout)
	root.SetErr(h.stderr)
	root.SetArgs(append([]string{"get", "items"}, args...))
	require.NoError(t, root.Execute())

	isQuiet, err := root.PersistentFlags().GetBool("quiet")
	require.NoError(t, err)
	rc := &RunContext{
		Config:    &config.Config{APIURL: srv.URL, APIKey: "test-key"},
		Logger:    zap.NewNop(),
		OutputFmt: outFmt,
		Quiet:     isQuiet,
	}
	withRunContext(leaf, rc)
	leaf.SetOut(h.stdout)
	leaf.SetErr(h.stderr)

	h.cmd = leaf
	return h
}

// useFakeSleeper swaps the shared wait function for one that records instead of
// waiting, so budget and Retry-After arithmetic is asserted without real delay.
func (h *paginateHarness) useFakeSleeper(t *testing.T, fn func(context.Context, time.Duration) error) {
	t.Helper()
	prev := paginateSleep
	t.Cleanup(func() { paginateSleep = prev })
	paginateSleep = func(ctx context.Context, d time.Duration) error {
		h.slept = append(h.slept, d)
		if fn != nil {
			return fn(ctx, d)
		}
		return nil
	}
}

func (h *paginateHarness) run(t *testing.T) error {
	t.Helper()
	paged, err := parsePagedFlags(h.cmd)
	require.NoError(t, err)

	c, err := newAPIClientFromCmd(h.cmd)
	require.NoError(t, err)
	h.client = c

	fetch := func(ctx context.Context, cursor string) ([]pageItem, *types.Meta, error) {
		params := url.Values{}
		if cursor != "" {
			params.Set(flagCursor, cursor)
		}
		params.Set(flagLimit, strconv.Itoa(paged.Limit))
		return api.DoEnvelopeGet[[]pageItem](ctx, c, "/v1/items", params)
	}
	renderTable := func(w io.Writer, items []pageItem, _ *types.Meta) error {
		for _, it := range items {
			fmt.Fprintln(w, it.ID)
		}
		return nil
	}
	return runPaginatedList(h.cmd, c, paged, fetch, renderTable)
}

func (h *paginateHarness) emittedIDs(t *testing.T) []string {
	t.Helper()
	var envelope struct {
		Data []pageItem `json:"data"`
	}
	require.NoError(t, json.Unmarshal(h.stdout.Bytes(), &envelope), "stdout was not JSON: %s", h.stdout.String())
	ids := make([]string, 0, len(envelope.Data))
	for _, it := range envelope.Data {
		ids = append(ids, it.ID)
	}
	return ids
}

func TestPaginate_PartialResultsPreservedOn429(t *testing.T) {
	stuck := tooManyPage("60")
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1", "a2"}, "c1"),
		okPage([]string{"b1", "b2"}, "c2"),
		okPage([]string{"d1", "d2"}, "c3"),
	}, &stuck, "--paginate")
	h.useFakeSleeper(t, nil)

	err := h.run(t)

	require.Error(t, err, "a run cut short by 429 must surface the failure")
	assert.Equal(t, []string{"a1", "a2", "b1", "b2", "d1", "d2"}, h.emittedIDs(t),
		"every page fetched before the 429 must still reach stdout")
	assert.Contains(t, h.stdout.String(), `"partial": true`,
		"machine consumers need a truncation marker")
	assert.Contains(t, h.stdout.String(), `"resume_cursor": "c3"`,
		"the resume cursor must be machine-readable, not stderr-only")
	assert.Contains(t, h.stderr.String(), "--cursor=c3",
		"stderr must tell the user where to resume")
}

func TestPaginate_PacesBeforeExhaustingBudget(t *testing.T) {
	reset := time.Now().Add(2 * time.Second)
	pages := []pageSpec{}
	for remaining := 9; remaining >= 1; remaining-- {
		pages = append(pages, ratePage(remaining, reset, fmt.Sprintf("c%d", remaining)))
	}
	pages = append(pages, ratePage(0, reset, ""))

	h := newPaginateHarness(t, formatJSON, pages, nil, "--paginate")
	h.useFakeSleeper(t, nil)

	err := h.run(t)

	require.NoError(t, err, "pacing should let the run finish cleanly")
	require.NotEmpty(t, h.slept, "client must pause as the window nears exhaustion instead of blowing through it")
	assert.Equal(t, len(pages), h.api.callCount(), "every page should still be fetched")
}

func TestPaginate_RetryHonorsRetryAfter(t *testing.T) {
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1"}, "c1"),
		tooManyPage("2"),
		okPage([]string{"b1"}, ""),
	}, nil, "--paginate")
	h.useFakeSleeper(t, nil)

	err := h.run(t)

	require.NoError(t, err, "a transient 429 followed by success must not fail the run")
	assert.Equal(t, []string{"a1", "b1"}, h.emittedIDs(t))
	assert.Contains(t, h.slept, 2*time.Second,
		"retry must wait the server-supplied Retry-After, not a hardcoded default; slept=%v", h.slept)
}

func TestPaginate_MaxWaitBudgetExceededReturnsPartial(t *testing.T) {
	farReset := time.Now().Add(10 * time.Minute)
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1"}, "c1"),
		ratePage(0, farReset, "c2"),
	}, nil, "--paginate", "--max-wait=1s")
	h.useFakeSleeper(t, nil)

	err := h.run(t)

	require.Error(t, err, "exceeding the wait budget must not look like success")
	assert.Contains(t, err.Error(), "max-wait")
	assert.Equal(t, []string{"a1", "r0"}, h.emittedIDs(t), "pages fetched before the budget ran out must survive")
	assert.Contains(t, h.stdout.String(), `"partial": true`)
	assert.Contains(t, h.stderr.String(), "--cursor=c2")
	assert.Empty(t, h.slept, "the budget check must happen before sleeping, not after")
}

// HasMore is the authoritative end-of-results signal, so HasMore with no
// cursor is a server contract violation, not a terminal page. Returning it as a
// clean run handed the caller a silently truncated answer marked complete.
func TestPaginate_HasMoreWithoutCursorIsReportedAsPartial(t *testing.T) {
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1"}, "c1"),
		{items: []pageItem{{ID: "b1"}}, hasMore: true},
	}, nil, "--paginate")

	err := h.run(t)

	require.Error(t, err, "an inconsistent pagination block must not read as a complete run")
	assert.Equal(t, []string{"a1", "b1"}, h.emittedIDs(t), "everything fetched must still be returned")
	assert.Contains(t, h.stdout.String(), `"partial": true`,
		"the document must say the answer may be short")
}

// Under -o table the only truncation signal was stderr prose, which is
// routinely redirected away; a short table then looks like a complete answer.
func TestPaginate_TableModeSignalsPartialOnStdout(t *testing.T) {
	farReset := time.Now().Add(10 * time.Minute)
	h := newPaginateHarness(t, formatTable, []pageSpec{
		okPage([]string{"a1"}, "c1"),
		ratePage(0, farReset, "c2"),
	}, nil, "--paginate", "--max-wait=1s")
	h.useFakeSleeper(t, nil)

	err := h.runClusters(t)

	require.Error(t, err)
	assert.Contains(t, h.stdout.String(), "PARTIAL RESULTS",
		"a truncated table run must be unmissable on stdout, not only on stderr")
	assert.Contains(t, h.stdout.String(), "--cursor=c2",
		"stdout must carry the resume cursor for the truncated run")
}

func TestPaginate_TableModePartialSurvivesQuiet(t *testing.T) {
	output.SetQuiet(true)
	t.Cleanup(func() { output.SetQuiet(false) })

	farReset := time.Now().Add(10 * time.Minute)
	h := newPaginateHarness(t, formatTable, []pageSpec{
		okPage([]string{"a1"}, "c1"),
		ratePage(0, farReset, "c2"),
	}, nil, "--paginate", "--max-wait=1s", "--quiet")
	h.useFakeSleeper(t, nil)

	err := h.runClusters(t)

	require.Error(t, err)
	assert.Contains(t, h.stdout.String(), "PARTIAL RESULTS",
		"--quiet drops chrome; a truncated result set is not chrome")
	assert.NotContains(t, h.stdout.String(), "Showing",
		"precondition: --quiet still suppresses the descriptive footer")
}

func TestPaginate_CursorExpiredMidRunReturnsPartialNotRestart(t *testing.T) {
	expired := pageSpec{
		status:     http.StatusGone,
		errCode:    api.CodeCursorExpired,
		errMessage: "cursor expired",
	}
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1", "a2"}, "c1"),
		expired,
	}, &expired, "--paginate")
	h.useFakeSleeper(t, nil)

	err := h.run(t)

	require.Error(t, err)
	assert.True(t, api.IsCursorExpired(err), "expected CURSOR_EXPIRED to survive wrapping, got %v", err)
	assert.Equal(t, []string{"a1", "a2"}, h.emittedIDs(t), "expired cursor must not duplicate already-fetched rows")
	assert.Contains(t, h.stderr.String(), "--cursor=c1")

	firstPageFetches := 0
	for _, c := range h.api.seenCursors() {
		if c == "" {
			firstPageFetches++
		}
	}
	assert.Equal(t, 1, firstPageFetches, "page 1 must never be re-fetched; cursors=%v", h.api.seenCursors())
}

func TestPaginate_ContextCancelDuringSleepAborts(t *testing.T) {
	reset := time.Now().Add(30 * time.Second)
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1"}, "c1"),
		ratePage(0, reset, "c2"),
	}, nil, "--paginate")
	h.useFakeSleeper(t, func(ctx context.Context, _ time.Duration) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		<-cctx.Done()
		return cctx.Err()
	})

	err := h.run(t)

	require.Error(t, err, "a cancelled wait must abort the run")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, []string{"a1", "r0"}, h.emittedIDs(t), "Ctrl-C must not throw away completed pages")
	assert.Contains(t, h.stderr.String(), "--cursor=c2")
}

func ignoredPage(ids []string, next string, ignored ...string) pageSpec {
	p := okPage(ids, next)
	p.ignored = ignored
	return p
}

// runClusters drives the same paginator through the real table renderer so the
// pagination footer (which --quiet does suppress) is actually produced.
func (h *paginateHarness) runClusters(t *testing.T) error {
	t.Helper()
	paged, err := parsePagedFlags(h.cmd)
	require.NoError(t, err)

	c, err := newAPIClientFromCmd(h.cmd)
	require.NoError(t, err)
	h.client = c

	fetch := func(ctx context.Context, cursor string) ([]types.Cluster, *types.Meta, error) {
		params := url.Values{}
		if cursor != "" {
			params.Set(flagCursor, cursor)
		}
		params.Set(flagLimit, strconv.Itoa(paged.Limit))
		return api.DoEnvelopeGet[[]types.Cluster](ctx, c, "/v1/items", params)
	}
	return runPaginatedList(h.cmd, c, paged, fetch, output.RenderClusters)
}

func TestIgnoredParams_WarnsOnStderr(t *testing.T) {
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		ignoredPage([]string{"a1"}, "", "namespaces"),
	}, nil)
	h.useFakeSleeper(t, nil)

	require.NoError(t, h.run(t))

	assert.Contains(t, h.stderr.String(), "namespaces",
		"the operator must be told which parameter was dropped")
	assert.Contains(t, strings.ToLower(h.stderr.String()), "warning",
		"the line must read as a warning, not as incidental chatter")
	assert.NotContains(t, h.stdout.String(), "warning",
		"stdout must stay pipeable into jq; the warning belongs on stderr")
	assert.Equal(t, []string{"a1"}, h.emittedIDs(t),
		"stdout must remain clean, parseable JSON")
}

func TestIgnoredParams_WarnsOncePerRunNotPerPage(t *testing.T) {
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		ignoredPage([]string{"a1"}, "c1", "namespaces"),
		ignoredPage([]string{"b1"}, "c2", "namespaces"),
		ignoredPage([]string{"d1"}, "c3", "namespaces"),
		ignoredPage([]string{"e1"}, "", "namespaces"),
	}, nil, "--paginate")
	h.useFakeSleeper(t, nil)

	require.NoError(t, h.run(t))

	assert.Equal(t, 1, strings.Count(h.stderr.String(), "namespaces"),
		"every page carries the same ignored_params; repeating the warning per page is its own defect. stderr=%q", h.stderr.String())
}

func TestIgnoredParams_SilentWhenAbsent(t *testing.T) {
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1"}, ""),
	}, nil)
	h.useFakeSleeper(t, nil)

	require.NoError(t, h.run(t))

	assert.NotContains(t, strings.ToLower(h.stderr.String()), "warning",
		"a request with no ignored params must not warn")
}

func TestIgnoredParams_NotSuppressedByQuiet(t *testing.T) {
	output.SetQuiet(true)
	t.Cleanup(func() { output.SetQuiet(false) })

	h := newPaginateHarness(t, formatTable, []pageSpec{
		ignoredPage([]string{"a1"}, "", "namespaces"),
	}, nil, "--quiet")
	h.useFakeSleeper(t, nil)

	require.NoError(t, h.runClusters(t))

	assert.Contains(t, h.stderr.String(), "namespaces",
		"--quiet drops non-essential chrome; wrong numbers are not chrome")
	assert.NotContains(t, h.stdout.String(), "Showing",
		"precondition: --quiet must still suppress the pagination footer")
}

func TestIgnoredParams_LiteralWarningText(t *testing.T) {
	tests := []struct {
		name    string
		ignored []string
		want    string
	}{
		{
			name:    "single",
			ignored: []string{"namespaces"},
			want:    "kubeadapt: warning: the API ignored unknown parameter \"namespaces\" - results are NOT filtered by it.\n",
		},
		{
			name:    "multiple",
			ignored: []string{"namespaces", "clusterId", "min_cost"},
			want:    "kubeadapt: warning: the API ignored unknown parameters \"namespaces\", \"clusterId\", \"min_cost\" - results are NOT filtered by them.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newPaginateHarness(t, formatJSON, []pageSpec{
				ignoredPage([]string{"a1"}, "", tt.ignored...),
			}, nil)
			h.useFakeSleeper(t, nil)

			require.NoError(t, h.run(t))
			assert.Equal(t, tt.want, h.stderr.String())
		})
	}
}

func TestPaginate_WarnsOnSmallLimitWithPaginate(t *testing.T) {
	h := newPaginateHarness(t, formatJSON, []pageSpec{
		okPage([]string{"a1"}, ""),
	}, nil, "--paginate", "--limit=5")
	h.useFakeSleeper(t, nil)

	require.NoError(t, h.run(t))
	assert.Contains(t, strings.ToLower(h.stderr.String()), "--limit",
		"a tiny page size under --paginate should suggest a larger one")
}
