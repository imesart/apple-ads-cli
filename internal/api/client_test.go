package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aclsReq "github.com/imesart/apple-ads-cli/internal/api/requests/acls"
	adRejectionsReq "github.com/imesart/apple-ads-cli/internal/api/requests/ad_rejections"
	adgroupsReq "github.com/imesart/apple-ads-cli/internal/api/requests/adgroups"
	adsReq "github.com/imesart/apple-ads-cli/internal/api/requests/ads"
	appsReq "github.com/imesart/apple-ads-cli/internal/api/requests/apps"
	budgetordersReq "github.com/imesart/apple-ads-cli/internal/api/requests/budgetorders"
	campaignsReq "github.com/imesart/apple-ads-cli/internal/api/requests/campaigns"
	creativesReq "github.com/imesart/apple-ads-cli/internal/api/requests/creatives"
	geoReq "github.com/imesart/apple-ads-cli/internal/api/requests/geo"
	impressionShareReq "github.com/imesart/apple-ads-cli/internal/api/requests/impression_share"
	keywordsReq "github.com/imesart/apple-ads-cli/internal/api/requests/keywords"
	negAdgroupReq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_adgroup"
	negCampaignReq "github.com/imesart/apple-ads-cli/internal/api/requests/negatives_campaign"
	productPagesReq "github.com/imesart/apple-ads-cli/internal/api/requests/product_pages"
	reportsReq "github.com/imesart/apple-ads-cli/internal/api/requests/reports"
)

// mockRequest implements the Request interface for testing.
type mockRequest struct {
	method string
	path   string
	body   any
	query  url.Values
}

func (r mockRequest) Method() string    { return r.method }
func (r mockRequest) Path() string      { return r.path }
func (r mockRequest) Body() any         { return r.body }
func (r mockRequest) Query() url.Values { return r.query }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestServer creates a test client backed by a mock RoundTripper.
func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	client := &Client{
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				rec := &responseRecorder{
					header: make(http.Header),
					body:   &strings.Builder{},
				}
				handler(rec, req)
				return rec.response(req), nil
			}),
		},
		baseURL:         "https://example.test",
		middlewares:     []Middleware{},
		mutatingLimiter: make(chan struct{}, 8),
	}
	return client
}

type responseRecorder struct {
	code   int
	header http.Header
	body   *strings.Builder
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.body.Write(data)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.code = statusCode
}

func (r *responseRecorder) response(req *http.Request) *http.Response {
	code := r.code
	if code == 0 {
		code = http.StatusOK
	}
	return &http.Response{
		StatusCode: code,
		Status:     http.StatusText(code),
		Header:     r.header.Clone(),
		Body:       io.NopCloser(strings.NewReader(r.body.String())),
		Request:    req,
	}
}

func TestClient_Do_GET_Success(t *testing.T) {
	type response struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"id": 123},
		}); err != nil {
			t.Fatalf("encoding response: %v", err)
		}
	})

	var result response
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
	}, &result)

	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
	if result.Data.ID != 123 {
		t.Errorf("Data.ID = %d, want 123", result.Data.ID)
	}
}

func TestClient_Do_POST_Success(t *testing.T) {
	type requestBody struct {
		Name string `json:"name"`
	}

	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}

		var got requestBody
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshalling body: %v", err)
		}
		if got.Name != "test campaign" {
			t.Errorf("body.Name = %q, want %q", got.Name, "test campaign")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data":{"id":456}}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	type response struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	var result response
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodPost,
		path:   "/campaigns",
		body:   requestBody{Name: "test campaign"},
	}, &result)

	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
	if result.Data.ID != 456 {
		t.Errorf("Data.ID = %d, want 456", result.Data.ID)
	}
}

func TestClient_Do_QueryParams(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if got := q.Get("limit"); got != "10" {
			t.Errorf("query param limit = %q, want %q", got, "10")
		}
		if got := q.Get("offset"); got != "20" {
			t.Errorf("query param offset = %q, want %q", got, "20")
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	var result map[string]any
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
		query: url.Values{
			"limit":  []string{"10"},
			"offset": []string{"20"},
		},
	}, &result)

	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
}

