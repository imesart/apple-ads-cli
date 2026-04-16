package types

import "time"

// AccessToken represents an OAuth2 access token from Apple ID.
type AccessToken struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"` // seconds

	// ObtainedAt is set locally when the token is received.
	ObtainedAt time.Time `json:"obtained_at"`
}

// IsExpired returns true if the token has expired or will expire within 60 seconds.
func (t *AccessToken) IsExpired() bool {
	if t == nil {
		return true
	}
	expiry := t.ObtainedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
	return time.Now().After(expiry.Add(-60 * time.Second))
}
