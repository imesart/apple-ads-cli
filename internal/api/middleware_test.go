package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestInjectHost(t *testing.T) {
	mw := InjectHost("api.searchads.apple.com")

	req, err := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if err := mw(req); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	if got := req.URL.Scheme; got != "https" {
		t.Errorf("Scheme = %q, want %q", got, "https")
	}
	if got := req.URL.Host; got != "api.searchads.apple.com" {
		t.Errorf("Host = %q, want %q", got, "api.searchads.apple.com")
	}
}

func TestInjectAcceptHeaders(t *testing.T) {
	mw := InjectAcceptHeaders()

	// Request with a body should get both Accept and Content-Type.
	req, err := http.NewRequest(http.MethodPost, "http://localhost/test", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if err := mw(req); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want %q", got, "application/json")
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
}

func TestInjectAcceptHeaders_NoBody(t *testing.T) {
	mw := InjectAcceptHeaders()

	req, err := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if err := mw(req); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want %q", got, "application/json")
	}
	// Content-Type should NOT be set when there is no body.
	if got := req.Header.Get("Content-Type"); got != "" {
		t.Errorf("Content-Type = %q, want empty (no body)", got)
	}
}

func TestInjectAuthorization(t *testing.T) {
	getToken := func(_ context.Context) (string, error) {
		return "test-token-abc123", nil
	}

	mw := InjectAuthorization(getToken)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if err := mw(req); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	want := "Bearer test-token-abc123"
	if got := req.Header.Get("Authorization"); got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

func TestInjectAuthorization_Error(t *testing.T) {
	tokenErr := errors.New("token provider failure")
	getToken := func(_ context.Context) (string, error) {
		return "", tokenErr
	}

	mw := InjectAuthorization(getToken)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	err = mw(req)
	if err == nil {
		t.Fatal("expected error from middleware, got nil")
	}
	if !errors.Is(err, tokenErr) {
		t.Errorf("error = %v, want wrapping %v", err, tokenErr)
	}
}

func TestInjectOrgContext(t *testing.T) {
	mw := InjectOrgContext("12345")

	req, err := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if err := mw(req); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	want := "orgId=12345"
	if got := req.Header.Get("X-AP-Context"); got != want {
		t.Errorf("X-AP-Context = %q, want %q", got, want)
	}
}

func TestInjectOrgContext_Empty(t *testing.T) {
	mw := InjectOrgContext("")

	req, err := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if err := mw(req); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	if got := req.Header.Get("X-AP-Context"); got != "" {
		t.Errorf("X-AP-Context = %q, want empty when orgID is empty", got)
	}
}