func TestClient_Do_Error_400(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"error":{"errors":[{"message":"bad request","messageCode":"INVALID"}]}}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	var result map[string]any
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
	}, &result)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(apiErr.Errors))
	}
	if derefStr(apiErr.Errors[0].Message) != "bad request" {
		t.Errorf("Errors[0].Message = %q, want %q", derefStr(apiErr.Errors[0].Message), "bad request")
	}
	if derefStr(apiErr.Errors[0].MessageCode) != "INVALID" {
		t.Errorf("Errors[0].MessageCode = %q, want %q", derefStr(apiErr.Errors[0].MessageCode), "INVALID")
	}
}

func TestClient_Do_Error_401(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"error":{"errors":[{"message":"Unauthorized"}]}}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
	}, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error should wrap ErrUnauthorized, got: %v", err)
	}
}

func TestClient_Do_RetriesOnceAfter401(t *testing.T) {
	attempts := 0
	tokenIndex := 0
	invalidateCalls := 0
	tokens := []string{"stale-token", "fresh-token"}

	client := &Client{
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if got := req.Header.Get("Authorization"); got != "Bearer "+tokens[tokenIndex] {
					t.Fatalf("Authorization = %q, want %q", got, "Bearer "+tokens[tokenIndex])
				}

				rec := &responseRecorder{
					header: make(http.Header),
					body:   &strings.Builder{},
				}
				if attempts == 1 {
					rec.WriteHeader(http.StatusUnauthorized)
					_, _ = rec.Write([]byte(`{"error":{"errors":[{"message":"Unauthorized"}]}}`))
				} else {
					rec.Header().Set("Content-Type", "application/json")
					rec.WriteHeader(http.StatusOK)
					_, _ = rec.Write([]byte(`{"data":{"id":123}}`))
				}
				return rec.response(req), nil
			}),
		},
		baseURL: "https://example.test",
		getToken: func(context.Context) (string, error) {
			return tokens[tokenIndex], nil
		},
		invalidateToken: func() {
			invalidateCalls++
			tokenIndex = 1
		},
		middlewares: []Middleware{
			InjectAuthorization(func(context.Context) (string, error) {
				return tokens[tokenIndex], nil
			}),
		},
		mutatingLimiter: make(chan struct{}, 8),
	}

	var result struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
	}, &result)

	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if invalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", invalidateCalls)
	}
	if result.Data.ID != 123 {
		t.Fatalf("Data.ID = %d, want 123", result.Data.ID)
	}
}

func TestClient_Do_StopsAfterSecond401(t *testing.T) {
	attempts := 0
	invalidateCalls := 0

	client := &Client{
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				rec := &responseRecorder{
					header: make(http.Header),
					body:   &strings.Builder{},
				}
				rec.WriteHeader(http.StatusUnauthorized)
				_, _ = rec.Write([]byte(`{"error":{"errors":[{"message":"Unauthorized"}]}}`))
				return rec.response(req), nil
			}),
		},
		baseURL: "https://example.test",
		invalidateToken: func() {
			invalidateCalls++
		},
		mutatingLimiter: make(chan struct{}, 8),
	}

	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
	}, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("error should wrap ErrUnauthorized, got: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if invalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", invalidateCalls)
	}
}

