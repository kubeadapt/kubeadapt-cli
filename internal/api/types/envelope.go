package types

// APIErrorBody is the raw error payload inside an Envelope. Consumers should
// rely on api.APIError, which the HTTP client decodes this into.
type APIErrorBody struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []map[string]any `json:"details,omitempty"`
}

// Envelope wraps every API response: success has non-nil Data and nil Error,
// failure the inverse. Data holds T directly, so a nil-able payload needs *T
// as the parameter (e.g. Envelope[*Cluster]).
type Envelope[T any] struct {
	Data  T             `json:"data"`
	Meta  Meta          `json:"meta"`
	Error *APIErrorBody `json:"error,omitempty"`
}
