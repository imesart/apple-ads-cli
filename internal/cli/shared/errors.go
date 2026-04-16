package shared

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/imesart/apple-ads-cli/internal/api"
)

// reportedError is an error that has already been printed to stderr.
type reportedError struct {
	err error
}

func (e *reportedError) Error() string { return e.err.Error() }
func (e *reportedError) Unwrap() error { return e.err }

// ReportError prints the error to stderr and wraps it as reported.
func ReportError(err error) error {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return &reportedError{err: err}
}

// IsReportedError returns true if the error has already been reported to the user.
func IsReportedError(err error) bool {
	var re *reportedError
	return errors.As(err, &re)
}

// safetyError represents a safety limit violation.
type safetyError struct {
	msg string
}

func (e *safetyError) Error() string { return e.msg }

// NewSafetyError creates a safety limit error.
func NewSafetyError(msg string) error {
	return &safetyError{msg: msg}
}

// IsSafetyError returns true if the error is a safety limit violation.
func IsSafetyError(err error) bool {
	var se *safetyError
	return errors.As(err, &se)
}

// IsAuthError returns true if the error is an authentication error.
func IsAuthError(err error) bool {
	return api.IsAuthError(err)
}

// IsAPIError returns true if the error is an API error (4xx).
func IsAPIError(err error) bool {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
	}
	return false
}

// IsNetworkError returns true if the error is a network/server error.
func IsNetworkError(err error) bool {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	return false
}

// usageError wraps an error message with flag.ErrHelp so ffcli prints usage.
type usageError struct {
	msg string
}

func (e *usageError) Error() string { return e.msg }
func (e *usageError) Unwrap() error { return flag.ErrHelp }

// UsageError prints an error message to stderr and returns an error that
// triggers ffcli's usage display.
func UsageError(msg string) error {
	fmt.Fprintf(os.Stderr, "\nError: %s\n\n", msg)
	return &usageError{msg: msg}
}

// UsageErrorf prints a formatted error message and triggers usage display.
func UsageErrorf(format string, args ...any) error {
	return UsageError(fmt.Sprintf(format, args...))
}

// IsUsageError returns true if the error is a usage error.
func IsUsageError(err error) bool {
	var ue *usageError
	return errors.As(err, &ue)
}

// validationError represents a recognized flag/input value error that should
// exit with usage status but should not print full command help.
type validationError struct {
	msg string
}

func (e *validationError) Error() string { return e.msg }

// ValidationError prints an error message to stderr and returns an error that
// exits with usage status without triggering ffcli's usage display.
func ValidationError(msg string) error {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	return &validationError{msg: msg}
}

// ValidationErrorf prints a formatted validation error message.
func ValidationErrorf(format string, args ...any) error {
	return ValidationError(fmt.Sprintf(format, args...))
}

// IsValidationError returns true if the error is a validation error.
func IsValidationError(err error) bool {
	var ve *validationError
	return errors.As(err, &ve)
}
