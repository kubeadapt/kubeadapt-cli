package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

const (
	defaultTimeout = 30 * time.Second
	userAgent      = "kubeadapt-cli/dev"
	errorBodyLimit = 200
)

// Client is the Kubeadapt public API HTTP client. It speaks the envelope
// response protocol, captures rate-limit headers from every response, and
// optionally retries once on HTTP 429. The zero value is not usable; construct
// it via NewClient.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
	retryOnRL  bool
	maxRetries int
	sleeper    func(time.Duration)
	rateLimit  *rateLimitSnapshot
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout. It mutates the underlying
// http.Client; pair with WithHTTPClient if a custom transport is needed.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithLogger sets the debug logger.
func WithLogger(l *zap.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
	}
}

// WithRetryOnRateLimit enables or disables an automatic single retry after a
// 429 response. The retry respects Retry-After. Defaults to false.
func WithRetryOnRateLimit(enable bool) Option {
	return func(c *Client) {
		c.retryOnRL = enable
	}
}

// WithMaxRetries sets the maximum number of retries the client will issue
// after a 429 response. Values below zero are clamped to zero. Defaults to 1.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

// withSleeper overrides the sleep function used between retries. It is
// unexported and intended for tests in the same package.
func withSleeper(fn func(time.Duration)) Option {
	return func(c *Client) {
		if fn != nil {
			c.sleeper = fn
		}
	}
}

// NewClient creates a new API client.
func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		logger:     zap.NewNop(),
		retryOnRL:  false,
		maxRetries: 1,
		sleeper:    time.Sleep,
		rateLimit:  &rateLimitSnapshot{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// RateLimit returns the most recent rate-limit snapshot captured from API
// responses. The returned value is a copy and safe to retain.
func (c *Client) RateLimit() RateLimit {
	return c.rateLimit.load()
}

// DoEnvelopeGet issues a GET request against the API and decodes the envelope
// response. It returns the unwrapped Data payload, a pointer to the Meta
// block, and an error.
//
// Errors fall into three buckets:
//   - Transport / decode failures: returned as a plain error wrapping the
//     underlying cause. Callers can check via errors.Is on context errors.
//   - API-level errors (envelope.Error populated, or non-2xx status): returned
//     as *APIError with StatusCode, Code, Message, and Details. On a 429 the
//     RetryAfter field is populated from the Retry-After header.
//   - Success: nil error, populated data and meta.
//
// When WithRetryOnRateLimit(true) is set, a 429 response triggers up to
// maxRetries additional attempts after sleeping for the Retry-After delay
// (defaulting to one second if absent or malformed).
func DoEnvelopeGet[T any](ctx context.Context, c *Client, path string, params url.Values) (T, *types.Meta, error) {
	var zero T

	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL = fullURL + "?" + params.Encode()
	}

	attempt := 0
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return zero, nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)

		start := time.Now()
		c.logger.Debug("api request",
			zap.String("method", http.MethodGet),
			zap.String("url", fullURL),
			zap.Int("attempt", attempt),
		)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return zero, nil, fmt.Errorf("executing request: %w", err)
		}

		c.rateLimit.captureFromHeaders(resp.Header)

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return zero, nil, fmt.Errorf("reading response: %w", readErr)
		}

		c.logger.Debug("api response",
			zap.String("url", fullURL),
			zap.Int("status", resp.StatusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", resp.Header.Get("X-Request-ID")),
		)

		if resp.StatusCode == http.StatusTooManyRequests && c.retryOnRL && attempt < c.maxRetries {
			d := ParseRetryAfter(resp.Header.Get("Retry-After"))
			if d <= 0 {
				d = time.Second
			}
			c.sleeper(d)
			attempt++
			continue
		}

		var env types.Envelope[T]
		decodeErr := json.Unmarshal(body, &env)

		if env.Error != nil {
			apiErr := &APIError{
				StatusCode: resp.StatusCode,
				Code:       ErrorCode(env.Error.Code),
				Message:    env.Error.Message,
				Details:    env.Error.Details,
			}
			if apiErr.Code == CodeRateLimited {
				apiErr.RetryAfter = ParseRetryAfter(resp.Header.Get("Retry-After"))
			}
			return zero, nil, apiErr
		}

		if resp.StatusCode >= 400 {
			excerpt := string(body)
			if len(excerpt) > errorBodyLimit {
				excerpt = excerpt[:errorBodyLimit]
			}
			apiErr := &APIError{
				StatusCode: resp.StatusCode,
				Message:    excerpt,
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				apiErr.Code = CodeRateLimited
				apiErr.RetryAfter = ParseRetryAfter(resp.Header.Get("Retry-After"))
			}
			return zero, nil, apiErr
		}

		if decodeErr != nil {
			return zero, nil, fmt.Errorf("decoding response: %w", decodeErr)
		}

		return env.Data, &env.Meta, nil
	}
}

// appendCursorParams adds cursor, limit, and include_total to params using the
// standard query keys understood by the Kubeadapt public API. A nil params
// map is allocated lazily so callers can write
// p := appendCursorParams(nil, ...). limit<=0 omits the limit param.
func appendCursorParams(params url.Values, cursor string, limit int, includeTotal bool) url.Values {
	if params == nil {
		params = url.Values{}
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if includeTotal {
		params.Set("include_total", "true")
	}
	return params
}

// pickScopedOrFlat chooses between the cluster-scoped path (when a single
// cluster ID is supplied) and the flat path (otherwise). It returns the path
// to call plus the comma-separated cluster_id value to attach as a query
// param. When using the scoped path the cluster_id is consumed by the URL
// itself and the returned csvParam is empty.
func pickScopedOrFlat(
	scopedPathFn func(clusterID string) string,
	flatPath string,
	clusterIDs []string,
) (path, csvParam string) {
	nonEmpty := make([]string, 0, len(clusterIDs))
	for _, id := range clusterIDs {
		if id != "" {
			nonEmpty = append(nonEmpty, id)
		}
	}
	if len(nonEmpty) == 1 {
		return scopedPathFn(nonEmpty[0]), ""
	}
	return flatPath, strings.Join(nonEmpty, ",")
}

// validateNoCostMode returns an *APIError with code INVALID_COST_MODE if the
// caller has set the cost_mode query param. Per-resource methods that call
// endpoints which reject cost_mode (cluster, node, node-group, recommendation,
// organization root) should invoke this BEFORE sending the request so that the
// CLI rejects locally — fast feedback, no wasted network round-trip.
func validateNoCostMode(params url.Values, endpoint string) error {
	if params == nil {
		return nil
	}
	if v := params.Get("cost_mode"); v != "" {
		return &APIError{
			StatusCode: http.StatusUnprocessableEntity,
			Code:       CodeInvalidCostMode,
			Message:    fmt.Sprintf("%s does not accept cost_mode", endpoint),
			Details: []map[string]any{{
				"field":   "cost_mode",
				"allowed": []string{},
			}},
		}
	}
	return nil
}
