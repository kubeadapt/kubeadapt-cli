package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "nil receiver",
			err:  nil,
			want: "<nil APIError>",
		},
		{
			name: "with code",
			err:  &APIError{StatusCode: 401, Code: CodeUnauthorized, Message: "missing api key"},
			want: "API error (HTTP 401, code=UNAUTHORIZED): missing api key",
		},
		{
			name: "without code",
			err:  &APIError{StatusCode: 502, Message: "bad gateway"},
			want: "API error (HTTP 502): bad gateway",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

// TestAPIError_HelperPredicates covers every package-level helper that maps
// 1:1 to a single ErrorCode. IsNotFound is excluded (multiple codes); see
// TestAPIError_IsNotFound.
func TestAPIError_HelperPredicates(t *testing.T) {
	type helperCase struct {
		name     string
		fn       func(error) bool
		matching ErrorCode
	}
	helpers := []helperCase{
		{"IsUnauthorized", IsUnauthorized, CodeUnauthorized},
		{"IsForbidden", IsForbidden, CodeForbidden},
		{"IsClusterAccessDenied", IsClusterAccessDenied, CodeClusterAccessDenied},
		{"IsRateLimited", IsRateLimited, CodeRateLimited},
		{"IsRateLimitUnavailable", IsRateLimitUnavailable, CodeRateLimitUnavailable},
		{"IsCursorExpired", IsCursorExpired, CodeCursorExpired},
		{"IsInvalidCursor", IsInvalidCursor, CodeInvalidCursor},
		{"IsInvalidCostMode", IsInvalidCostMode, CodeInvalidCostMode},
		{"IsValidationError", IsValidationError, CodeValidationError},
		{"IsBadRequest", IsBadRequest, CodeBadRequest},
	}
	for _, h := range helpers {
		t.Run(h.name, func(t *testing.T) {
			matching := &APIError{Code: h.matching}
			assert.True(t, h.fn(matching), "%s(%s) = false, want true", h.name, h.matching)

			nonMatching := &APIError{Code: CodeInternalError}
			if h.matching == CodeInternalError {
				nonMatching = &APIError{Code: CodeUnauthorized}
			}
			assert.False(t, h.fn(nonMatching), "%s(%s) = true, want false", h.name, nonMatching.Code)

			assert.False(t, h.fn(nil), "%s(nil) = true, want false", h.name)
			assert.False(t, h.fn(errors.New("plain error")), "%s(non-APIError) = true, want false", h.name)
		})
	}

	t.Run("IsServerError 500-599 true", func(t *testing.T) {
		for _, sc := range []int{500, 502, 503, 599} {
			e := &APIError{StatusCode: sc}
			assert.True(t, IsServerError(e), "IsServerError(%d) = false, want true", sc)
		}
	})
	t.Run("IsServerError out of range false", func(t *testing.T) {
		for _, sc := range []int{200, 400, 404, 600} {
			e := &APIError{StatusCode: sc}
			assert.False(t, IsServerError(e), "IsServerError(%d) = true, want false", sc)
		}
	})
	t.Run("IsServerError nil/plain", func(t *testing.T) {
		assert.False(t, IsServerError(nil))
		assert.False(t, IsServerError(errors.New("plain")))
	})

	t.Run("IsCode unwraps via errors.As", func(t *testing.T) {
		base := &APIError{StatusCode: 401, Code: CodeUnauthorized, Message: "no key"}
		wrapped := fmt.Errorf("call to /api/v1/foo failed: %w", base)
		assert.True(t, IsUnauthorized(wrapped))
		assert.True(t, IsCode(wrapped, CodeUnauthorized))
	})
}

func TestAPIError_IsNotFound(t *testing.T) {
	notFoundCodes := []ErrorCode{
		CodeClusterNotFound,
		CodeNamespaceNotFound,
		CodeWorkloadNotFound,
		CodeNodeNotFound,
		CodeNodeGroupNotFound,
		CodeTeamNotFound,
		CodeDepartmentNotFound,
		CodeRecommendationNotFound,
	}
	for _, c := range notFoundCodes {
		t.Run(string(c), func(t *testing.T) {
			e := &APIError{StatusCode: 404, Code: c}
			assert.True(t, IsNotFound(e), "IsNotFound(%s) = false, want true", c)
		})
	}

	t.Run("status-only 404 fallback", func(t *testing.T) {
		e := &APIError{StatusCode: http.StatusNotFound}
		assert.True(t, IsNotFound(e))
	})

	t.Run("non-matching code", func(t *testing.T) {
		e := &APIError{StatusCode: 401, Code: CodeUnauthorized}
		assert.False(t, IsNotFound(e))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, IsNotFound(nil))
	})

	t.Run("non-APIError", func(t *testing.T) {
		assert.False(t, IsNotFound(errors.New("plain")))
	})
}

