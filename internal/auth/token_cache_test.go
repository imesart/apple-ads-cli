package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/imesart/apple-ads-cli/internal/types"
)

// writeTestKey creates a PEM EC key file in dir and returns its path.
func writeTestKey(t *testing.T, dir string) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshalling key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	})
	keyPath := filepath.Join(dir, "test_key.pem")
	if err := os.WriteFile(keyPath, pemBytes, 0o600); err != nil {
		t.Fatalf("writing key file: %v", err)
	}
	return keyPath
}

// newTestOAuthServer stubs the default transport to return a valid access token response.
// The returned cleanup function restores http.DefaultTransport.
func newTestOAuthServer(t *testing.T) func() {
	t.Helper()
	respBody, err := json.Marshal(map[string]interface{}{
		"access_token": "fresh-token-from-server",
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(respBody))),
			Request:    req,
		}, nil
	})

	return func() {
		http.DefaultTransport = origTransport
	}
}

func TestTokenStore_GetToken_Fresh(t *testing.T) {
	dir := t.TempDir()
	keyPath := writeTestKey(t, dir)
	cachePath := filepath.Join(dir, "token.json")

	cleanup := newTestOAuthServer(t)
	defer cleanup()

	store := NewTokenStore("team-id", "client-id", "key-id", keyPath, cachePath)

	tok, err := store.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if tok != "fresh-token-from-server" {
		t.Errorf("GetToken() = %q, want %q", tok, "fresh-token-from-server")
	}

	// Token cache file should exist after fetching
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("cache file should have been created")
	}
}

func TestTokenStore_GetToken_Cached(t *testing.T) {
	dir := t.TempDir()
	keyPath := writeTestKey(t, dir)
	cachePath := filepath.Join(dir, "token.json")

	store := NewTokenStore("team-id", "client-id", "key-id", keyPath, cachePath)

	// Directly set a valid token in the store (bypasses network)
	store.token = &types.AccessToken{
		Token:      "cached-token",
		TokenType:  "Bearer",
		ExpiresIn:  3600,
		ObtainedAt: time.Now(),
	}

	tok, err := store.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if tok != "cached-token" {
		t.Errorf("GetToken() = %q, want %q (should return cached)", tok, "cached-token")
	}

	// Call again to confirm no refresh happens
	tok2, err := store.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() second call error: %v", err)
	}
	if tok2 != "cached-token" {
		t.Errorf("GetToken() second call = %q, want %q", tok2, "cached-token")
	}
}

func TestTokenStore_Invalidate(t *testing.T) {
	dir := t.TempDir()
	keyPath := writeTestKey(t, dir)
	cachePath := filepath.Join(dir, "token.json")

	cleanup := newTestOAuthServer(t)
	defer cleanup()

	store := NewTokenStore("team-id", "client-id", "key-id", keyPath, cachePath)

	// Set a cached token
	store.token = &types.AccessToken{
		Token:      "old-token",
		TokenType:  "Bearer",
		ExpiresIn:  3600,
		ObtainedAt: time.Now(),
	}

	// Confirm cached token is returned
	tok, err := store.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() before invalidate error: %v", err)
	}
	if tok != "old-token" {
		t.Errorf("GetToken() = %q, want %q before invalidation", tok, "old-token")
	}

	// Invalidate
	store.Invalidate()

	// Next GetToken should fetch fresh token from server
	tok, err = store.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() after invalidate error: %v", err)
	}
	if tok != "fresh-token-from-server" {
		t.Errorf("GetToken() after invalidate = %q, want %q", tok, "fresh-token-from-server")
	}
}

func TestTokenStore_CacheFile(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "subdir", "token.json")

	// Write a token to the cache file via saveCache
	store := NewTokenStore("team", "client", "key", "/unused", cachePath)
	token := &types.AccessToken{
		Token:      "persisted-token",
		TokenType:  "Bearer",
		ExpiresIn:  7200,
		ObtainedAt: time.Now(),
	}

	if err := store.saveCache(token); err != nil {
		t.Fatalf("saveCache() error: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}

	// Read back via loadCache
	loaded, err := store.loadCache()
	if err != nil {
		t.Fatalf("loadCache() error: %v", err)
	}
	if loaded.Token != "persisted-token" {
		t.Errorf("loaded Token = %q, want %q", loaded.Token, "persisted-token")
	}
	if loaded.TokenType != "Bearer" {
		t.Errorf("loaded TokenType = %q, want %q", loaded.TokenType, "Bearer")
	}
	if loaded.ExpiresIn != 7200 {
		t.Errorf("loaded ExpiresIn = %d, want %d", loaded.ExpiresIn, 7200)
	}

	// Verify a new TokenStore can load from the same cache file
	store2 := NewTokenStore("team", "client", "key", "/unused", cachePath)
	loaded2, err := store2.loadCache()
	if err != nil {
		t.Fatalf("loadCache() from new store error: %v", err)
	}
	if loaded2.Token != "persisted-token" {
		t.Errorf("loaded2 Token = %q, want %q", loaded2.Token, "persisted-token")
	}

	// Verify the cached file content is valid JSON
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("reading cache file: %v", err)
	}
	var parsedToken types.AccessToken
	if err := json.Unmarshal(data, &parsedToken); err != nil {
		t.Fatalf("cache file contains invalid JSON: %v", err)
	}
	if parsedToken.Token != "persisted-token" {
		t.Errorf("parsed Token = %q, want %q", parsedToken.Token, "persisted-token")
	}
}

func TestTokenStore_CacheFile_LoadOnGetToken(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "token.json")

	// Pre-populate the cache file with a valid (non-expired) token
	cachedToken := &types.AccessToken{
		Token:      "disk-cached-token",
		TokenType:  "Bearer",
		ExpiresIn:  7200,
		ObtainedAt: time.Now(),
	}
	data, err := json.MarshalIndent(cachedToken, "", "  ")
	if err != nil {
		t.Fatalf("marshalling token: %v", err)
	}
	if err := os.WriteFile(cachePath, data, 0o600); err != nil {
		t.Fatalf("writing cache file: %v", err)
	}

	// Create a store with no in-memory token -- it should load from disk
	store := NewTokenStore("team", "client", "key", "/unused", cachePath)

	tok, err := store.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if tok != "disk-cached-token" {
		t.Errorf("GetToken() = %q, want %q (should load from disk cache)", tok, "disk-cached-token")
	}
}
