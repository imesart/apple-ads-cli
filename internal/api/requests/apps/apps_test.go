package apps

import (
	"testing"
)

func TestSearchRequest(t *testing.T) {
	req := SearchRequest{SearchQuery: "weather", ReturnOwnedApps: true, Limit: 10, Offset: 20}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/search/apps" {
		t.Errorf("Path() = %q, want %q", got, "/search/apps")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}

	query := req.Query()
	if query == nil {
		t.Fatal("Query() is nil, want non-nil")
	}
	if got := query.Get("query"); got != "weather" {
		t.Errorf("Query().Get(query) = %q, want %q", got, "weather")
	}
	if got := query.Get("returnOwnedApps"); got != "true" {
		t.Errorf("Query().Get(returnOwnedApps) = %q, want %q", got, "true")
	}
	if got := query.Get("limit"); got != "10" {
		t.Errorf("Query().Get(limit) = %q, want %q", got, "10")
	}
	if got := query.Get("offset"); got != "20" {
		t.Errorf("Query().Get(offset) = %q, want %q", got, "20")
	}
}

func TestSearchRequest_QueryOnly(t *testing.T) {
	req := SearchRequest{SearchQuery: "fitness tracker"}

	query := req.Query()
	if query == nil {
		t.Fatal("Query() is nil, want non-nil")
	}
	if got := query.Get("query"); got != "fitness tracker" {
		t.Errorf("Query().Get(query) = %q, want %q", got, "fitness tracker")
	}
	if got := query.Get("returnOwnedApps"); got != "" {
		t.Errorf("Query().Get(returnOwnedApps) = %q, want empty (not set)", got)
	}
}

func TestSearchRequest_ReturnOwnedAppsOnly(t *testing.T) {
	req := SearchRequest{ReturnOwnedApps: true}

	query := req.Query()
	if query == nil {
		t.Fatal("Query() is nil, want non-nil")
	}
	if got := query.Get("query"); got != "" {
		t.Errorf("Query().Get(query) = %q, want empty (not set)", got)
	}
	if got := query.Get("returnOwnedApps"); got != "true" {
		t.Errorf("Query().Get(returnOwnedApps) = %q, want %q", got, "true")
	}
}

func TestSearchRequest_NoParams(t *testing.T) {
	req := SearchRequest{}

	query := req.Query()
	if query.Get("query") != "" {
		t.Errorf("Query().Get(query) = %q, want empty", query.Get("query"))
	}
	if query.Get("returnOwnedApps") != "" {
		t.Errorf("Query().Get(returnOwnedApps) = %q, want empty", query.Get("returnOwnedApps"))
	}
}

func TestSearchRequest_ReturnOwnedAppsFalse(t *testing.T) {
	req := SearchRequest{SearchQuery: "test", ReturnOwnedApps: false}

	query := req.Query()
	// When ReturnOwnedApps is false, it should not be included in the query
	if got := query.Get("returnOwnedApps"); got != "" {
		t.Errorf("Query().Get(returnOwnedApps) = %q, want empty when false", got)
	}
}

func TestSearchRequest_SpecialCharactersInQuery(t *testing.T) {
	req := SearchRequest{SearchQuery: "photo & video"}

	query := req.Query()
	if got := query.Get("query"); got != "photo & video" {
		t.Errorf("Query().Get(query) = %q, want %q", got, "photo & video")
	}
}