func TestClient_Do_Error_429(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		if _, err := w.Write([]byte(`{"error":{"errors":[{"message":"Too Many Requests"}]}}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/campaigns",
	}, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrRateLimit) {
		t.Errorf("error should wrap ErrRateLimit, got: %v", err)
	}
}

func TestClient_Do_Middleware(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Custom-Header"); got != "test-value" {
			t.Errorf("X-Custom-Header = %q, want %q", got, "test-value")
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	// Add a custom middleware that sets a header.
	client.middlewares = append(client.middlewares, func(r *http.Request) error {
		r.Header.Set("X-Custom-Header", "test-value")
		return nil
	})

	var result map[string]any
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/test",
	}, &result)

	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
}

func TestClient_Do_NilTarget(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return empty body.
	})

	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/test",
	}, nil)

	if err != nil {
		t.Fatalf("Do() with nil target returned error: %v", err)
	}
}

func TestClient_Do_NilTarget_WithBody(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data":{"id":1}}`)); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	// nil target should succeed even if server returns a body.
	err := client.Do(context.Background(), mockRequest{
		method: http.MethodDelete,
		path:   "/campaigns/1",
	}, nil)

	if err != nil {
		t.Fatalf("Do() with nil target returned error: %v", err)
	}
}

func TestClient_Do_MiddlewareError(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when middleware errors")
	})

	mwErr := errors.New("middleware failure")
	client.middlewares = append(client.middlewares, func(r *http.Request) error {
		return mwErr
	})

	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/test",
	}, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mwErr) {
		t.Errorf("error should wrap middleware error, got: %v", err)
	}
}

func TestClient_Do_ErrorFallbackRawBody(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Fatalf("writing response: %v", err)
		}
	})

	err := client.Do(context.Background(), mockRequest{
		method: http.MethodGet,
		path:   "/test",
	}, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	// Raw body should be captured as an ErrorItem message.
	if len(apiErr.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(apiErr.Errors))
	}
	if derefStr(apiErr.Errors[0].Message) != "Internal Server Error" {
		t.Errorf("Errors[0].Message = %q, want %q", derefStr(apiErr.Errors[0].Message), "Internal Server Error")
	}
}

func TestClient_DownloadToWriter_RetriesOnceAfter401(t *testing.T) {
	attempts := 0
	tokenIndex := 0
	invalidateCalls := 0
	tokens := []string{"stale-token", "fresh-token"}

	client := &Client{
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if got := req.Header.Get("Authorization"); got != "Bearer "+tokens[tokenIndex] {
					t.Fatalf("Authorization = %q, want %q", got, "Bearer "+tokens[tokenIndex])
				}

				rec := &responseRecorder{
					header: make(http.Header),
					body:   &strings.Builder{},
				}
				if attempts == 1 {
					rec.WriteHeader(http.StatusUnauthorized)
					_, _ = rec.Write([]byte("Unauthorized"))
				} else {
					rec.WriteHeader(http.StatusOK)
					_, _ = rec.Write([]byte("report,data\n1,2\n"))
				}
				return rec.response(req), nil
			}),
		},
		getToken: func(context.Context) (string, error) {
			return tokens[tokenIndex], nil
		},
		invalidateToken: func() {
			invalidateCalls++
			tokenIndex = 1
		},
		orgID: "12345",
	}

	var buf bytes.Buffer
	err := client.DownloadToWriter(context.Background(), "https://api.searchads.apple.com/report.csv", &buf)
	if err != nil {
		t.Fatalf("DownloadToWriter() returned error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if invalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", invalidateCalls)
	}
	if got := buf.String(); got != "report,data\n1,2\n" {
		t.Fatalf("downloaded body = %q, want %q", got, "report,data\n1,2\n")
	}
}

func TestClient_DownloadToWriter_StopsAfterSecond401(t *testing.T) {
	attempts := 0
	invalidateCalls := 0

	client := &Client{
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				rec := &responseRecorder{
					header: make(http.Header),
					body:   &strings.Builder{},
				}
				rec.WriteHeader(http.StatusUnauthorized)
				_, _ = rec.Write([]byte("Unauthorized"))
				return rec.response(req), nil
			}),
		},
		getToken: func(context.Context) (string, error) {
			return "stale-token", nil
		},
		invalidateToken: func() {
			invalidateCalls++
		},
	}

	var buf bytes.Buffer
	err := client.DownloadToWriter(context.Background(), "https://api.searchads.apple.com/report.csv", &buf)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("error should wrap ErrUnauthorized, got: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if invalidateCalls != 1 {
		t.Fatalf("invalidateCalls = %d, want 1", invalidateCalls)
	}
	if buf.Len() != 0 {
		t.Fatalf("downloaded body length = %d, want 0", buf.Len())
	}
}

