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
	"github.com/kubeadapt/kubeadapt-cli/internal/version"
)

const (
	defaultTimeout = 30 * time.Second

	errorBodyLimit = 200
)

// The zero Client is not usable; construct it via NewClient.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
	retryOnRL  bool
	maxRetries int
	sleeper    func(context.Context, time.Duration) error
	rateLimit  *rateLimitSnapshot
}

type Option func(*Client)

// Mutates the underlying http.Client, so pair with WithHTTPClient when a custom
// transport is needed.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
	}
}

// Retries respect Retry-After. Defaults to false.
func WithRetryOnRateLimit(enable bool) Option {
	return func(c *Client) {
		c.retryOnRL = enable
	}
}

// Negative values clamp to zero. Defaults to 1.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

// Callers that also pace against rate-limit headers share one sleeper, so a
// single wait budget covers both kinds of wait.
func WithSleeper(fn func(context.Context, time.Duration) error) Option {
	return func(c *Client) {
		if fn != nil {
			c.sleeper = fn
		}
	}
}

func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		logger:     zap.NewNop(),
		retryOnRL:  false,
		maxRetries: 1,
		sleeper:    SleepContext,
		rateLimit:  &rateLimitSnapshot{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// The returned snapshot is a copy and safe to retain.
func (c *Client) RateLimit() RateLimit {
	return c.rateLimit.load()
}

// Returns the unwrapped Data payload and Meta. API-level failures come back as
// *APIError (RetryAfter set on 429); transport and decode failures come back as
// a plain wrapped error.
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
		req.Header.Set("User-Agent", buildUserAgent())

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
			zap.String("request_id", resolveRequestID(resp.Header, body)),
		)

		if resp.StatusCode == http.StatusTooManyRequests && c.retryOnRL && attempt < c.maxRetries {
			d := ParseRetryAfter(resp.Header.Get("Retry-After"))
			if d <= 0 {
				d = time.Second
			}
			if sleepErr := c.sleeper(ctx, d); sleepErr != nil {
				return zero, nil, sleepErr
			}
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

// A nil params map is allocated lazily, so callers may pass nil. limit<=0 omits
// the limit param.
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

// Returns the path plus the CSV cluster_id query value. On the scoped path the
// cluster ID is consumed by the URL itself, so csvParam comes back empty.
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

// Named once so the CLI's pre-request refusal and the server's 422 cannot drift
// onto different endpoint names.
const (
	EndpointClusters        = "clusters"
	EndpointCluster         = "cluster"
	EndpointNodes           = "nodes"
	EndpointNode            = "node"
	EndpointNodeGroups      = "node-groups"
	EndpointNodeGroup       = "node-group"
	EndpointRecommendations = "recommendations"
	EndpointRecommendation  = "recommendation"
	EndpointOrganization    = "organization"
	EndpointTeamAssignments = "team-assignments"
)

// Names the --cost-mode flag, not the cost_mode param: it is raised before any
// request, so the user's own input is all they can act on.
type CostModeUnsupportedError struct {
	Endpoint string
}

func (e *CostModeUnsupportedError) Error() string {
	return "--cost-mode is not accepted by the " + e.Endpoint + " endpoint"
}

// Called before a request is built, so the refusal costs no round-trip.
func RejectCostMode(endpoint, costMode string) error {
	if costMode == "" {
		return nil
	}
	return &CostModeUnsupportedError{Endpoint: endpoint}
}

// Wire-shaped form of RejectCostMode, mirroring the server's 422.
func validateNoCostMode(params url.Values, endpoint string) error {
	rejection := RejectCostMode(endpoint, params.Get("cost_mode"))
	if rejection == nil {
		return nil
	}
	return &APIError{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       CodeInvalidCostMode,
		Message:    rejection.Error(),
		Details: []map[string]any{{
			"field":   "cost_mode",
			"allowed": []string{},
		}},
	}
}

func buildUserAgent() string {
	return "kubeadapt-cli/" + version.Version
}

// The API returns the correlation id in the envelope body, not a header, so a
// header-only lookup always came back empty. Parsing just this one field keeps
// malformed or non-envelope responses logging cleanly instead of erroring.
func resolveRequestID(h http.Header, body []byte) string {
	if id := h.Get("X-Request-ID"); id != "" {
		return id
	}
	var envelope struct {
		Meta struct {
			RequestID string `json:"request_id"`
		} `json:"meta"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return ""
	}
	return envelope.Meta.RequestID
}
