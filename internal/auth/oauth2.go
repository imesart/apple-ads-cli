package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/imesart/apple-ads-cli/internal/types"
)

const tokenURL = "https://appleid.apple.com/auth/oauth2/token"

// httpClient is used for token exchange requests instead of http.DefaultClient
// so that a reasonable timeout is enforced.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// SetHTTPClientTimeout updates the timeout used for token exchange requests.
func SetHTTPClientTimeout(timeout time.Duration) {
	httpClient.Timeout = timeout
}

// ExchangeToken exchanges a JWT client secret for an access token.
func ExchangeToken(ctx context.Context, clientID, clientSecret string) (*types.AccessToken, error) {
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {"searchadsorg"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
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
