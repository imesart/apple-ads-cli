package api

import (
	"errors"
	"strings"
	"testing"

	"github.com/imesart/apple-ads-cli/internal/types"
)

func strPtr(s string) *string { return &s }

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 400,
		Errors: []types.ErrorItem{
			{Message: strPtr("bad request"), MessageCode: strPtr("INVALID")},
			{Field: strPtr("campaignId"), Message: strPtr("invalid value")},
		},
	}

	got := apiErr.Error()
	if !strings.Contains(got, "400") {
		t.Errorf("error string %q should contain status code 400", got)
	}
	if !strings.Contains(got, "bad request") {
		t.Errorf("error string %q should contain first error message", got)
	}
	if !strings.Contains(got, "campaignId") {
		t.Errorf("error string %q should contain field name", got)
	}
	if !strings.Contains(got, "invalid value") {
		t.Errorf("error string %q should contain second error message", got)
	}
}

func TestAPIError_Error_NoErrors(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 500,
		Errors:     nil,
	}

	got := apiErr.Error()
	want := "api error: HTTP 500"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_Error_MessageCodeFallback(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 400,
		Errors: []types.ErrorItem{
			{MessageCode: strPtr("INVALID_FIELD")},
		},
	}

	got := apiErr.Error()
	if !strings.Contains(got, "INVALID_FIELD") {
		t.Errorf("error string %q should fall back to messageCode when message is empty", got)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"429 rate limit", 429, true},
		{"500 server error", 500, true},
		{"503 service unavailable", 503, true},
		{"502 bad gateway", 502, true},
		{"400 bad request", 400, false},
		{"401 unauthorized", 401, false},
		{"404 not found", 404, false},
		{"200 ok", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			got := IsRetryable(err)
			if got != tt.want {
				t.Errorf("IsRetryable(status=%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestIsRetryable_NonAPIError(t *testing.T) {
	err := errors.New("some random error")
	if IsRetryable(err) {
		t.Error("IsRetryable should return false for non-APIError")
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"401 unauthorized", 401, true},
		{"403 forbidden", 403, false},
		{"200 ok", 200, false},
		{"429 rate limit", 429, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			got := IsAuthError(err)
			if got != tt.want {
				t.Errorf("IsAuthError(status=%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestIsAuthError_SentinelError(t *testing.T) {
	if !IsAuthError(ErrUnauthorized) {
		t.Error("IsAuthError(ErrUnauthorized) should return true")
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		sentinel   error
	}{
		{"401 wraps ErrUnauthorized", 401, ErrUnauthorized},
		{"403 wraps ErrForbidden", 403, ErrForbidden},
		{"404 wraps ErrNotFound", 404, ErrNotFound},
		{"429 wraps ErrRateLimit", 429, ErrRateLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &APIError{StatusCode: tt.statusCode}
			if !errors.Is(apiErr, tt.sentinel) {
				t.Errorf("errors.Is(APIError{%d}, %v) = false, want true", tt.statusCode, tt.sentinel)
			}
		})
	}
}

func TestAPIError_Unwrap_NoSentinel(t *testing.T) {
	apiErr := &APIError{StatusCode: 500}
	if apiErr.Unwrap() != nil {
		t.Errorf("Unwrap() for status 500 should return nil, got %v", apiErr.Unwrap())
	}
}

func TestAPIError_ImplementsError(t *testing.T) {
	var _ error = &APIError{}
}
