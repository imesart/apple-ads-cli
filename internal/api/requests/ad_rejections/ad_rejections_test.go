package ad_rejections

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestAllRequests(t *testing.T) {
	body := json.RawMessage(`{}`)

	tests := []struct {
		name string
		req  interface {
			Method() string
			Path() string
			Body() any
			Query() url.Values
		}
		method  string
		path    string
		hasBody bool
	}{
		{
			name:   "Get",
			req:    GetRequest{ID: "42"},
			method: "GET",
			path:   "/product-page-reasons/42",
		},
		{
			name:    "Find",
			req:     FindRequest{RawBody: body},
			method:  "POST",
			path:    "/product-page-reasons/find",
			hasBody: true,
		},
		{
			name:    "FindAssets",
			req:     FindAssetsRequest{AdamID: "789", RawBody: body},
			method:  "POST",
			path:    "/apps/789/assets/find",
			hasBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.Method(); got != tt.method {
				t.Errorf("Method() = %q, want %q", got, tt.method)
			}
			if got := tt.req.Path(); got != tt.path {
				t.Errorf("Path() = %q, want %q", got, tt.path)
			}
			if tt.hasBody && tt.req.Body() == nil {
				t.Error("Body() is nil, want non-nil")
			}
			if !tt.hasBody && tt.req.Body() != nil {
				t.Errorf("Body() = %v, want nil", tt.req.Body())
			}
			if tt.req.Query() != nil {
				t.Errorf("Query() = %v, want nil", tt.req.Query())
			}
		})
	}
}

func TestGetRequest_PathIncludesID(t *testing.T) {
	req := GetRequest{ID: "abc123"}
	if got := req.Path(); got != "/product-page-reasons/abc123" {
		t.Errorf("Path() = %q, want %q", got, "/product-page-reasons/abc123")
	}
}

func TestFindAssetsRequest_PathIncludesAdamID(t *testing.T) {
	req := FindAssetsRequest{AdamID: "555"}
	if got := req.Path(); got != "/apps/555/assets/find" {
		t.Errorf("Path() = %q, want %q", got, "/apps/555/assets/find")
	}
}
