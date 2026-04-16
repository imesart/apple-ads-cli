package ads

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
			req:     CreateRequest{CampaignID: "1", AdGroupID: "2", RawBody: body},
			method:  "POST",
			path:    "/campaigns/1/adgroups/2/ads",
			hasBody: true,
		},
		{
			name:   "Get",
			req:    GetRequest{CampaignID: "1", AdGroupID: "2", AdID: "3"},
			method: "GET",
			path:   "/campaigns/1/adgroups/2/ads/3",
		},
		{
			name:     "List",
			req:      ListRequest{CampaignID: "1", AdGroupID: "2", Limit: 10, Offset: 5},
			method:   "GET",
			path:     "/campaigns/1/adgroups/2/ads",
			hasQuery: true,
		},
		{
			name:    "Find",
			req:     FindRequest{CampaignID: "1", AdGroupID: "2", RawBody: body},
			method:  "POST",
			path:    "/campaigns/1/adgroups/2/ads/find",
			hasBody: true,
		},
		{
			name:    "FindAll",
			req:     FindAllRequest{RawBody: body},
			method:  "POST",
			path:    "/ads/find",
			hasBody: true,
		},
		{
			name:    "Update",
			req:     UpdateRequest{CampaignID: "1", AdGroupID: "2", AdID: "3", RawBody: body},
			method:  "PUT",
			path:    "/campaigns/1/adgroups/2/ads/3",
			hasBody: true,
		},
		{
			name:   "Delete",
			req:    DeleteRequest{CampaignID: "1", AdGroupID: "2", AdID: "3"},
			method: "DELETE",
			path:   "/campaigns/1/adgroups/2/ads/3",
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
				q := tt.req.Query()
				if q == nil {
					t.Fatal("Query() is nil, want non-nil")
				}
			}
			if !tt.hasQuery && tt.req.Query() != nil && len(tt.req.Query()) > 0 {
				t.Errorf("Query() = %v, want nil or empty", tt.req.Query())
			}
		})
	}
}

func TestListRequest_Pagination(t *testing.T) {
	req := ListRequest{CampaignID: "1", AdGroupID: "2", Limit: 20, Offset: 40}
	q := req.Query()
	if got := q.Get("limit"); got != "20" {
		t.Errorf("limit = %q, want %q", got, "20")
	}
	if got := q.Get("offset"); got != "40" {
		t.Errorf("offset = %q, want %q", got, "40")
	}
}

func TestListRequest_ZeroValues(t *testing.T) {
	req := ListRequest{CampaignID: "1", AdGroupID: "2"}
	q := req.Query()
	if q.Get("limit") != "" {
		t.Errorf("limit = %q, want empty", q.Get("limit"))
	}
	if q.Get("offset") != "" {
		t.Errorf("offset = %q, want empty", q.Get("offset"))
	}
}

func TestPathIncludesIDs(t *testing.T) {
	req := GetRequest{CampaignID: "99", AdGroupID: "88", AdID: "77"}
	if got := req.Path(); got != "/campaigns/99/adgroups/88/ads/77" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/99/adgroups/88/ads/77")
	}
}
