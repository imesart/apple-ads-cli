package acls

import (
	"testing"
)

func TestListRequest(t *testing.T) {
	req := ListRequest{}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/acls" {
		t.Errorf("Path() = %q, want %q", got, "/acls")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

func TestMeRequest(t *testing.T) {
	req := MeRequest{}

	if got := req.Method(); got != "GET" {
		t.Errorf("Method() = %q, want %q", got, "GET")
	}
	if got := req.Path(); got != "/me" {
		t.Errorf("Path() = %q, want %q", got, "/me")
	}
	if req.Body() != nil {
		t.Errorf("Body() = %v, want nil", req.Body())
	}
	if req.Query() != nil {
		t.Errorf("Query() = %v, want nil", req.Query())
	}
}

// TestACLRequestsAreSimpleGET verifies both ACL endpoints are parameterless GET requests.
func TestACLRequestsAreSimpleGET(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "List", method: "GET", path: "/acls"},
		{name: "Me", method: "GET", path: "/me"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req interface {
				Method() string
				Path() string
				Body() any
			}
			switch tt.name {
			case "List":
				req = ListRequest{}
			case "Me":
				req = MeRequest{}
			}
			if got := req.Method(); got != tt.method {
				t.Errorf("Method() = %q, want %q", got, tt.method)
			}
			if got := req.Path(); got != tt.path {
				t.Errorf("Path() = %q, want %q", got, tt.path)
			}
			if req.Body() != nil {
				t.Errorf("Body() = %v, want nil", req.Body())
			}
		})
	}
}
