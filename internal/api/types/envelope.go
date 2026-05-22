// Package types contains shared API response types used by the kubeadapt-cli
// to decode envelope-shaped responses from the Kubeadapt public API.
package types

// APIErrorBody is the structured error payload returned inside an Envelope when
// the API rejects a request. It is decoded into internal/api.APIError by the
// HTTP client; downstream consumers should rely on api.APIError, not on this
// raw shape.
type APIErrorBody struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []map[string]any `json:"details,omitempty"`
}

// Envelope is the universal response wrapper used by the Kubeadapt public API.
// Every successful response has a non-nil Data payload (which may itself be an
// empty slice); every error response has a non-nil Error and a nil/zero Data.
//
// The type parameter T is the resource payload (e.g. *Cluster, []Workload,
// *OrganizationDashboard). The Data field is the typed payload itself, not a
// pointer to it — callers that want a nil-able payload should use *T as the
// type parameter (e.g. Envelope[*Cluster]).
type Envelope[T any] struct {
	Data  T             `json:"data"`
	Meta  Meta          `json:"meta"`
	Error *APIErrorBody `json:"error,omitempty"`
}
