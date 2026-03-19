package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
)

// FlagError indicates a usage/flag error. When returned, the CLI prints usage help.
type FlagError struct {
	Err error
}

func (e *FlagError) Error() string { return e.Err.Error() }
func (e *FlagError) Unwrap() error { return e.Err }

// flagErrorf creates a new FlagError.
func flagErrorf(format string, a ...any) *FlagError {
	return &FlagError{Err: fmt.Errorf(format, a...)}
}

// friendlyError translates raw errors into user-friendly messages with actionable guidance.
func friendlyError(err error) string {
	if err == nil {
		return ""
	}

	// API errors — use the status code helpers
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.IsAuthError():
			return "Authentication failed: API key is invalid or expired.\n  Run 'kubeadapt auth login' to re-authenticate."
		case apiErr.IsForbidden():
			return "Access denied: you don't have permission for this resource.\n  Check your API key permissions at https://app.kubeadapt.io/settings/api-keys"
		case apiErr.IsNotFound():
			return fmt.Sprintf("Resource not found: %s\n  Use 'kubeadapt get clusters' to list available resources.", apiErr.Message)
		case apiErr.IsRateLimited():
			return "Rate limited: too many requests.\n  Wait a moment and try again."
		case apiErr.IsServerError():
			return "Kubeadapt API error: the service is experiencing issues.\n  Check https://status.kubeadapt.io or try again later."
		default:
			return fmt.Sprintf("API error (HTTP %d): %s", apiErr.StatusCode, apiErr.Message)
		}
	}

	// Network errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fmt.Sprintf("Cannot resolve %s: check your network connection and --api-url flag.", dnsErr.Name)
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "dial" {
			return "Cannot connect to Kubeadapt API: connection refused.\n  Check your network connection or use --api-url to set a different endpoint."
		}
		return fmt.Sprintf("Network error: %s\n  Check your network connection.", opErr.Err)
	}

	// Timeout
	if isTimeout(err) {
		return "Request timed out: the API did not respond in time.\n  Try again or use --verbose to see request details."
	}

	// Auth-related errors (from helpers.go)
	msg := err.Error()
	if strings.Contains(msg, "not authenticated") || strings.Contains(msg, "no API key") {
		return msg // already user-friendly
	}

	return msg
}

// isTimeout checks if an error is a timeout.
func isTimeout(err error) bool {
	type timeouter interface {
		Timeout() bool
	}
	var t timeouter
	return errors.As(err, &t) && t.Timeout()
}