func TestClient_DownloadToFile_PreservesExistingFileOnFailure(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("download failed"))
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "report.csv")
	if err := os.WriteFile(path, []byte("existing,data\n1,2\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}

	err := client.DownloadToFile(context.Background(), "https://example.test/report.csv", path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q): %v", path, readErr)
	}
	if got := string(data); got != "existing,data\n1,2\n" {
		t.Fatalf("file contents = %q, want %q", got, "existing,data\n1,2\n")
	}

	matches, globErr := filepath.Glob(filepath.Join(dir, "report.csv.*.tmp"))
	if globErr != nil {
		t.Fatalf("Glob() error: %v", globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}

func TestClient_DownloadToFile_ReplacesDestinationOnSuccess(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("new,data\n3,4\n"))
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "report.csv")
	if err := os.WriteFile(path, []byte("old,data\n1,2\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}

	if err := client.DownloadToFile(context.Background(), "https://example.test/report.csv", path); err != nil {
		t.Fatalf("DownloadToFile() returned error: %v", err)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q): %v", path, readErr)
	}
	if got := string(data); got != "new,data\n3,4\n" {
		t.Fatalf("file contents = %q, want %q", got, "new,data\n3,4\n")
	}

	matches, globErr := filepath.Glob(filepath.Join(dir, "report.csv.*.tmp"))
	if globErr != nil {
		t.Fatalf("Glob() error: %v", globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}

// TestBuildRequest_NoDoubledBasePath verifies that the full URL produced by
// buildRequest does not duplicate the /api/v5 prefix. This guards against the
// bug where baseURL already contains /api/v5 and the request Path() also
// starts with /api/v5, yielding /api/v5/api/v5/…
func TestBuildRequest_NoDoubledBasePath(t *testing.T) {
	client := &Client{
		httpClient:      http.DefaultClient,
		baseURL:         "https://" + baseHost + basePath,
		mutatingLimiter: make(chan struct{}, 8),
	}

	tests := []struct {
		name    string
		req     Request
		wantURL string
	}{
		{
			name:    "simple resource path",
			req:     mockRequest{method: "GET", path: "/campaigns"},
			wantURL: "https://api.searchads.apple.com/api/v5/campaigns",
		},
		{
			name:    "nested resource path",
			req:     mockRequest{method: "GET", path: "/campaigns/123/adgroups/456/targetingkeywords"},
			wantURL: "https://api.searchads.apple.com/api/v5/campaigns/123/adgroups/456/targetingkeywords",
		},
		{
			name:    "find endpoint",
			req:     mockRequest{method: "POST", path: "/campaigns/find"},
			wantURL: "https://api.searchads.apple.com/api/v5/campaigns/find",
		},
		{
			name:    "reports endpoint",
			req:     mockRequest{method: "POST", path: "/reports/campaigns"},
			wantURL: "https://api.searchads.apple.com/api/v5/reports/campaigns",
		},
		{
			name: "with query params",
			req: mockRequest{
				method: "GET",
				path:   "/campaigns/123/adgroups",
				query:  url.Values{"limit": []string{"20"}},
			},
			wantURL: "https://api.searchads.apple.com/api/v5/campaigns/123/adgroups?limit=20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq, err := client.buildRequest(context.Background(), tt.req)
			if err != nil {
				t.Fatalf("buildRequest() error: %v", err)
			}
			if got := httpReq.URL.String(); got != tt.wantURL {
				t.Errorf("URL = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

// TestRequestPaths_NoAPIVersionPrefix ensures that all real request Path()
// methods return paths relative to the base URL, without the /api/v5 prefix.
// If a Path() starts with /api/v5, the URL will be doubled when combined
// with baseURL which already contains /api/v5.
func TestRequestPaths_NoAPIVersionPrefix(t *testing.T) {
	requests := []struct {
		name string
		req  Request
	}{
		// Campaigns
		{"campaigns.List", campaignsReq.ListRequest{}},
		{"campaigns.Get", campaignsReq.GetRequest{CampaignID: "1"}},
		{"campaigns.Create", campaignsReq.CreateRequest{}},
		{"campaigns.Find", campaignsReq.FindRequest{}},
		{"campaigns.Update", campaignsReq.UpdateRequest{CampaignID: "1"}},
		{"campaigns.Delete", campaignsReq.DeleteRequest{CampaignID: "1"}},
		// Ad Groups
		{"adgroups.List", adgroupsReq.ListRequest{CampaignID: "1"}},
		{"adgroups.Get", adgroupsReq.GetRequest{CampaignID: "1", AdGroupID: "2"}},
		{"adgroups.Create", adgroupsReq.CreateRequest{CampaignID: "1"}},
		{"adgroups.Find", adgroupsReq.FindRequest{CampaignID: "1"}},
		{"adgroups.FindAll", adgroupsReq.FindAllRequest{}},
		{"adgroups.Update", adgroupsReq.UpdateRequest{CampaignID: "1", AdGroupID: "2"}},
		{"adgroups.Delete", adgroupsReq.DeleteRequest{CampaignID: "1", AdGroupID: "2"}},
		// Keywords
		{"keywords.List", keywordsReq.ListRequest{CampaignID: "1", AdGroupID: "2"}},
		{"keywords.Get", keywordsReq.GetRequest{CampaignID: "1", AdGroupID: "2", KeywordID: "3"}},
		{"keywords.Create", keywordsReq.CreateRequest{CampaignID: "1", AdGroupID: "2"}},
		{"keywords.Find", keywordsReq.FindRequest{CampaignID: "1"}},
		{"keywords.Update", keywordsReq.UpdateRequest{CampaignID: "1", AdGroupID: "2"}},
		{"keywords.DeleteOne", keywordsReq.DeleteOneRequest{CampaignID: "1", AdGroupID: "2", KeywordID: "3"}},
		{"keywords.DeleteBulk", keywordsReq.DeleteBulkRequest{CampaignID: "1", AdGroupID: "2"}},
		// Ads
		{"ads.List", adsReq.ListRequest{CampaignID: "1", AdGroupID: "2"}},
		{"ads.Get", adsReq.GetRequest{CampaignID: "1", AdGroupID: "2", AdID: "3"}},
		{"ads.Create", adsReq.CreateRequest{CampaignID: "1", AdGroupID: "2"}},
		{"ads.Find", adsReq.FindRequest{CampaignID: "1", AdGroupID: "2"}},
		{"ads.FindAll", adsReq.FindAllRequest{}},
		{"ads.Update", adsReq.UpdateRequest{CampaignID: "1", AdGroupID: "2", AdID: "3"}},
		{"ads.Delete", adsReq.DeleteRequest{CampaignID: "1", AdGroupID: "2", AdID: "3"}},
		// Reports
		{"reports.Campaigns", reportsReq.CampaignsRequest{}},
		{"reports.AdGroups", reportsReq.AdGroupsRequest{CampaignID: "1"}},
		{"reports.Keywords", reportsReq.KeywordsRequest{CampaignID: "1"}},
		{"reports.SearchTerms", reportsReq.SearchTermsRequest{CampaignID: "1"}},
		{"reports.Ads", reportsReq.AdsRequest{CampaignID: "1"}},
		// ACLs
		{"acls.List", aclsReq.ListRequest{}},
		{"acls.Me", aclsReq.MeRequest{}},
		// Apps
		{"apps.Search", appsReq.SearchRequest{}},
		{"apps.Details", appsReq.DetailsRequest{AdamID: "1"}},
		{"apps.Localized", appsReq.LocalizedRequest{AdamID: "1"}},
		{"apps.Eligibility", appsReq.EligibilityRequest{}},
		// Creatives
		{"creatives.List", creativesReq.ListRequest{}},
		{"creatives.Get", creativesReq.GetRequest{CreativeID: "1"}},
		{"creatives.Create", creativesReq.CreateRequest{}},
		{"creatives.Find", creativesReq.FindRequest{}},
		// Budget Orders
		{"budgetorders.List", budgetordersReq.ListRequest{}},
		{"budgetorders.Get", budgetordersReq.GetRequest{BudgetOrderID: "1"}},
		{"budgetorders.Create", budgetordersReq.CreateRequest{}},
		{"budgetorders.Update", budgetordersReq.UpdateRequest{BudgetOrderID: "1"}},
		// Negative Keywords (Campaign)
		{"negCampaign.List", negCampaignReq.ListRequest{CampaignID: "1"}},
		{"negCampaign.Get", negCampaignReq.GetRequest{CampaignID: "1", KeywordID: "2"}},
		{"negCampaign.Create", negCampaignReq.CreateRequest{CampaignID: "1"}},
		{"negCampaign.Find", negCampaignReq.FindRequest{CampaignID: "1"}},
		{"negCampaign.Update", negCampaignReq.UpdateRequest{CampaignID: "1"}},
		{"negCampaign.Delete", negCampaignReq.DeleteBulkRequest{CampaignID: "1"}},
		// Negative Keywords (Ad Group)
		{"negAdgroup.List", negAdgroupReq.ListRequest{CampaignID: "1", AdGroupID: "2"}},
		{"negAdgroup.Get", negAdgroupReq.GetRequest{CampaignID: "1", AdGroupID: "2", KeywordID: "3"}},
		{"negAdgroup.Create", negAdgroupReq.CreateRequest{CampaignID: "1", AdGroupID: "2"}},
		{"negAdgroup.Find", negAdgroupReq.FindRequest{CampaignID: "1"}},
		{"negAdgroup.Update", negAdgroupReq.UpdateRequest{CampaignID: "1", AdGroupID: "2"}},
		{"negAdgroup.Delete", negAdgroupReq.DeleteBulkRequest{CampaignID: "1", AdGroupID: "2"}},
		// Ad Rejections
		{"adRejections.Find", adRejectionsReq.FindRequest{}},
		{"adRejections.Get", adRejectionsReq.GetRequest{ID: "1"}},
		{"adRejections.FindAssets", adRejectionsReq.FindAssetsRequest{AdamID: "1"}},
		// Product Pages
		{"productPages.List", productPagesReq.ListRequest{AdamID: "1"}},
		{"productPages.Get", productPagesReq.GetRequest{AdamID: "1", ProductPageID: "2"}},
		{"productPages.Locales", productPagesReq.LocalesRequest{AdamID: "1", ProductPageID: "2"}},
		{"productPages.Countries", productPagesReq.CountriesRequest{}},
		{"productPages.Devices", productPagesReq.DevicesRequest{}},
		// Geo
		{"geo.Search", geoReq.SearchRequest{}},
		{"geo.Get", geoReq.GetRequest{ID: "US", Entity: "Country"}},
		// Impression Share
		{"impressionShare.List", impressionShareReq.ListRequest{}},
		{"impressionShare.Get", impressionShareReq.GetRequest{ReportID: "1"}},
		{"impressionShare.Create", impressionShareReq.CreateRequest{}},
	}

	for _, tt := range requests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.req.Path()
			if len(path) < 1 || path[0] != '/' {
				t.Errorf("Path() = %q, want path starting with /", path)
			}
			if strings.HasPrefix(path, basePath) {
				t.Errorf("Path() = %q, must not start with %q (baseURL already contains it)", path, basePath)
			}
		})
	}
}

func TestIsMutating(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{http.MethodGet, false},
		{http.MethodHead, false},
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodDelete, true},
		{http.MethodPatch, true},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			if got := isMutating(tt.method); got != tt.want {
				t.Errorf("isMutating(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}
