package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imesart/apple-ads-cli/internal/types"
)

const (
	// baseHost is the Apple Search Ads API host.
	baseHost = "api.searchads.apple.com"

	// basePath is the API version prefix.
	basePath = "/api/v5"

	// maxMutating is the concurrency limit for mutating requests (POST/PUT/DELETE).
	maxMutating = 8

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
)

// Client handles HTTP communication with the Apple Search Ads API.
type Client struct {
	httpClient      *http.Client
	baseURL         string
	middlewares     []Middleware
	getToken        func(context.Context) (string, error)
	invalidateToken func()
	orgID           string
	verbose         bool

	// mutatingLimiter limits concurrent mutating requests.
	mutatingLimiter chan struct{}
}

// NewClient creates a new API client with the standard middleware chain.
// It accepts either:
//   - getToken, orgID, verbose
//   - getToken, invalidateToken, orgID, verbose
func NewClient(getToken func(context.Context) (string, error), args ...any) *Client {
	var (
		invalidateToken func()
		orgID           string
		verbose         bool
	)

	switch len(args) {
	case 2:
		orgID, _ = args[0].(string)
		verbose, _ = args[1].(bool)
	case 3:
		if args[0] != nil {
			invalidateToken, _ = args[0].(func())
		}
		orgID, _ = args[1].(string)
		verbose, _ = args[2].(bool)
	default:
		panic("NewClient requires getToken plus either (orgID, verbose) or (invalidateToken, orgID, verbose)")
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:         "https://" + baseHost + basePath,
		getToken:        getToken,
		invalidateToken: invalidateToken,
		orgID:           orgID,
		verbose:         verbose,
		middlewares: []Middleware{
			InjectHost(baseHost),
			InjectAcceptHeaders(),
			InjectAuthorization(getToken),
			InjectOrgContext(orgID),
		},
		mutatingLimiter: make(chan struct{}, maxMutating),
	}
	return c
}

// SetHTTPClientForTesting replaces the underlying HTTP client.
// Intended for tests that need to stub transport behavior.
func (c *Client) SetHTTPClientForTesting(httpClient *http.Client) {
	c.httpClient = httpClient
}

// SetTimeout updates the HTTP client timeout.
func (c *Client) SetTimeout(timeout time.Duration) {
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	c.httpClient.Timeout = timeout
}

// Do executes an API request and decodes the JSON response into target.
// If target is nil, the response body is discarded.
func (c *Client) Do(ctx context.Context, req Request, target any) error {
	// Limit concurrent mutating requests
	if isMutating(req.Method()) {
		c.mutatingLimiter <- struct{}{}
		defer func() { <-c.mutatingLimiter }()
	}

	_, err := WithAuthRetry(ctx, c.invalidateAuthToken, func() (struct{}, error) {
		httpReq, err := c.buildRequest(ctx, req)
		if err != nil {
			return struct{}{}, fmt.Errorf("building request: %w", err)
		}

		// Apply all middlewares for each attempt so auth can fetch a fresh token.
		for _, mw := range c.middlewares {
			if err := mw(httpReq); err != nil {
				return struct{}{}, fmt.Errorf("applying middleware: %w", err)
			}
		}

		if c.verbose {
			c.logRequest(httpReq)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return struct{}{}, fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()

		if c.verbose {
			c.logResponse(resp)
		}

		return struct{}{}, c.handleResponse(resp, target)
	})
	return err
}

func (c *Client) invalidateAuthToken() {
	if c.invalidateToken != nil {
		c.invalidateToken()
	}
}

// DoList executes a paginated list request.
// It is a convenience wrapper around Do for list endpoints.
func (c *Client) DoList(ctx context.Context, req Request, target any) error {
	return c.Do(ctx, req, target)
}

// DownloadToWriter fetches an absolute URL and copies the response body to w.
func (c *Client) DownloadToWriter(ctx context.Context, rawURL string, w io.Writer) error {
	_, err := WithAuthRetry(ctx, c.invalidateAuthToken, func() (struct{}, error) {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return struct{}{}, fmt.Errorf("creating download request: %w", err)
		}

		if strings.EqualFold(httpReq.URL.Hostname(), baseHost) {
			if c.getToken != nil {
				token, err := c.getToken(ctx)
				if err != nil {
					return struct{}{}, fmt.Errorf("getting access token: %w", err)
				}
				httpReq.Header.Set("Authorization", "Bearer "+token)
			}
			if c.orgID != "" {
				httpReq.Header.Set("X-AP-Context", fmt.Sprintf("orgId=%s", c.orgID))
			}
		}

		if c.verbose {
			c.logRequest(httpReq)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return struct{}{}, fmt.Errorf("executing download request: %w", err)
		}
		defer resp.Body.Close()

		if c.verbose {
			c.logResponse(resp)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return struct{}{}, fmt.Errorf("reading download error body: %w", readErr)
			}
			apiErr := &APIError{StatusCode: resp.StatusCode}
			trimmed := strings.TrimSpace(string(bodyBytes))
			if trimmed != "" {
				apiErr.Errors = []types.ErrorItem{{Message: &trimmed}}
			}
			return struct{}{}, apiErr
		}

		if _, err := io.Copy(w, resp.Body); err != nil {
			return struct{}{}, fmt.Errorf("writing download output: %w", err)
		}
		return struct{}{}, nil
	})
	return err
}

