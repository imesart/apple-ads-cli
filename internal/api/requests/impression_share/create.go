package impression_share

import (
	"encoding/json"
	"net/url"
)

// CreateRequest creates a custom impression share report.
type CreateRequest struct {
	RawBody json.RawMessage
}

func (r CreateRequest) Method() string    { return "POST" }
func (r CreateRequest) Path() string      { return "/custom-reports" }
func (r CreateRequest) Body() any         { return r.RawBody }
func (r CreateRequest) Query() url.Values { return nil }
