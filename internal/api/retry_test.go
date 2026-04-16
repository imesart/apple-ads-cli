package api

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRetry_Success_NoRetry(t *testing.T) {
	var calls int32

	result, err := WithRetry(context.Background(), 3, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("WithRetry returned error: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %q, want %q", result, "ok")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestWithRetry_RetryOn429(t *testing.T) {
	// This test uses a short context timeout to prevent the full backoff delay.
	// The first attempt fails with 429, then we rely on the backoff delay being
	// interrupted by context. To actually test successful retry we need to wait
	// for the backoff. We use maxAttempts=2 and accept the ~2s delay.
	if testing.Short() {
		t.Skip("skipping slow retry test in short mode")
	}

	var calls int32

	result, err := WithRetry(context.Background(), 2, func() (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return "", &APIError{StatusCode: 429}
		}
		return "recovered", nil
	})

	if err != nil {
		t.Fatalf("WithRetry returned error: %v", err)
	}
	if result != "recovered" {
		t.Errorf("result = %q, want %q", result, "recovered")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestWithRetry_RetryOn500(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow retry test in short mode")
	}

	var calls int32

	result, err := WithRetry(context.Background(), 2, func() (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return "", &APIError{StatusCode: 500}
		}
		return "recovered", nil
	})

	if err != nil {
		t.Fatalf("WithRetry returned error: %v", err)
	}
	if result != "recovered" {
		t.Errorf("result = %q, want %q", result, "recovered")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestWithRetry_ExhaustedRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow retry test in short mode")
	}

	var calls int32

	_, err := WithRetry(context.Background(), 2, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &APIError{StatusCode: 429}
	})

	if err == nil {
		t.Fatal("expected error after exhausted retries, got nil")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}

	// The final error should wrap the original APIError.
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("error should wrap *APIError, got: %v", err)
	}
}

func TestWithRetry_NonRetryable(t *testing.T) {
	var calls int32

	_, err := WithRetry(context.Background(), 3, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &APIError{StatusCode: 400}
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (no retry for non-retryable errors)", got)
	}
}

func TestWithRetry_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var calls int32

	// Cancel context after first call to prevent sleeping through backoff.
	_, err := WithRetry(ctx, 3, func() (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			// Cancel context so the retry backoff select picks up ctx.Done().
			cancel()
			return "", &APIError{StatusCode: 429}
		}
		return "should not reach", nil
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error should wrap context.Canceled, got: %v", err)
	}
}

func TestWithRetry_MaxAttemptsLessThanOne(t *testing.T) {
	var calls int32

	_, err := WithRetry(context.Background(), 0, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &APIError{StatusCode: 500}
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// maxAttempts < 1 is clamped to 1, so exactly 1 call should be made.
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestWithAuthRetry_Success(t *testing.T) {
	var calls int32
	var invalidated bool

	result, err := WithAuthRetry(context.Background(), func() {
		invalidated = true
	}, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("WithAuthRetry returned error: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %q, want %q", result, "ok")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
	if invalidated {
		t.Error("invalidateToken should not be called on success")
	}
}

func TestWithAuthRetry_RefreshOnce(t *testing.T) {
	var calls int32
	var invalidated bool

	result, err := WithAuthRetry(context.Background(), func() {
		invalidated = true
	}, func() (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return "", &APIError{StatusCode: 401}
		}
		return "refreshed", nil
	})

	if err != nil {
		t.Fatalf("WithAuthRetry returned error: %v", err)
	}
	if result != "refreshed" {
		t.Errorf("result = %q, want %q", result, "refreshed")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
	if !invalidated {
		t.Error("invalidateToken should have been called")
	}
}

func TestWithAuthRetry_DoubleFailure(t *testing.T) {
	var calls int32
	var invalidateCount int32

	_, err := WithAuthRetry(context.Background(), func() {
		atomic.AddInt32(&invalidateCount, 1)
	}, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &APIError{StatusCode: 401}
	})

	if err == nil {
		t.Fatal("expected error after double 401, got nil")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2 (initial + 1 retry)", got)
	}
	if got := atomic.LoadInt32(&invalidateCount); got != 1 {
		t.Errorf("invalidateToken calls = %d, want 1", got)
	}

	// Should still be an auth error.
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error should wrap ErrUnauthorized, got: %v", err)
	}
}

func TestWithAuthRetry_NonAuthError_NoRetry(t *testing.T) {
	var calls int32
	var invalidated bool

	_, err := WithAuthRetry(context.Background(), func() {
		invalidated = true
	}, func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &APIError{StatusCode: 400}
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (no retry for non-auth errors)", got)
	}
	if invalidated {
		t.Error("invalidateToken should not be called for non-auth errors")
	}
}

func TestWithRetry_TimingBasic(t *testing.T) {
	// Verify that retry actually delays (sanity check).
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}

	start := time.Now()

	_, _ = WithRetry(context.Background(), 2, func() (string, error) {
		return "", &APIError{StatusCode: 429}
	})

	elapsed := time.Since(start)
	// The first retry delay is ~2s (with jitter 1.5s-2.5s).
	// We check it's at least 1 second to verify delay happened.
	if elapsed < 1*time.Second {
		t.Errorf("retry should have delayed at least 1s, elapsed: %v", elapsed)
	}
}
