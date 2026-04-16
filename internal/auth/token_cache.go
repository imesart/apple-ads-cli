package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/imesart/apple-ads-cli/internal/types"
)

// TokenStore manages access token lifecycle: caching, refresh, and invalidation.
type TokenStore struct {
	mu        sync.Mutex
	token     *types.AccessToken
	cachePath string

	// Credentials for refresh
	teamID         string
	clientID       string
	keyID          string
	privateKeyPath string
}

func NewTokenStore(teamID, clientID, keyID, privateKeyPath, cachePath string) *TokenStore {
	return &TokenStore{
		teamID:         teamID,
		clientID:       clientID,
		keyID:          keyID,
		privateKeyPath: privateKeyPath,
		cachePath:      cachePath,
	}
}

// GetToken returns a valid access token, refreshing if necessary.
func (s *TokenStore) GetToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != nil && !s.token.IsExpired() {
		return s.token.Token, nil
	}

	// Try loading from cache
	if s.token == nil {
		cached, err := s.loadCache()
		if err == nil && cached != nil && !cached.IsExpired() {
			s.token = cached
			return s.token.Token, nil
		}
	}

	// Refresh
	return s.refreshLocked(ctx)
}

// Invalidate forces a token refresh on next GetToken call.
func (s *TokenStore) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = nil
}

// RefreshToken forces a new token exchange.
func (s *TokenStore) RefreshToken(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.refreshLocked(ctx)
	return err
}

func (s *TokenStore) refreshLocked(ctx context.Context) (string, error) {
	jwtToken, err := BuildJWT(s.teamID, s.clientID, s.keyID, s.privateKeyPath, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("building JWT: %w", err)
	}

	token, err := ExchangeToken(ctx, s.clientID, jwtToken)
	if err != nil {
		return "", fmt.Errorf("exchanging token: %w", err)
	}

	s.token = token
	_ = s.saveCache(token) // best effort
	return token.Token, nil
}

func (s *TokenStore) loadCache() (*types.AccessToken, error) {
	data, err := os.ReadFile(s.cachePath)
	if err != nil {
		return nil, err
	}
	var token types.AccessToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *TokenStore) saveCache(token *types.AccessToken) error {
	dir := filepath.Dir(s.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cachePath, data, 0o600)
}
