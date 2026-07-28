package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ErrorCode is deliberately separate from the HTTP status: several codes share
// one status (404 covers every *_NOT_FOUND), so dispatch on the code.
type ErrorCode string

// Keep in sync with the Kubeadapt public API error code spec.
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
	CodeNotFound               ErrorCode = "NOT_FOUND"
)

// APIError is returned on any non-2xx status or populated error envelope.
// RetryAfter comes from the Retry-After header and is zero when absent.
type APIError struct { //nolint:revive
	StatusCode int              `json:"status_code"`
	Code       ErrorCode        `json:"code"`
	Message    string           `json:"message"`
	Details    []map[string]any `json:"details,omitempty"`
	RetryAfter time.Duration    `json:"-"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil APIError>"
	}
	if e.Code != "" {
		return fmt.Sprintf("API error (HTTP %d, code=%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsCode is the preferred way to dispatch on an API error code.
func IsCode(err error, code ErrorCode) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Code == code
}

func IsUnauthorized(err error) bool { return IsCode(err, CodeUnauthorized) }

// IsForbidden is triggered by missing scopes.
func IsForbidden(err error) bool { return IsCode(err, CodeForbidden) }

// IsClusterAccessDenied means the API key is restricted to other clusters.
func IsClusterAccessDenied(err error) bool { return IsCode(err, CodeClusterAccessDenied) }

// Retrying callers should also honor APIError.RetryAfter.
func IsRateLimited(err error) bool { return IsCode(err, CodeRateLimited) }

// IsRateLimitUnavailable means the server's rate limiter is down and the server
// is fail-closed.
func IsRateLimitUnavailable(err error) bool { return IsCode(err, CodeRateLimitUnavailable) }

// Also true when the cursor was issued against a different query (hash mismatch).
func IsCursorExpired(err error) bool { return IsCode(err, CodeCursorExpired) }

func IsInvalidCursor(err error) bool { return IsCode(err, CodeInvalidCursor) }

// IsInvalidCostMode means cost_mode= was sent to an endpoint that rejects it.
func IsInvalidCostMode(err error) bool { return IsCode(err, CodeInvalidCostMode) }

// IsValidationError covers schema failures other than cost_mode.
func IsValidationError(err error) bool { return IsCode(err, CodeValidationError) }

func IsBadRequest(err error) bool { return IsCode(err, CodeBadRequest) }

// IsNotFound matches any of the *_NOT_FOUND codes.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.Code {
	case CodeClusterNotFound, CodeNamespaceNotFound, CodeWorkloadNotFound,
		CodeNodeNotFound, CodeNodeGroupNotFound, CodeTeamNotFound,
		CodeDepartmentNotFound, CodeRecommendationNotFound, CodeNotFound:
		return true
	}
	// Fallback for transport-layer 404s raised before the body is parsed.
	return apiErr.StatusCode == http.StatusNotFound
}

func IsServerError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode >= 500 && apiErr.StatusCode <= 599
}

// Superseded by the package-level IsUnauthorized.
func (e *APIError) IsAuthError() bool {
	if e == nil {
		return false
	}
	return e.Code == CodeUnauthorized || e.StatusCode == http.StatusUnauthorized
}

// Superseded by the package-level IsForbidden.
func (e *APIError) IsForbidden() bool {
	if e == nil {
		return false
	}
	return e.Code == CodeForbidden || e.Code == CodeClusterAccessDenied ||
		e.StatusCode == http.StatusForbidden
}

// Superseded by the package-level IsNotFound.
func (e *APIError) IsNotFound() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusNotFound || IsNotFound(e)
}

// Superseded by the package-level IsRateLimited.
func (e *APIError) IsRateLimited() bool {
	if e == nil {
		return false
	}
	return e.Code == CodeRateLimited || e.StatusCode == http.StatusTooManyRequests
}

// Superseded by the package-level IsServerError.
func (e *APIError) IsServerError() bool {
	if e == nil {
		return false
	}
	return e.StatusCode >= 500 && e.StatusCode <= 599
}

// ParseRetryAfter accepts both RFC 7231 §7.1.3 forms - integer seconds or an
// HTTP-date - and returns zero when empty or malformed.
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
