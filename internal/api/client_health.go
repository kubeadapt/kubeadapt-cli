package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HealthStatus is the unenveloped response shape of GET /health. The endpoint
// is unauthenticated and does not use the {data, meta} envelope.
type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// Health calls GET /health on the API server. The endpoint is unauthenticated
// (no bearer token sent) and returns a bare JSON object — not the standard
// envelope. A non-2xx response is reported as an *APIError with the raw body
// as the message.
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		excerpt := string(body)
		if len(excerpt) > errorBodyLimit {
			excerpt = excerpt[:errorBodyLimit]
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: excerpt}
	}

	var status HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode health: %w", err)
	}
	return &status, nil
}
