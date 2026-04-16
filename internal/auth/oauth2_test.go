package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/imesart/apple-ads-cli/internal/types"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// exchangeTokenWithURL mirrors ExchangeToken but uses a custom token endpoint URL.
// This is necessary because ExchangeToken uses a package-level const for the URL.
func exchangeTokenWithURL(ctx context.Context, tokenEndpoint, clientID, clientSecret string) (*types.AccessToken, error) {
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {"searchadsorg"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var token types.AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	token.ObtainedAt = time.Now()
	return &token, nil
}

func withMockDefaultTransport(t *testing.T, fn roundTripFunc) {
	t.Helper()
	orig := http.DefaultTransport
	http.DefaultTransport = fn
	t.Cleanup(func() { http.DefaultTransport = orig })
}

func TestExchangeToken_Success(t *testing.T) {
	var receivedContentType string
	var receivedBody string

	withMockDefaultTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		respBody, _ := json.Marshal(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(respBody))),
		}, nil
	}))

	token, err := exchangeTokenWithURL(context.Background(), "https://example.test/token", "test-client", "test-secret")
	if err != nil {
		t.Fatalf("exchangeTokenWithURL() error: %v", err)
	}

	// Verify token fields
	if token.Token != "test-token" {
		t.Errorf("Token = %q, want %q", token.Token, "test-token")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", token.TokenType, "Bearer")
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want %d", token.ExpiresIn, 3600)
	}
	if token.ObtainedAt.IsZero() {
		t.Error("ObtainedAt should not be zero")
	}

	// Verify request Content-Type
	if receivedContentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want %q", receivedContentType, "application/x-www-form-urlencoded")
	}

	// Verify request body contains expected form values
	values, err := url.ParseQuery(receivedBody)
	if err != nil {
		t.Fatalf("parsing request body: %v", err)
	}
	if v := values.Get("grant_type"); v != "client_credentials" {
		t.Errorf("grant_type = %q, want %q", v, "client_credentials")
	}
	if v := values.Get("client_id"); v != "test-client" {
		t.Errorf("client_id = %q, want %q", v, "test-client")
	}
	if v := values.Get("client_secret"); v != "test-secret" {
		t.Errorf("client_secret = %q, want %q", v, "test-secret")
	}
	if v := values.Get("scope"); v != "searchadsorg" {
		t.Errorf("scope = %q, want %q", v, "searchadsorg")
	}
}

func TestExchangeToken_Error(t *testing.T) {
	withMockDefaultTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`{"error":"invalid_client"}`)),
		}, nil
	}))

	_, err := exchangeTokenWithURL(context.Background(), "https://example.test/token", "bad-client", "bad-secret")
	if err == nil {
		t.Fatal("exchangeTokenWithURL() should return error on 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should mention status 400, got: %v", err)
	}
}

func TestExchangeToken_InvalidJSON(t *testing.T) {
	withMockDefaultTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{not valid json`)),
		}, nil
	}))

	_, err := exchangeTokenWithURL(context.Background(), "https://example.test/token", "client", "secret")
	if err == nil {
		t.Fatal("exchangeTokenWithURL() should return error on invalid JSON")
	}
	if !strings.Contains(err.Error(), "decoding") {
		t.Errorf("error should mention decoding, got: %v", err)
	}
}

func TestExchangeToken_RequestFormat(t *testing.T) {
	var method string
	var receivedValues url.Values

	withMockDefaultTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		method = r.Method
		body, _ := io.ReadAll(r.Body)
		receivedValues, _ = url.ParseQuery(string(body))

		respBody, _ := json.Marshal(map[string]interface{}{
			"access_token": "tok",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(respBody))),
		}, nil
	}))

	_, err := exchangeTokenWithURL(context.Background(), "https://example.test/token", "my-client", "my-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if method != "POST" {
		t.Errorf("method = %q, want POST", method)
	}
	if v := receivedValues.Get("grant_type"); v != "client_credentials" {
		t.Errorf("grant_type = %q, want client_credentials", v)
	}
	if v := receivedValues.Get("client_id"); v != "my-client" {
		t.Errorf("client_id = %q, want my-client", v)
	}
	if v := receivedValues.Get("scope"); v != "searchadsorg" {
		t.Errorf("scope = %q, want searchadsorg", v)
	}
}