func TestAPIError_MethodAliases(t *testing.T) {
	t.Run("IsAuthError via code", func(t *testing.T) {
		e := &APIError{Code: CodeUnauthorized}
		assert.True(t, e.IsAuthError())
	})
	t.Run("IsAuthError via status", func(t *testing.T) {
		e := &APIError{StatusCode: 401}
		assert.True(t, e.IsAuthError())
	})
	t.Run("IsAuthError negative", func(t *testing.T) {
		e := &APIError{StatusCode: 200}
		assert.False(t, e.IsAuthError())
	})
	t.Run("IsAuthError nil", func(t *testing.T) {
		var e *APIError
		assert.False(t, e.IsAuthError())
	})

	t.Run("IsForbidden method via code", func(t *testing.T) {
		e := &APIError{Code: CodeForbidden}
		assert.True(t, e.IsForbidden())
	})
	t.Run("IsForbidden method via cluster-access code", func(t *testing.T) {
		e := &APIError{Code: CodeClusterAccessDenied}
		assert.True(t, e.IsForbidden())
	})
	t.Run("IsForbidden method via status", func(t *testing.T) {
		e := &APIError{StatusCode: 403}
		assert.True(t, e.IsForbidden())
	})
	t.Run("IsForbidden method negative + nil", func(t *testing.T) {
		assert.False(t, (&APIError{StatusCode: 200}).IsForbidden())
		var e *APIError
		assert.False(t, e.IsForbidden())
	})

	t.Run("IsNotFound method via status", func(t *testing.T) {
		e := &APIError{StatusCode: 404}
		assert.True(t, e.IsNotFound())
	})
	t.Run("IsNotFound method via code", func(t *testing.T) {
		e := &APIError{Code: CodeClusterNotFound}
		assert.True(t, e.IsNotFound())
	})
	t.Run("IsNotFound method negative + nil", func(t *testing.T) {
		assert.False(t, (&APIError{StatusCode: 200}).IsNotFound())
		var e *APIError
		assert.False(t, e.IsNotFound())
	})

	t.Run("IsRateLimited method via code", func(t *testing.T) {
		e := &APIError{Code: CodeRateLimited}
		assert.True(t, e.IsRateLimited())
	})
	t.Run("IsRateLimited method via status", func(t *testing.T) {
		e := &APIError{StatusCode: 429}
		assert.True(t, e.IsRateLimited())
	})
	t.Run("IsRateLimited method negative + nil", func(t *testing.T) {
		assert.False(t, (&APIError{StatusCode: 200}).IsRateLimited())
		var e *APIError
		assert.False(t, e.IsRateLimited())
	})

	t.Run("IsServerError method", func(t *testing.T) {
		for _, sc := range []int{500, 503, 599} {
			assert.True(t, (&APIError{StatusCode: sc}).IsServerError(), "IsServerError(%d) = false, want true", sc)
		}
		for _, sc := range []int{200, 400, 404, 600} {
			assert.False(t, (&APIError{StatusCode: sc}).IsServerError(), "IsServerError(%d) = true, want false", sc)
		}
		var e *APIError
		assert.False(t, e.IsServerError())
	})
}

func TestParseRetryAfter(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ParseRetryAfter(""))
	})
	t.Run("60 seconds", func(t *testing.T) {
		assert.Equal(t, 60*time.Second, ParseRetryAfter("60"))
	})
	t.Run("zero", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ParseRetryAfter("0"))
	})
	t.Run("negative seconds", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ParseRetryAfter("-1"))
	})
	t.Run("garbage", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), ParseRetryAfter("abc"))
	})
	t.Run("http-date in future", func(t *testing.T) {
		future := time.Now().Add(5 * time.Second).UTC().Format(http.TimeFormat)
		got := ParseRetryAfter(future)
		assert.GreaterOrEqual(t, got, 3*time.Second, "ParseRetryAfter(future+5s) = %v, want ~5s", got)
		assert.LessOrEqual(t, got, 6*time.Second, "ParseRetryAfter(future+5s) = %v, want ~5s", got)
	})
	t.Run("http-date in past", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
		assert.Equal(t, time.Duration(0), ParseRetryAfter(past))
	})
}
