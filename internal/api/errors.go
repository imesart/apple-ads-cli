package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/types"
)

// Sentinel errors for common API error conditions.
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrRateLimit    = errors.New("rate limited")
)

// APIError represents an error response from the Apple Search Ads API.
type APIError struct {
	StatusCode int
	Errors     []types.ErrorItem
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("api error: HTTP %d", e.StatusCode)
	}

	var parts []string
	for _, item := range e.Errors {
		msg := derefStr(item.Message)
		if msg == "" {
			msg = derefStr(item.MessageCode)
		}
		if field := derefStr(item.Field); field != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
		} else {
			parts = append(parts, msg)
		}
	}
	return fmt.Sprintf("api error: HTTP %d: %s", e.StatusCode, strings.Join(parts, "; "))
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Unwrap returns the appropriate sentinel error based on status code.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 429:
		return ErrRateLimit
	default:
		return nil
	}
}

// IsRetryable returns true if the error represents a retryable condition (429 or 5xx).
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}
	return false
}

// IsAuthError returns true if the error represents an authentication failure (401).
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return errors.Is(err, ErrUnauthorized)
}
