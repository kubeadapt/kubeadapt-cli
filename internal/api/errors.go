// Package api contains the HTTP client and error types for the Kubeadapt
// public API.
package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ErrorCode is a stable string identifier returned by the Kubeadapt public API
// in the response envelope's "error.code" field. It is intentionally separated
// from the HTTP status code because (a) some codes share a status (404 covers
// all *_NOT_FOUND codes) and (b) tooling should switch on the code, not the
// status, for stable behavior.
type ErrorCode string

// Error code constants. Keep this list in sync with the Kubeadapt public API
// error code spec.
const (
	CodeUnauthorized           ErrorCode = "UNAUTHORIZED"
	CodeForbidden              ErrorCode = "FORBIDDEN"
	CodeClusterAccessDenied    ErrorCode = "CLUSTER_ACCESS_DENIED"
	CodeRateLimited            ErrorCode = "RATE_LIMITED"
	CodeRateLimitUnavailable   ErrorCode = "RATE_LIMIT_UNAVAILABLE"
	CodeBadRequest             ErrorCode = "BAD_REQUEST"
	CodeValidationError        ErrorCode = "VALIDATION_ERROR"
	CodeInvalidCursor          ErrorCode = "INVALID_CURSOR"
	CodeCursorExpired          ErrorCode = "CURSOR_EXPIRED"
	CodeInvalidCostMode        ErrorCode = "INVALID_COST_MODE"
	CodeInvalidClusterID       ErrorCode = "INVALID_CLUSTER_ID"
	CodeClusterNotFound        ErrorCode = "CLUSTER_NOT_FOUND"
	CodeNamespaceNotFound      ErrorCode = "NAMESPACE_NOT_FOUND"
	CodeWorkloadNotFound       ErrorCode = "WORKLOAD_NOT_FOUND"
	CodeNodeNotFound           ErrorCode = "NODE_NOT_FOUND"
	CodeNodeGroupNotFound      ErrorCode = "NODE_GROUP_NOT_FOUND"
	CodeTeamNotFound           ErrorCode = "TEAM_NOT_FOUND"
	CodeDepartmentNotFound     ErrorCode = "DEPARTMENT_NOT_FOUND"
	CodeRecommendationNotFound ErrorCode = "RECOMMENDATION_NOT_FOUND"
	CodeInternalError          ErrorCode = "INTERNAL_ERROR"
	CodeServiceUnavailable     ErrorCode = "SERVICE_UNAVAILABLE"
)

