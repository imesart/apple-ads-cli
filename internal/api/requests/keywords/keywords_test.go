package keywords

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestCreateRequest(t *testing.T) {
	body := json.RawMessage(`[{"text":"running shoes","matchType":"BROAD"}]`)
	req := CreateRequest{CampaignID: "100", AdGroupID: "200", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200/targetingkeywords/bulk" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200/targetingkeywords/bulk")
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

func TestCreateRequest_PathIncludesBothIDs(t *testing.T) {
	req := CreateRequest{CampaignID: "11", AdGroupID: "22"}

	if got := req.Path(); got != "/campaigns/11/adgroups/22/targetingkeywords/bulk" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/11/adgroups/22/targetingkeywords/bulk")
	}
}

func TestGetRequest(t *testing.T) {
	req := GetRequest{CampaignID: "100", AdGroupID: "200", KeywordID: "300"}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200/targetingkeywords/300" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200/targetingkeywords/300")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestGetRequest_PathIncludesAllIDs(t *testing.T) {
	req := GetRequest{CampaignID: "1", AdGroupID: "2", KeywordID: "3"}

	if got := req.Path(); got != "/campaigns/1/adgroups/2/targetingkeywords/3" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/1/adgroups/2/targetingkeywords/3")
	}
}

func TestListRequest(t *testing.T) {
	req := ListRequest{CampaignID: "100", AdGroupID: "200", Limit: 20, Offset: 40}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200/targetingkeywords" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200/targetingkeywords")
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
	req := ListRequest{CampaignID: "100", AdGroupID: "200"}

	query := req.Query()
	if query.Get("limit") != "" {
		t.Errorf("Query().Get(limit) = %q, want empty for zero value", query.Get("limit"))
	}
	if query.Get("offset") != "" {
		t.Errorf("Query().Get(offset) = %q, want empty for zero value", query.Get("offset"))
	}
}

func TestListRequest_PathIncludesBothIDs(t *testing.T) {
	req := ListRequest{CampaignID: "55", AdGroupID: "66"}

	if got := req.Path(); got != "/campaigns/55/adgroups/66/targetingkeywords" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/55/adgroups/66/targetingkeywords")
	}
}

func TestFindRequest(t *testing.T) {
	body := json.RawMessage(`{"conditions":[{"field":"matchType","operator":"EQUALS","values":["EXACT"]}]}`)
	req := FindRequest{CampaignID: "100", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/targetingkeywords/find" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/targetingkeywords/find")
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

func TestFindRequest_PathIsCampaignScoped(t *testing.T) {
	req := FindRequest{CampaignID: "88"}

	if got := req.Path(); got != "/campaigns/88/adgroups/targetingkeywords/find" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/88/adgroups/targetingkeywords/find")
	}
}

func TestUpdateRequest(t *testing.T) {
	body := json.RawMessage(`[{"id":300,"status":"PAUSED"}]`)
	req := UpdateRequest{CampaignID: "100", AdGroupID: "200", RawBody: body}

	if got := req.Method(); got != "PUT" {
		t.Errorf("Method() = %q, want %q", got, "PUT")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200/targetingkeywords/bulk" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200/targetingkeywords/bulk")
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

func TestDeleteOneRequest(t *testing.T) {
	req := DeleteOneRequest{CampaignID: "100", AdGroupID: "200", KeywordID: "300"}

	if got := req.Method(); got != "DELETE" {
		t.Errorf("Method() = %q, want %q", got, "DELETE")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200/targetingkeywords/300" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200/targetingkeywords/300")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestDeleteOneRequest_PathIncludesAllIDs(t *testing.T) {
	req := DeleteOneRequest{CampaignID: "5", AdGroupID: "6", KeywordID: "7"}

	if got := req.Path(); got != "/campaigns/5/adgroups/6/targetingkeywords/7" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/5/adgroups/6/targetingkeywords/7")
	}
}

func TestDeleteBulkRequest(t *testing.T) {
	body := json.RawMessage(`[300,301,302]`)
	req := DeleteBulkRequest{CampaignID: "100", AdGroupID: "200", RawBody: body}

	if got := req.Method(); got != "POST" {
		t.Errorf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/campaigns/100/adgroups/200/targetingkeywords/delete/bulk" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/100/adgroups/200/targetingkeywords/delete/bulk")
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

func TestDeleteBulkRequest_PathIncludesBothIDs(t *testing.T) {
	req := DeleteBulkRequest{CampaignID: "10", AdGroupID: "20"}

	if got := req.Path(); got != "/campaigns/10/adgroups/20/targetingkeywords/delete/bulk" {
		t.Errorf("Path() = %q, want %q", got, "/campaigns/10/adgroups/20/targetingkeywords/delete/bulk")
	}
}

// TestAllRequestMethods verifies all seven keyword requests use the expected HTTP methods.
func TestAllRequestMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		hasBody  bool
		hasQuery bool
	}{
		{name: "Create", method: "POST", path: "/campaigns/1/adgroups/2/targetingkeywords/bulk", hasBody: true},
		{name: "Get", method: "GET", path: "/campaigns/1/adgroups/2/targetingkeywords/3"},
		{name: "List", method: "GET", path: "/campaigns/1/adgroups/2/targetingkeywords", hasQuery: true},
		{name: "Find", method: "POST", path: "/campaigns/1/adgroups/targetingkeywords/find", hasBody: true},
		{name: "Update", method: "PUT", path: "/campaigns/1/adgroups/2/targetingkeywords/bulk", hasBody: true},
		{name: "DeleteOne", method: "DELETE", path: "/campaigns/1/adgroups/2/targetingkeywords/3"},
		{name: "DeleteBulk", method: "POST", path: "/campaigns/1/adgroups/2/targetingkeywords/delete/bulk", hasBody: true},
	}

	body := json.RawMessage(`{}`)
	requests := map[string]interface {
		Method() string
		Path() string
		Body() any
		Query() url.Values
	}{
		"Create":     CreateRequest{CampaignID: "1", AdGroupID: "2", RawBody: body},
		"Get":        GetRequest{CampaignID: "1", AdGroupID: "2", KeywordID: "3"},
		"List":       ListRequest{CampaignID: "1", AdGroupID: "2", Limit: 10},
		"Find":       FindRequest{CampaignID: "1", RawBody: body},
		"Update":     UpdateRequest{CampaignID: "1", AdGroupID: "2", RawBody: body},
		"DeleteOne":  DeleteOneRequest{CampaignID: "1", AdGroupID: "2", KeywordID: "3"},
		"DeleteBulk": DeleteBulkRequest{CampaignID: "1", AdGroupID: "2", RawBody: body},
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
