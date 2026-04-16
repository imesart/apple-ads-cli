package api

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

// WithRetry wraps an operation with exponential backoff retry.
// It retries on 429 (rate limit) and 5xx (server error) responses.
// Delays follow exponential backoff: 2s, 4s, 8s, 16s with jitter of +/-25%.
func WithRetry[T any](ctx context.Context, maxAttempts int, fn func() (T, error)) (T, error) {
	var zero T
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	baseDelay := 2 * time.Second

	var lastErr error
	for attempt := range maxAttempts {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !IsRetryable(err) {
			return zero, err
		}

		// Don't sleep after the last attempt
		if attempt == maxAttempts-1 {
			break
		}

		// Exponential backoff: 2s, 4s, 8s, 16s
		delay := baseDelay << uint(attempt)

		// Apply jitter of +/-25%
		jitter := 1.0 + (rand.Float64()-0.5)*0.5 // range [0.75, 1.25]
		delay = time.Duration(float64(delay) * jitter)

		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// continue to next attempt
		}
	}

	return zero, fmt.Errorf("all %d attempts failed: %w", maxAttempts, lastErr)
}

// WithAuthRetry wraps an operation with authentication retry.
// On a 401 response, it invalidates the current token and retries once.
func WithAuthRetry[T any](ctx context.Context, invalidateToken func(), fn func() (T, error)) (T, error) {
	result, err := fn()
	if err == nil {
		return result, nil
	}

	if !IsAuthError(err) {
		return result, err
	}

	// Invalidate the cached token and retry once
	invalidateToken()
	return fn()
}