// APIError is the error type returned by HTTP client methods when the API
// responds with a non-2xx status or a populated error envelope.
//
// It carries both the HTTP status code (for callers who want to dispatch on
// the transport layer) and the structured error envelope fields. RetryAfter
// is populated from the Retry-After response header when the server signals
// rate limiting or temporary unavailability; it is the zero Duration when
// absent.
type APIError struct { //nolint:revive
	StatusCode int              `json:"status_code"`
	Code       ErrorCode        `json:"code"`
	Message    string           `json:"message"`
	Details    []map[string]any `json:"details,omitempty"`
	RetryAfter time.Duration    `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil APIError>"
	}
	if e.Code != "" {
		return fmt.Sprintf("API error (HTTP %d, code=%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsCode reports whether the wrapped error is an APIError with the given code.
// It is the preferred way to dispatch on error code from caller code.
func IsCode(err error, code ErrorCode) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Code == code
}

// IsUnauthorized reports whether the error is an UNAUTHORIZED (HTTP 401) error.
func IsUnauthorized(err error) bool { return IsCode(err, CodeUnauthorized) }

// IsForbidden reports whether the error is a FORBIDDEN (HTTP 403) error
// triggered by missing scopes.
func IsForbidden(err error) bool { return IsCode(err, CodeForbidden) }

// IsClusterAccessDenied reports whether the error is CLUSTER_ACCESS_DENIED
// (HTTP 403) triggered by an API key restricted to a different set of clusters.
func IsClusterAccessDenied(err error) bool { return IsCode(err, CodeClusterAccessDenied) }

// IsRateLimited reports whether the error is RATE_LIMITED (HTTP 429). Callers
// implementing automatic retry should also check the APIError.RetryAfter
// field to honor the Retry-After header.
func IsRateLimited(err error) bool { return IsCode(err, CodeRateLimited) }

// IsRateLimitUnavailable reports whether the error is RATE_LIMIT_UNAVAILABLE
// (HTTP 503) — the server's rate limiter (Valkey) is down and the server is
// fail-closed.
func IsRateLimitUnavailable(err error) bool { return IsCode(err, CodeRateLimitUnavailable) }

// IsCursorExpired reports whether the pagination cursor has expired (HTTP 410)
// or was issued against a different query (query_hash mismatch).
func IsCursorExpired(err error) bool { return IsCode(err, CodeCursorExpired) }

// IsInvalidCursor reports whether the supplied cursor was malformed (HTTP 400).
func IsInvalidCursor(err error) bool { return IsCode(err, CodeInvalidCursor) }

// IsInvalidCostMode reports whether the caller passed cost_mode= to an endpoint
// that does not accept it (HTTP 422).
func IsInvalidCostMode(err error) bool { return IsCode(err, CodeInvalidCostMode) }

// IsValidationError reports whether the request failed schema validation
// (HTTP 422) for reasons other than cost_mode.
func IsValidationError(err error) bool { return IsCode(err, CodeValidationError) }

// IsBadRequest reports whether the request was malformed (HTTP 400).
func IsBadRequest(err error) bool { return IsCode(err, CodeBadRequest) }

// IsNotFound reports whether the error indicates a missing resource — that is,
// any of the *_NOT_FOUND error codes returned by the API at HTTP 404.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.Code {
	case CodeClusterNotFound, CodeNamespaceNotFound, CodeWorkloadNotFound,
		CodeNodeNotFound, CodeNodeGroupNotFound, CodeTeamNotFound,
		CodeDepartmentNotFound, CodeRecommendationNotFound:
		return true
	}
	// Fallback for non-code 404s (e.g., transport-layer 404s before the body
	// is parsed).
	return apiErr.StatusCode == http.StatusNotFound
}

// IsServerError reports whether the error is a 5xx response (HTTP 500–599).
func IsServerError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode >= 500 && apiErr.StatusCode <= 599
}

// IsAuthError is a backward-compatibility alias for IsUnauthorized; new code
// should use the package-level IsUnauthorized helper.
func (e *APIError) IsAuthError() bool {
	if e == nil {
		return false
	}
	return e.Code == CodeUnauthorized || e.StatusCode == http.StatusUnauthorized
}

// IsForbidden is a backward-compatibility method alias; new code should use
// the package-level IsForbidden helper.
func (e *APIError) IsForbidden() bool {
	if e == nil {
		return false
	}
	return e.Code == CodeForbidden || e.Code == CodeClusterAccessDenied ||
		e.StatusCode == http.StatusForbidden
}

// IsNotFound is a backward-compatibility method alias; new code should use
// the package-level IsNotFound helper.
func (e *APIError) IsNotFound() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusNotFound || IsNotFound(e)
}

// IsRateLimited is a backward-compatibility method alias; new code should use
// the package-level IsRateLimited helper.
func (e *APIError) IsRateLimited() bool {
	if e == nil {
		return false
	}
	return e.Code == CodeRateLimited || e.StatusCode == http.StatusTooManyRequests
}

// IsServerError is a backward-compatibility method alias; new code should use
// the package-level IsServerError helper.
func (e *APIError) IsServerError() bool {
	if e == nil {
		return false
	}
	return e.StatusCode >= 500 && e.StatusCode <= 599
}

// ParseRetryAfter parses the Retry-After header value as a duration. It accepts
// the two forms allowed by RFC 7231 §7.1.3: a non-negative integer number of
// seconds, or an HTTP-date. If the header is empty or malformed the returned
// duration is zero.
func ParseRetryAfter(headerValue string) time.Duration {
	if headerValue == "" {
		return 0
	}
	if secs, err := strconv.Atoi(headerValue); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(headerValue); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}
