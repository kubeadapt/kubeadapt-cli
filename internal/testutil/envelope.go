// Package testutil provides an in-process mock of the envelope-shaped /v1
// protocol: {data, meta} on success, {data: null, meta, error} on failure.
package testutil

import (
	"encoding/json"
	"net/http"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// Deterministic so tests stay reproducible across runs and machines.
const (
	mockRequestID = "00000000-0000-0000-0000-mockrequest"
	mockAppliedAt = "2025-05-20T14:30:00Z"

	envKeyData = "data"
	envKeyMeta = "meta"
	envKeyErr  = "error"
)

func defaultMeta() types.Meta {
	return types.Meta{
		RequestID: mockRequestID,
		AppliedAt: mockAppliedAt,
	}
}

// Missing RequestID / AppliedAt are filled with the mock defaults. Does not set
// status; use WriteError for non-2xx.
func WriteEnvelope(w http.ResponseWriter, data any, meta types.Meta) {
	if meta.RequestID == "" {
		meta.RequestID = mockRequestID
	}
	if meta.AppliedAt == "" {
		meta.AppliedAt = mockAppliedAt
	}
	body := struct {
		Data any        `json:"data"`
		Meta types.Meta `json:"meta"`
	}{Data: data, Meta: meta}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, body)
}

func WritePaginated(w http.ResponseWriter, items any, meta types.Meta, pagination types.Pagination) {
	meta.Pagination = &pagination
	WriteEnvelope(w, items, meta)
}

// For CodeRateLimited a Retry-After: 1 header is set before the status line is
// flushed, which the client's retry path requires in order to see it.
func WriteError(w http.ResponseWriter, status int, code api.ErrorCode, msg string, details ...map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	if code == api.CodeRateLimited {
		w.Header().Set("Retry-After", "1")
	}
	w.WriteHeader(status)
	errBody := map[string]any{
		"code":    string(code),
		"message": msg,
	}
	if len(details) > 0 {
		errBody["details"] = details
	}
	body := map[string]any{
		envKeyData: nil,
		envKeyMeta: map[string]any{
			"request_id": mockRequestID,
			"applied_at": mockAppliedAt,
		},
		envKeyErr: errBody,
	}
	writeJSON(w, body)
}

// Encoding errors are dropped: the mock controls its inputs, and the body has
// already begun flushing by the time Encode returns.
func writeJSON(w http.ResponseWriter, body any) {
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	_, _ = w.Write(b)
}

// Lets tests set ForceError without also knowing the code-to-status mapping.
func errorStatusFor(code api.ErrorCode) int { //nolint:gocyclo // simple lookup table
	switch code {
	case api.CodeUnauthorized:
		return http.StatusUnauthorized
	case api.CodeForbidden, api.CodeClusterAccessDenied:
		return http.StatusForbidden
	case api.CodeRateLimited:
		return http.StatusTooManyRequests
	case api.CodeRateLimitUnavailable, api.CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case api.CodeBadRequest, api.CodeInvalidCursor:
		return http.StatusBadRequest
	case api.CodeValidationError, api.CodeInvalidCostMode, api.CodeInvalidClusterID:
		return http.StatusUnprocessableEntity
	case api.CodeCursorExpired:
		return http.StatusGone
	case api.CodeClusterNotFound, api.CodeNamespaceNotFound, api.CodeWorkloadNotFound,
		api.CodeNodeNotFound, api.CodeNodeGroupNotFound, api.CodeTeamNotFound,
		api.CodeDepartmentNotFound, api.CodeRecommendationNotFound:
		return http.StatusNotFound
	case api.CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
