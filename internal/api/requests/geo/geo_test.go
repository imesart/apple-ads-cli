package geo

import "testing"

func TestSearchRequest(t *testing.T) {
	req := SearchRequest{
		SearchQuery: "luxembourg",
		Entity:      "Locality",
		CountryCode: "LU",
		Limit:       10,
		Offset:      20,
	}

	if got := req.Method(); got != "GET" {
		t.Fatalf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/search/geo" {
		t.Fatalf("Path() = %q, want %q", got, "/search/geo")
	}
	if req.Body() != nil {
		t.Fatalf("Body() = %#v, want nil", req.Body())
	}

	q := req.Query()
	if got := q.Get("query"); got != "luxembourg" {
		t.Fatalf("query = %q, want %q", got, "luxembourg")
	}
	if got := q.Get("entity"); got != "Locality" {
		t.Fatalf("entity = %q, want %q", got, "Locality")
	}
	if got := q.Get("countrycode"); got != "LU" {
		t.Fatalf("countrycode = %q, want %q", got, "LU")
	}
	if got := q.Get("limit"); got != "10" {
		t.Fatalf("limit = %q, want %q", got, "10")
	}
	if got := q.Get("offset"); got != "20" {
		t.Fatalf("offset = %q, want %q", got, "20")
	}
}

func TestGetRequest(t *testing.T) {
	req := GetRequest{
		ID:     "geo-123",
		Entity: "Locality",
		Limit:  10,
		Offset: 20,
	}

	if got := req.Method(); got != "POST" {
		t.Fatalf("Method() = %q, want %q", got, "POST")
	}
	if got := req.Path(); got != "/search/geo" {
		t.Fatalf("Path() = %q, want %q", got, "/search/geo")
	}

	body, ok := req.Body().([]map[string]string)
	if !ok {
		t.Fatalf("Body() type = %T, want []map[string]string", req.Body())
	}
	if len(body) != 1 {
		t.Fatalf("len(Body()) = %d, want 1", len(body))
	}
	if body[0]["id"] != "geo-123" {
		t.Fatalf("Body()[0][id] = %q, want %q", body[0]["id"], "geo-123")
	}
	if body[0]["entity"] != "locality" {
		t.Fatalf("Body()[0][entity] = %q, want %q", body[0]["entity"], "locality")
	}
	q := req.Query()
	if got := q.Get("limit"); got != "10" {
		t.Fatalf("limit = %q, want %q", got, "10")
	}
	if got := q.Get("offset"); got != "20" {
		t.Fatalf("offset = %q, want %q", got, "20")
	}
}
