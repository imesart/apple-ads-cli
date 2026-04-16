package api

import "net/url"

// Request defines the interface for API endpoint requests.
// Every API endpoint implements this interface to describe how to build
// the corresponding HTTP request.
type Request interface {
	// Method returns the HTTP method (GET, POST, PUT, DELETE).
	Method() string

	// Path returns the URL path relative to the API base URL.
	Path() string

	// Body returns the request body to marshal as JSON, or nil for no body.
	Body() any

	// Query returns URL query parameters to append.
	Query() url.Values
}
