package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type widget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(body)
	require.NoError(t, err, "encode response")
}

func successEnvelope(data widget) types.Envelope[widget] {
	return types.Envelope[widget]{
		Data: data,
		Meta: types.Meta{
			RequestID: "req-abc-123",
			AppliedAt: "2026-01-01T00:00:00Z",
		},
	}
}

func TestDoEnvelopeGet_HappyPath(t *testing.T) {
	want := widget{ID: "w-1", Name: "alpha"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/widgets/w-1", r.URL.Path)
		writeJSON(t, w, http.StatusOK, successEnvelope(want))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	got, meta, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets/w-1", nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	require.NotNil(t, meta)
	assert.Equal(t, "req-abc-123", meta.RequestID)
}

func TestDoEnvelopeGet_EnvelopeError_422(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusUnprocessableEntity, types.Envelope[widget]{
			Error: &types.APIErrorBody{
				Code:    string(CodeInvalidCostMode),
				Message: "this endpoint does not accept cost_mode",
				Details: []map[string]any{{"field": "cost_mode"}},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.Error(t, err)
	assert.True(t, IsInvalidCostMode(err), "expected IsInvalidCostMode, got: %v", err)
	apiErr, ok := errors.AsType[*APIError](err)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.StatusCode)
	assert.Equal(t, CodeInvalidCostMode, apiErr.Code)
	assert.NotEmpty(t, apiErr.Details, "expected details to be populated")
}

func TestDoEnvelopeGet_RateLimited429_NoRetry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "60")
		writeJSON(t, w, http.StatusTooManyRequests, types.Envelope[widget]{
			Error: &types.APIErrorBody{
				Code:    string(CodeRateLimited),
				Message: "slow down",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.Error(t, err)
	assert.True(t, IsRateLimited(err), "expected IsRateLimited, got: %v", err)
	apiErr, _ := errors.AsType[*APIError](err)
	assert.Equal(t, 60*time.Second, apiErr.RetryAfter)
	assert.Equal(t, int32(1), calls.Load())
}

func TestDoEnvelopeGet_RateLimited429_WithRetry(t *testing.T) {
	var calls atomic.Int32
	want := widget{ID: "w-2", Name: "beta"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			writeJSON(t, w, http.StatusTooManyRequests, types.Envelope[widget]{
				Error: &types.APIErrorBody{
					Code:    string(CodeRateLimited),
					Message: "slow down",
				},
			})
			return
		}
		writeJSON(t, w, http.StatusOK, successEnvelope(want))
	}))
	defer srv.Close()

	var slept []time.Duration
	c := NewClient(srv.URL, "key",
		WithRetryOnRateLimit(true),
		WithMaxRetries(1),
		withSleeper(func(d time.Duration) { slept = append(slept, d) }),
	)
	got, meta, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.NotNil(t, meta, "expected non-nil meta")
	assert.Equal(t, int32(2), calls.Load())
	require.Len(t, slept, 1)
	assert.Equal(t, time.Second, slept[0])
}

