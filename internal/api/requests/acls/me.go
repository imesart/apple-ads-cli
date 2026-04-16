package acls

import "net/url"

// MeRequest retrieves the current user details.
type MeRequest struct{}

func (r MeRequest) Method() string    { return "GET" }
func (r MeRequest) Path() string      { return "/me" }
func (r MeRequest) Body() any         { return nil }
func (r MeRequest) Query() url.Values { return nil }
