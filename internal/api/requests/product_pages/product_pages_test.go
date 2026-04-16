package product_pages

import (
	"net/url"
	"testing"
)

func TestAllRequests(t *testing.T) {
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
		hasQuery bool
	}{
		{
			name:   "Get",
			req:    GetRequest{AdamID: "789", ProductPageID: "pp-1"},
			method: "GET",
			path:   "/apps/789/product-pages/pp-1",
		},
		{
			name:     "List",
			req:      ListRequest{AdamID: "789", Limit: 10, Offset: 5},
			method:   "GET",
			path:     "/apps/789/product-pages",
			hasQuery: true,
		},
		{
			name:   "Locales",
			req:    LocalesRequest{AdamID: "789", ProductPageID: "pp-1"},
			method: "GET",
			path:   "/apps/789/product-pages/pp-1/locale-details",
		},
		{
			name:   "Countries",
			req:    CountriesRequest{},
			method: "GET",
			path:   "/countries-or-regions",
		},
		{
			name:   "Devices",
			req:    DevicesRequest{},
			method: "GET",
			path:   "/creativeappmappings/devices",
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
			if tt.req.Body() != nil {
				t.Errorf("Body() = %v, want nil", tt.req.Body())
			}
			if tt.hasQuery {
				if tt.req.Query() == nil {
					t.Fatal("Query() is nil, want non-nil")
				}
			}
			if !tt.hasQuery && tt.req.Query() != nil {
				t.Errorf("Query() = %v, want nil", tt.req.Query())
			}
		})
	}
}

func TestListRequest_Pagination(t *testing.T) {
	req := ListRequest{AdamID: "789", Limit: 20, Offset: 40}
	q := req.Query()
	if got := q.Get("limit"); got != "20" {
		t.Errorf("limit = %q, want %q", got, "20")
	}
	if got := q.Get("offset"); got != "40" {
		t.Errorf("offset = %q, want %q", got, "40")
	}
}

func TestListRequest_ZeroValues(t *testing.T) {
	req := ListRequest{AdamID: "789"}
	q := req.Query()
	if q.Get("limit") != "" {
		t.Errorf("limit = %q, want empty", q.Get("limit"))
	}
	if q.Get("offset") != "" {
		t.Errorf("offset = %q, want empty", q.Get("offset"))
	}
}

func TestGetRequest_PathIncludesIDs(t *testing.T) {
	req := GetRequest{AdamID: "111", ProductPageID: "222"}
	if got := req.Path(); got != "/apps/111/product-pages/222" {
		t.Errorf("Path() = %q, want %q", got, "/apps/111/product-pages/222")
	}
}

func TestLocalesRequest_PathIncludesIDs(t *testing.T) {
	req := LocalesRequest{AdamID: "111", ProductPageID: "222"}
	if got := req.Path(); got != "/apps/111/product-pages/222/locale-details" {
		t.Errorf("Path() = %q, want %q", got, "/apps/111/product-pages/222/locale-details")
	}
}
