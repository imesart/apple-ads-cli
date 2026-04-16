package adgroups

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestCreateRequest(t *testing.T) {
	body := json.RawMessage(`{"name":"Test AdGroup","defaultBidAmount":{"amount":"1.50","currency":"USD"}}`)
	req := CreateRequest{CampaignID: "100", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups")
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

func TestCreateRequest_PathIncludesCampaignID(t *testing.T) {
	req := CreateRequest{CampaignID: "55555"}

	if got := req.Path(); got != "/campaigns/55555/adgroups" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/55555/adgroups")
	}
}

func TestGetRequest(t *testing.T) {
	req := GetRequest{CampaignID: "100", AdGroupID: "200"}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestGetRequest_PathIncludesBothIDs(t *testing.T) {
	req := GetRequest{CampaignID: "42", AdGroupID: "99"}

	if got := req.Path(); got != "/campaigns/42/adgroups/99" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/42/adgroups/99")
	}
}

func TestListRequest(t *testing.T) {
	req := ListRequest{CampaignID: "100", Limit: 20, Offset: 40}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups")
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
	req := ListRequest{CampaignID: "100"}

	query := req.Query()
	if query.Get("limit") != "" {
		t.Errorf("Query().Get(limit) = %q, want empty for zero value", query.Get("limit"))
	}
	if query.Get("offset") != "" {
		t.Errorf("Query().Get(offset) = %q, want empty for zero value", query.Get("offset"))
	}
}

func TestListRequest_PathIncludesCampaignID(t *testing.T) {
	req := ListRequest{CampaignID: "77"}

	if got := req.Path(); got != "/campaigns/77/adgroups" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/77/adgroups")
	}
}

func TestFindRequest(t *testing.T) {
	body := json.RawMessage(`{"conditions":[{"field":"status","operator":"EQUALS","values":["ENABLED"]}]}`)
	req := FindRequest{CampaignID: "100", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/find" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/find")
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

func TestFindAllRequest(t *testing.T) {
	body := json.RawMessage(`{"conditions":[{"field":"name","operator":"CONTAINS","values":["test"]}]}`)
	req := FindAllRequest{RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/adgroups/find" {
		t.Errorf("Path() = %q, want %q", got, "/adgroups/find")
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
	body := json.RawMessage(`{"name":"Updated AdGroup","status":"PAUSED"}`)
	req := UpdateRequest{CampaignID: "100", AdGroupID: "200", RawBody: body}

	if got := req.Method(); got != "PUT" {
		t.Errorf("Method() = %q, want %q", got, "PUT")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200")
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

func TestUpdateRequest_PathIncludesBothIDs(t *testing.T) {
	req := UpdateRequest{CampaignID: "11", AdGroupID: "22"}

	if got := req.Path(); got != "/campaigns/11/adgroups/22" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/11/adgroups/22")
	}
}

func TestDeleteRequest(t *testing.T) {
	req := DeleteRequest{CampaignID: "100", AdGroupID: "200"}

	if got := req.Method(); got != "DELETE" {
		t.Errorf("Method() = %q, want %q", got, "DELETE")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestDeleteRequest_PathIncludesBothIDs(t *testing.T) {
	req := DeleteRequest{CampaignID: "33", AdGroupID: "44"}

	if got := req.Path(); got != "/campaigns/33/adgroups/44" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/33/adgroups/44")
	}
}

// TestAllRequestMethods verifies all seven ad group requests use the expected HTTP methods.
func TestAllRequestMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		hasBody  bool
		hasQuery bool
	}{
		{name: "Create", method: "POST", path: "/campaigns/1/adgroups", hasBody: true},
		{name: "Get", method: "GET", path: "/campaigns/1/adgroups/2"},
		{name: "List", method: "GET", path: "/campaigns/1/adgroups", hasQuery: true},
		{name: "Find", method: "POST", path: "/campaigns/1/adgroups/find", hasBody: true},
		{name: "FindAll", method: "POST", path: "/adgroups/find", hasBody: true},
		{name: "Update", method: "PUT", path: "/campaigns/1/adgroups/2", hasBody: true},
		{name: "Delete", method: "DELETE", path: "/campaigns/1/adgroups/2"},
	}

	body := json.RawMessage(`{}`)
	requests := map[string]interface {
		Method() string
		Path() string
		Body() any
		Query() url.Values
	}{
		"Create":  CreateRequest{CampaignID: "1", RawBody: body},
		"Get":     GetRequest{CampaignID: "1", AdGroupID: "2"},
		"List":    ListRequest{CampaignID: "1", Limit: 10},
		"Find":    FindRequest{CampaignID: "1", RawBody: body},
		"FindAll": FindAllRequest{RawBody: body},
		"Update":  UpdateRequest{CampaignID: "1", AdGroupID: "2", RawBody: body},
		"Delete":  DeleteRequest{CampaignID: "1", AdGroupID: "2"},
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
		})
	}
}
