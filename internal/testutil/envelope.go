// Package testutil provides an in-process mock HTTP server and deterministic
// fixtures for testing the Kubeadapt public API client. The mock speaks the
// envelope-shaped /v1 response protocol — every successful response is wrapped
// in {data, meta} and every error response in {data: null, meta, error}.
package testutil

import (
	"encoding/json"
	"net/http"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// Deterministic values used in every mock response so tests stay reproducible
// across runs and across machines.
const (
	mockRequestID = "00000000-0000-0000-0000-mockrequest"
	mockAppliedAt = "2025-05-20T14:30:00Z"

	envKeyData = "data"
	envKeyMeta = "meta"
	envKeyErr  = "error"
)

// defaultMeta returns a Meta block with the canonical mock RequestID and
// AppliedAt populated. Callers can mutate the returned value before passing
// it to WriteEnvelope (e.g. to set CostMode or Pagination).
func defaultMeta() types.Meta {
	return types.Meta{
		RequestID: mockRequestID,
		AppliedAt: mockAppliedAt,
	}
}

// WriteEnvelope writes a 200-OK enveloped response containing the supplied
// payload as `data` and `meta` as the response metadata block. Missing
// RequestID / AppliedAt fields on `meta` are filled in with the canonical
// mock defaults so tests do not need to set them explicitly.
//
// WriteEnvelope does NOT set status — callers that need a non-2xx status
// should use WriteError instead.
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

// WritePaginated wraps a slice payload with a Pagination block and writes
// the enveloped response. The pagination value is attached to meta.Pagination
// before the envelope is serialized.
func WritePaginated(w http.ResponseWriter, items any, meta types.Meta, pagination types.Pagination) {
	meta.Pagination = &pagination
	WriteEnvelope(w, items, meta)
}

// WriteError writes an envelope error response with the supplied HTTP status,
// error code, and message. Optional `details` blocks are attached to the
// `error.details` array as-is (one map per details argument).
//
// When `code` is CodeRateLimited, a Retry-After: 1 header is set before the
// status line is flushed — this is required for the client's retry path to
// observe the header.
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

// writeJSON serializes body as JSON to w. Encoding errors are silently
// dropped — the mock controls its own inputs so any error here is a test
// bug, not a runtime concern, and the response body will already have
// started flushing by the time Encode returns.
func writeJSON(w http.ResponseWriter, body any) {
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	_, _ = w.Write(b)
}

// errorStatusFor returns the conventional HTTP status code for an ErrorCode.
// Used by the mock server when ForceError is set without a matching ForceStatus
// so tests don't need to know the code→status mapping by heart.
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
