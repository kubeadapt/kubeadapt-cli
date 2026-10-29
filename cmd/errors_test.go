package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestFriendlyError_RendersDetails(t *testing.T) {
	err := &api.APIError{
		StatusCode: 422,
		Code:       api.CodeValidationError,
		Message:    "invalid provider value: BOGUS",
		Details: []map[string]any{
			{"field": "provider", "allowed": []any{"aws", "gcp"}},
		},
	}

	got := friendlyError(err)

	assert.Contains(t, got, "invalid provider value: BOGUS")
	assert.Contains(t, got, "provider")
	assert.Contains(t, got, "aws")
	assert.Contains(t, got, "gcp")
}

func TestFriendlyError_RendersForbiddenDetails(t *testing.T) {
	err := &api.APIError{
		StatusCode: 403,
		Code:       api.CodeForbidden,
		Message:    "insufficient scope",
		Details: []map[string]any{
			{"granted": []any{"read"}, "required": []any{"write"}},
		},
	}

	got := friendlyError(err)

	assert.Contains(t, got, "granted")
	assert.Contains(t, got, "read")
	assert.Contains(t, got, "required")
	assert.Contains(t, got, "write")
}

func TestFriendlyError_RateLimitUsesRetryAfter(t *testing.T) {
	err := &api.APIError{
		StatusCode: 429,
		Code:       api.CodeRateLimited,
		Message:    "too many requests",
		RetryAfter: 60 * time.Second,
	}

	got := friendlyError(err)

	assert.Contains(t, got, "1 minute")
	assert.NotContains(t, got, "Wait a moment")
}

func TestFriendlyError_RateLimitWithoutRetryAfterKeepsGenericHint(t *testing.T) {
	err := &api.APIError{
		StatusCode: 429,
		Code:       api.CodeRateLimited,
		Message:    "too many requests",
	}

	got := friendlyError(err)

	assert.Contains(t, got, "Rate limited")
	assert.Contains(t, got, "try again")
}

func TestFriendlyError_NotFoundHintIsResourceAware(t *testing.T) {
	tests := []struct {
		code api.ErrorCode
		want string
	}{
		{api.CodeClusterNotFound, "kubeadapt get clusters"},
		{api.CodeNamespaceNotFound, "kubeadapt get namespaces"},
		{api.CodeWorkloadNotFound, "kubeadapt get workloads"},
		{api.CodeNodeNotFound, "kubeadapt get nodes"},
		{api.CodeNodeGroupNotFound, "kubeadapt get node-groups"},
		{api.CodeTeamNotFound, "kubeadapt get teams"},
		{api.CodeDepartmentNotFound, "kubeadapt get departments"},
		{api.CodeRecommendationNotFound, "kubeadapt get recommendations"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := &api.APIError{StatusCode: 404, Code: tt.code, Message: "nope"}
			assert.Contains(t, friendlyError(err), tt.want)
		})
	}
}

func TestFriendlyError_NotFoundUnknownCodeFallsBackToGenericHint(t *testing.T) {
	err := &api.APIError{StatusCode: 404, Code: "SOMETHING_NOT_FOUND", Message: "nope"}

	got := friendlyError(err)

	assert.Contains(t, got, "Resource not found: nope")
	assert.Contains(t, got, "kubeadapt get clusters")
}

func TestFriendlyError_UnknownDetailKeysDoNotPanic(t *testing.T) {
	err := &api.APIError{
		StatusCode: 422,
		Code:       api.CodeValidationError,
		Message:    "bad request",
		Details: []map[string]any{
			{
				"field":  "spec",
				"nested": map[string]any{"deep": []any{1, 2, map[string]any{"x": true}}},
				"count":  float64(3),
				"nilval": nil,
			},
			{},
		},
	}

	assert.NotPanics(t, func() {
		got := friendlyError(err)
		assert.Contains(t, got, "bad request")
	})
}

// Every code the API sends, at the status it sends it with.
func TestFriendlyError_EveryCatalogCode(t *testing.T) {
	tests := []struct {
		code   api.ErrorCode
		status int
		want   string
	}{
		{api.CodeUnauthorized, 401, "Authentication failed"},
		{api.CodeForbidden, 403, "Access denied"},
		{api.CodeClusterAccessDenied, 403, "Access denied"},
		{api.CodeRateLimited, 429, "Rate limited"},
		{api.CodeRateLimitUnavailable, 503, "cannot check rate limits"},
		{api.CodeBadRequest, 400, "server says"},
		{api.CodeBadRequest, 404, "Resource not found"},
		{api.CodeBadRequest, 405, "server says"},
		{api.CodeValidationError, 422, "server says"},
		{api.CodeInvalidCursor, 400, "server says"},
		{api.CodeCursorExpired, 410, "server says"},
		{api.CodeInvalidCostMode, 422, "server says"},
		{api.CodeInvalidClusterID, 400, "server says"},
		{api.CodeClusterNotFound, 404, "Resource not found"},
		{api.CodeNamespaceNotFound, 404, "Resource not found"},
		{api.CodeWorkloadNotFound, 404, "Resource not found"},
		{api.CodeNodeNotFound, 404, "Resource not found"},
		{api.CodeNodeGroupNotFound, 404, "Resource not found"},
		{api.CodeTeamNotFound, 404, "Resource not found"},
		{api.CodeDepartmentNotFound, 404, "Resource not found"},
		{api.CodeRecommendationNotFound, 404, "Resource not found"},
		{api.CodeInternalError, 500, "experiencing issues"},
		{api.CodeServiceUnavailable, 503, "experiencing issues"},
		{api.CodeUpstreamUnavailable, 502, "service behind the API is unavailable"},
		{api.CodeUpstreamTimeout, 504, "timed out inside the API"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.code, tt.status), func(t *testing.T) {
			got := friendlyError(&api.APIError{StatusCode: tt.status, Code: tt.code, Message: "server says"})
			assert.Contains(t, got, tt.want)
		})
	}
}

func TestFriendlyError_NoClusterInScope(t *testing.T) {
	err := fmt.Errorf("get clusters: %w", &api.APIError{
		StatusCode: 403,
		Code:       api.CodeClusterAccessDenied,
		Message:    "no cluster in scope for this API key",
		Details:    []map[string]any{{"field": "cluster_id", "reason": "no_cluster_in_scope"}},
	})

	got := friendlyError(err)

	assert.Contains(t, got, "this API key has no cluster in scope for the request")
	assert.NotContains(t, got, "no_cluster_in_scope", "the raw reason is replaced, not repeated")
	assert.NotContains(t, got, "permission for this resource", "the generic 403 text would send the user to the wrong fix")
}

func TestFriendlyError_NonEnvelopeResponse(t *testing.T) {
	t.Run("html 404 is not a missing resource", func(t *testing.T) {
		got := friendlyError(&api.APIError{
			StatusCode:  404,
			Message:     "unexpected response from example.com (HTTP 404, text/html)",
			NotEnvelope: true,
		})
		assert.True(t, strings.HasPrefix(got, "unexpected response from example.com (HTTP 404, text/html)"), got)
		assert.Contains(t, got, "API URL")
		assert.NotContains(t, got, "Resource not found")
	})
	t.Run("proxy 429 keeps the retry guidance", func(t *testing.T) {
		got := friendlyError(&api.APIError{
			StatusCode:  429,
			Code:        api.CodeRateLimited,
			Message:     "unexpected response from example.com (HTTP 429, text/plain)",
			NotEnvelope: true,
			RetryAfter:  30 * time.Second,
		})
		assert.Contains(t, got, "Retry after 30 seconds")
	})
}