func TestDoEnvelopeGet_RateLimited429_RetryGivesUp(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "1")
		writeJSON(t, w, http.StatusTooManyRequests, types.Envelope[widget]{
			Error: &types.APIErrorBody{
				Code:    string(CodeRateLimited),
				Message: "slow down",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key",
		WithRetryOnRateLimit(true),
		WithMaxRetries(1),
		withSleeper(func(time.Duration) {}),
	)
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	assert.True(t, IsRateLimited(err), "expected IsRateLimited, got: %v", err)
	assert.Equal(t, int32(2), calls.Load(), "expected 2 calls (initial + 1 retry)")
}

func TestDoEnvelopeGet_RateLimitUnavailable503(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		writeJSON(t, w, http.StatusServiceUnavailable, types.Envelope[widget]{
			Error: &types.APIErrorBody{
				Code:    string(CodeRateLimitUnavailable),
				Message: "rate limiter unavailable",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key",
		WithRetryOnRateLimit(true),
		WithMaxRetries(3),
		withSleeper(func(time.Duration) {}),
	)
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	assert.True(t, IsRateLimitUnavailable(err), "expected IsRateLimitUnavailable, got: %v", err)
	assert.Equal(t, int32(1), calls.Load(), "503 must NOT trigger retry")
}

func TestDoEnvelopeGet_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		require.True(t, ok, "server does not support hijacking")
		conn, _, err := hj.Hijack()
		require.NoError(t, err, "hijack")
		_ = conn.Close()
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.Error(t, err, "expected transport error")
	_, ok := errors.AsType[*APIError](err)
	assert.False(t, ok, "transport error should not be *APIError, got %T: %v", err, err)
}

func TestDoEnvelopeGet_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.Error(t, err, "expected decode error")
	assert.Contains(t, err.Error(), "decoding response")
}

func TestDoEnvelopeGet_CursorParamsForwarded(t *testing.T) {
	var seen url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query()
		writeJSON(t, w, http.StatusOK, successEnvelope(widget{ID: "x"}))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	params := appendCursorParams(nil, "abc123", 50, true)
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", params)
	require.NoError(t, err)
	assert.Equal(t, "abc123", seen.Get("cursor"))
	assert.Equal(t, "50", seen.Get("limit"))
	assert.Equal(t, "true", seen.Get("include_total"))
}

func TestDoEnvelopeGet_AuthorizationHeader(t *testing.T) {
	var seenAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		writeJSON(t, w, http.StatusOK, successEnvelope(widget{ID: "x"}))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret-key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-key", seenAuth)
}

func TestDoEnvelopeGet_RateLimitHeadersCaptured(t *testing.T) {
	resetAt := time.Now().Add(2 * time.Minute).Unix()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "73")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))
		writeJSON(t, w, http.StatusOK, successEnvelope(widget{ID: "x"}))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	rl := c.RateLimit()
	assert.Equal(t, 100, rl.Limit)
	assert.Equal(t, 73, rl.Remaining)
	assert.Equal(t, resetAt, rl.Reset.Unix())
}

func TestDoEnvelopeGet_RequestIDLogged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "rid-server-issued-42")
		writeJSON(t, w, http.StatusOK, types.Envelope[widget]{
			Data: widget{ID: "x"},
			Meta: types.Meta{RequestID: "rid-server-issued-42"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, meta, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	assert.Equal(t, "rid-server-issued-42", meta.RequestID)
}

func TestDoEnvelopeGet_Non2xxWithoutEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("upstream is on fire"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	apiErr, ok := errors.AsType[*APIError](err)
	require.True(t, ok, "expected *APIError, got %T: %v", err, err)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "upstream is on fire")
}

func TestPickScopedOrFlat(t *testing.T) {
	scoped := func(id string) string { return "/v1/clusters/" + id + "/workloads" }
	flat := "/v1/workloads"

	tests := []struct {
		name      string
		clusterID []string
		wantPath  string
		wantCSV   string
	}{
		{"single cluster -> scoped", []string{"cls-1"}, "/v1/clusters/cls-1/workloads", ""},
		{"empty list -> flat", nil, flat, ""},
		{"empty strings only -> flat", []string{"", ""}, flat, ""},
		{"multiple -> flat with csv", []string{"cls-1", "cls-2"}, flat, "cls-1,cls-2"},
		{"single with empties trimmed -> scoped", []string{"", "cls-1", ""}, "/v1/clusters/cls-1/workloads", ""},
		{"three -> flat csv", []string{"a", "b", "c"}, flat, "a,b,c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, csv := pickScopedOrFlat(scoped, flat, tt.clusterID)
			assert.Equal(t, tt.wantPath, path, "path")
			assert.Equal(t, tt.wantCSV, csv, "csv")
		})
	}
}

