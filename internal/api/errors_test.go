package api

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	e := &APIError{StatusCode: 401, Message: "Unauthorized"}
	got := e.Error()
	want := "API error (HTTP 401): Unauthorized"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_IsAuthError(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{401, true},
		{200, false},
		{403, false},
		{500, false},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.statusCode}
		if got := e.IsAuthError(); got != tt.want {
			t.Errorf("IsAuthError(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestAPIError_IsForbidden(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{403, true},
		{200, false},
		{401, false},
		{500, false},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.statusCode}
		if got := e.IsForbidden(); got != tt.want {
			t.Errorf("IsForbidden(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestAPIError_IsNotFound(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{404, true},
		{200, false},
		{403, false},
		{500, false},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.statusCode}
		if got := e.IsNotFound(); got != tt.want {
			t.Errorf("IsNotFound(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestAPIError_IsRateLimited(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{429, true},
		{200, false},
		{401, false},
		{500, false},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.statusCode}
		if got := e.IsRateLimited(); got != tt.want {
			t.Errorf("IsRateLimited(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestAPIError_IsServerError(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{500, true},
		{502, true},
		{503, true},
		{400, false},
		{404, false},
		{401, false},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.statusCode}
		if got := e.IsServerError(); got != tt.want {
			t.Errorf("IsServerError(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}
