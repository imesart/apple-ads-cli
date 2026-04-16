package acls

import "net/url"

// ListRequest lists all organizations (ACLs).
type ListRequest struct{}

func (r ListRequest) Method() string    { return "GET" }
func (r ListRequest) Path() string      { return "/acls" }
func (r ListRequest) Body() any         { return nil }
func (r ListRequest) Query() url.Values { return nil }
