package campaigns

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestCreateRequest(t *testing.T) {
	body := json.RawMessage(`{"name":"Test Campaign","adamId":789}`)
	req := CreateRequest{RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns")
	}
	if req.Body() == nil {
		t.Error("Body() is nil, want non-nil")
	}
	if got, ok := req.Body().(json.RawMessage); !ok {
		t.Errorf("Body() type = %T, want json.RawMessage", req.Body())
	} else if string(got) != string(body) {
		t.Errorf("Body() = %s, want %s", got, body)
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestGetRequest(t *testing.T) {
	req := GetRequest{CampaignID: "123"}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/campaigns/123" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/123")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestGetRequest_DifferentID(t *testing.T) {
	req := GetRequest{CampaignID: "456789"}

	if got := req.Path(); got != "/campaigns/456789" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/456789")
	}
}

func TestListRequest(t *testing.T) {
	req := ListRequest{Limit: 20, Offset: 40}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/campaigns" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}

	query := req.Query()
	if query == nil {
		t.Fatal("Query() is nil, want non-nil")
	}
	if got := query.Get("limit"); got != "20" {
		t.Errorf("Query().Get(limit) = %q, want %q", got, "20")
	}
	if got := query.Get("offset"); got != "40" {
		t.Errorf("Query().Get(offset) = %q, want %q", got, "40")
	}
}

func TestListRequest_ZeroValues(t *testing.T) {
	req := ListRequest{}

	query := req.Query()
	if query.Get("limit") != "" {
		t.Errorf("Query().Get(limit) = %q, want empty for zero value", query.Get("limit"))
	}
	if query.Get("offset") != "" {
		t.Errorf("Query().Get(offset) = %q, want empty for zero value", query.Get("offset"))
	}
}

func TestListRequest_OnlyLimit(t *testing.T) {
	req := ListRequest{Limit: 50}

	query := req.Query()
	if got := query.Get("limit"); got != "50" {
		t.Errorf("Query().Get(limit) = %q, want %q", got, "50")
	}
	if query.Get("offset") != "" {
		t.Errorf("Query().Get(offset) = %q, want empty for zero value", query.Get("offset"))
	}
}

func TestFindRequest(t *testing.T) {
	body := json.RawMessage(`{"conditions":[{"field":"servingStatus","operator":"EQUALS","values":["RUNNING"]}]}`)
	req := FindRequest{RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/find" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/find")
	}
	if req.Body() == nil {
		t.Error("Body() is nil, want non-nil")
	}
	if got, ok := req.Body().(json.RawMessage); !ok {
		t.Errorf("Body() type = %T, want json.RawMessage", req.Body())
	} else if string(got) != string(body) {
		t.Errorf("Body() = %s, want %s", got, body)
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestUpdateRequest(t *testing.T) {
	body := json.RawMessage(`{"campaign":{"name":"Updated Campaign","status":"PAUSED"}}`)
	req := UpdateRequest{CampaignID: "123", RawBody: body}

	if got := req.Method(); got != "PUT" {
		t.Errorf("Method() = %q, want %q", got, "PUT")
	}
	if got := req.Path(); got != "/campaigns/123" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/123")
	}
	if req.Body() == nil {
		t.Error("Body() is nil, want non-nil")
	}
	if got, ok := req.Body().(json.RawMessage); !ok {
		t.Errorf("Body() type = %T, want json.RawMessage", req.Body())
	} else if string(got) != string(body) {
		t.Errorf("Body() = %s, want %s", got, body)
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestUpdateRequest_PathIncludesCampaignID(t *testing.T) {
	req := UpdateRequest{CampaignID: "99999"}

	if got := req.Path(); got != "/campaigns/99999" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/99999")
	}
}

func TestDeleteRequest(t *testing.T) {
	req := DeleteRequest{CampaignID: "123"}

	if got := req.Method(); got != "DELETE" {
		t.Errorf("Method() = %q, want %q", got, "DELETE")
	}
	if got := req.Path(); got != "/campaigns/123" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/123")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestDeleteRequest_PathIncludesCampaignID(t *testing.T) {
	req := DeleteRequest{CampaignID: "42"}

	if got := req.Path(); got != "/campaigns/42" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/42")
	}
}

// TestAllRequestMethods verifies all six campaign requests use the expected HTTP methods.
func TestAllRequestMethods(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		hasBody    bool
		hasQuery   bool
		queryCheck func(url.Values)
	}{
		{
			name:    "Create",
			method:  "POST",
			path:    "/campaigns",
			hasBody: true,
		},
		{
			name:   "Get",
			method: "GET",
			path:   "/campaigns/1",
		},
		{
			name:     "List",
			method:   "GET",
			path:     "/campaigns",
			hasQuery: true,
			queryCheck: func(v url.Values) {
				if v.Get("limit") != "10" {
					t.Errorf("List query limit = %q, want %q", v.Get("limit"), "10")
				}
			},
		},
		{
			name:    "Find",
			method:  "POST",
			path:    "/campaigns/find",
			hasBody: true,
		},
		{
			name:    "Update",
			method:  "PUT",
			path:    "/campaigns/1",
			hasBody: true,
		},
		{
			name:   "Delete",
			method: "DELETE",
			path:   "/campaigns/1",
		},
	}

	body := json.RawMessage(`{}`)
	requests := map[string]interface {
		Method() string
		Path() string
		Body() any
		Query() url.Values
	}{
		"Create": CreateRequest{RawBody: body},
		"Get":    GetRequest{CampaignID: "1"},
		"List":   ListRequest{Limit: 10, Offset: 0},
		"Find":   FindRequest{RawBody: body},
		"Update": UpdateRequest{CampaignID: "1", RawBody: body},
		"Delete": DeleteRequest{CampaignID: "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := requests[tt.name]
			if got := req.Method(); got != tt.method {
				t.Errorf("Method() = %q, want %q", got, tt.method)
			}
			if got := req.Path(); got != tt.path {
				t.Errorf("Path() = %q, want %q", got, tt.path)
			}
			if tt.hasBody && req.Body() == nil {
				t.Error("Body() is nil, want non-nil")
			}
			if !tt.hasBody && req.Body() != nil {
				t.Errorf("Body() = %v, want nil", req.Body())
			}
			if tt.hasQuery {
				query := req.Query()
				if query == nil {
					t.Fatal("Query() is nil, want non-nil")
				}
				if tt.queryCheck != nil {
					tt.queryCheck(query)
				}
			}
		})
	}
}