func TestAppendCursorParams(t *testing.T) {
	tests := []struct {
		name         string
		seed         url.Values
		cursor       string
		limit        int
		includeTotal bool
		assertions   func(t *testing.T, p url.Values)
	}{
		{
			name: "nil params allocated",
			assertions: func(t *testing.T, p url.Values) {
				require.NotNil(t, p, "expected non-nil params")
				assert.Empty(t, p)
			},
		},
		{
			name:   "cursor set when non-empty",
			cursor: "abc",
			assertions: func(t *testing.T, p url.Values) {
				assert.Equal(t, "abc", p.Get("cursor"))
			},
		},
		{
			name:   "cursor omitted when empty",
			cursor: "",
			assertions: func(t *testing.T, p url.Values) {
				assert.NotContains(t, p, "cursor", "cursor should be absent")
			},
		},
		{
			name:  "limit zero omitted",
			limit: 0,
			assertions: func(t *testing.T, p url.Values) {
				assert.NotContains(t, p, "limit", "limit should be absent when 0")
			},
		},
		{
			name:  "limit positive set",
			limit: 25,
			assertions: func(t *testing.T, p url.Values) {
				assert.Equal(t, "25", p.Get("limit"))
			},
		},
		{
			name:  "limit negative omitted",
			limit: -5,
			assertions: func(t *testing.T, p url.Values) {
				assert.NotContains(t, p, "limit", "limit should be absent when negative")
			},
		},
		{
			name:         "include_total true",
			includeTotal: true,
			assertions: func(t *testing.T, p url.Values) {
				assert.Equal(t, "true", p.Get("include_total"))
			},
		},
		{
			name:         "include_total false omitted",
			includeTotal: false,
			assertions: func(t *testing.T, p url.Values) {
				assert.NotContains(t, p, "include_total", "include_total should be absent when false")
			},
		},
		{
			name:         "existing params preserved",
			seed:         url.Values{"existing": []string{"keep"}},
			cursor:       "c1",
			limit:        10,
			includeTotal: true,
			assertions: func(t *testing.T, p url.Values) {
				assert.Equal(t, "keep", p.Get("existing"), "existing param dropped")
				assert.Equal(t, "c1", p.Get("cursor"))
				assert.Equal(t, "10", p.Get("limit"))
				assert.Equal(t, "true", p.Get("include_total"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendCursorParams(tt.seed, tt.cursor, tt.limit, tt.includeTotal)
			tt.assertions(t, got)
		})
	}
}

func TestValidateNoCostMode(t *testing.T) {
	t.Run("absent params -> ok", func(t *testing.T) {
		assert.NoError(t, validateNoCostMode(nil, "/v1/clusters"))
		assert.NoError(t, validateNoCostMode(url.Values{}, "/v1/nodes"))
	})
	t.Run("other params ignored", func(t *testing.T) {
		p := url.Values{"limit": []string{"10"}, "cursor": []string{"x"}}
		assert.NoError(t, validateNoCostMode(p, "/v1/node-groups"))
	})
	endpointCases := []string{
		"/v1/clusters",
		"/v1/nodes",
		"/v1/recommendations",
		"/v1/organization",
	}
	for _, endpoint := range endpointCases {
		t.Run("cost_mode rejected on "+endpoint, func(t *testing.T) {
			p := url.Values{"cost_mode": []string{"fully_loaded"}}
			err := validateNoCostMode(p, endpoint)
			require.True(t, IsInvalidCostMode(err), "expected IsInvalidCostMode, got %v", err)
			apiErr, _ := errors.AsType[*APIError](err)
			assert.Equal(t, http.StatusUnprocessableEntity, apiErr.StatusCode)
			assert.Contains(t, apiErr.Message, endpoint)
		})
	}
}

func TestRateLimitSnapshot_MalformedHeadersZeroed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "not-a-number")
		w.Header().Set("X-RateLimit-Remaining", "")
		w.Header().Set("X-RateLimit-Reset", "garbage")
		writeJSON(t, w, http.StatusOK, successEnvelope(widget{ID: "x"}))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	rl := c.RateLimit()
	assert.Zero(t, rl.Limit)
	assert.Zero(t, rl.Remaining)
	assert.True(t, rl.Reset.IsZero(), "expected zeroed Reset, got %v", rl.Reset)
}

func TestRateLimitSnapshot_HTTPDateReset(t *testing.T) {
	resetTime := time.Now().Add(90 * time.Second).UTC().Truncate(time.Second)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Reset", resetTime.Format(http.TimeFormat))
		writeJSON(t, w, http.StatusOK, successEnvelope(widget{ID: "x"}))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	got := c.RateLimit().Reset
	assert.True(t, got.Equal(resetTime), "expected reset %s, got %s", resetTime, got)
}

func TestDoEnvelopeGet_UserAgent(t *testing.T) {
	var ua string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		writeJSON(t, w, http.StatusOK, successEnvelope(widget{ID: "x"}))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key")
	_, _, err := DoEnvelopeGet[widget](t.Context(), c, "/v1/widgets", nil)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(ua, "kubeadapt-cli/"), "expected User-Agent to start with 'kubeadapt-cli/', got %q", ua)
}
