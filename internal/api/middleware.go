package api

import (
	"context"
	"fmt"
	"net/http"
)

// Middleware modifies an HTTP request before it is sent.
type Middleware func(r *http.Request) error

// InjectHost sets the request URL host and scheme.
func InjectHost(host string) Middleware {
	return func(r *http.Request) error {
		r.URL.Scheme = "https"
		r.URL.Host = host
		return nil
	}
}

// InjectAcceptHeaders sets the Accept and Content-Type headers to application/json.
func InjectAcceptHeaders() Middleware {
	return func(r *http.Request) error {
		r.Header.Set("Accept", "application/json")
		if r.Body != nil && r.Body != http.NoBody {
			r.Header.Set("Content-Type", "application/json")
		}
		return nil
	}
}

// InjectAuthorization adds a Bearer token from the provided token getter function.
func InjectAuthorization(getToken func(ctx context.Context) (string, error)) Middleware {
	return func(r *http.Request) error {
		token, err := getToken(r.Context())
		if err != nil {
			return fmt.Errorf("getting access token: %w", err)
		}
		r.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

// InjectOrgContext adds the X-AP-Context header with the organization ID.
func InjectOrgContext(orgID string) Middleware {
	return func(r *http.Request) error {
		if orgID != "" {
			r.Header.Set("X-AP-Context", fmt.Sprintf("orgId=%s", orgID))
		}
		return nil
	}
}
