package cmd

import (
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