// DownloadToFile fetches an absolute URL and writes the response body to path.
func (c *Client) DownloadToFile(ctx context.Context, rawURL string, path string) error {
	dir := filepath.Dir(path)
	pattern := filepath.Base(path) + ".*.tmp"

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return fmt.Errorf("creating temp download file: %w", err)
	}
	tempPath := file.Name()
	keepTemp := false
	defer func() {
		_ = file.Close()
		if !keepTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := c.DownloadToWriter(ctx, rawURL, file); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing temp download file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replacing download file: %w", err)
	}

	keepTemp = true
	return nil
}

// buildRequest creates an *http.Request from the Request interface.
func (c *Client) buildRequest(ctx context.Context, req Request) (*http.Request, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + req.Path())
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	// Merge query parameters
	if q := req.Query(); q != nil {
		u.RawQuery = q.Encode()
	}

	// Build body
	var body io.Reader
	if req.Body() != nil {
		data, err := json.Marshal(req.Body())
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		body = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method(), u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	return httpReq, nil
}

// handleResponse checks the status code and decodes the response body.
func (c *Client) handleResponse(resp *http.Response, target any) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	// Success range
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if target == nil || len(bodyBytes) == 0 {
			return nil
		}
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		return nil
	}

	// Error response -- try to parse the structured error
	apiErr := &APIError{StatusCode: resp.StatusCode}

	var errBody types.ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errBody); err == nil && len(errBody.Error.Errors) > 0 {
		apiErr.Errors = errBody.Error.Errors
	} else {
		// Fallback: use the raw body as the error message
		trimmed := strings.TrimSpace(string(bodyBytes))
		if trimmed != "" {
			apiErr.Errors = []types.ErrorItem{{Message: &trimmed}}
		}
	}

	return apiErr
}

// logRequest writes request details to stderr in verbose mode.
func (c *Client) logRequest(req *http.Request) {
	fmt.Fprintf(os.Stderr, ">>> %s %s\n", req.Method, req.URL.String())
	for key, vals := range req.Header {
		for _, v := range vals {
			if strings.EqualFold(key, "Authorization") {
				v = "Bearer <redacted>"
			}
			fmt.Fprintf(os.Stderr, ">>> %s: %s\n", key, v)
		}
	}
	fmt.Fprintln(os.Stderr, ">>>")
}

// logResponse writes response details to stderr in verbose mode.
func (c *Client) logResponse(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "<<< %s\n", resp.Status)
	for key, vals := range resp.Header {
		for _, v := range vals {
			fmt.Fprintf(os.Stderr, "<<< %s: %s\n", key, v)
		}
	}
	fmt.Fprintln(os.Stderr, "<<<")
}

// isMutating returns true if the HTTP method modifies resources.
func isMutating(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	default:
		return false
	}
}
