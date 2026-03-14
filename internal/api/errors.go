package api

import "fmt"

// APIError represents an error from the Kubeadapt API.
type APIError struct { //nolint:revive
	StatusCode int    `json:"status_code"`
	Message    string `json:"detail"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsAuthError returns true if the error is an authentication error (401).
func (e *APIError) IsAuthError() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if the error is a forbidden error (403).
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsNotFound returns true if the error is a not found error (404).
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsRateLimited returns true if the error is a rate limit error (429).
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

// IsServerError returns true if the error is a server error (5xx).
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}
