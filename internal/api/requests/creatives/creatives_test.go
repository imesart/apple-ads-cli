package creatives

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
		method   string
		path     string
		hasBody  bool
		hasQuery bool
	}{
		{
			name:    "Create",
			req:     CreateRequest{RawBody: body},
			method:  "POST",
			path:    "/creatives",
			hasBody: true,
		},
		{
			name:   "Get",
			req:    GetRequest{CreativeID: "456"},
			method: "GET",
			path:   "/creatives/456",
		},
		{
			name:     "List",
			req:      ListRequest{Limit: 10, Offset: 5},
			method:   "GET",
			path:     "/creatives",
			hasQuery: true,
		},
		{
			name:    "Find",
			req:     FindRequest{RawBody: body},
			method:  "POST",
			path:    "/creatives/find",
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
			if tt.hasQuery {
				if tt.req.Query() == nil {
					t.Fatal("Query() is nil, want non-nil")
				}
			}
		})
	}
}

func TestListRequest_Pagination(t *testing.T) {
	req := ListRequest{Limit: 20, Offset: 40}
	q := req.Query()
	if got := q.Get("limit"); got != "20" {
		t.Errorf("limit = %q, want %q", got, "20")
	}
	if got := q.Get("offset"); got != "40" {
		t.Errorf("offset = %q, want %q", got, "40")
	}
}

func TestListRequest_ZeroValues(t *testing.T) {
	req := ListRequest{}
	q := req.Query()
	if q.Get("limit") != "" {
		t.Errorf("limit = %q, want empty", q.Get("limit"))
	}
	if q.Get("offset") != "" {
		t.Errorf("offset = %q, want empty", q.Get("offset"))
	}
}

func TestGetRequest_PathIncludesID(t *testing.T) {
	req := GetRequest{CreativeID: "xyz"}
	if got := req.Path(); got != "/creatives/xyz" {
		t.Errorf("Path() = %q, want %q", got, "/creatives/xyz")
	}
}
