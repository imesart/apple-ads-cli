package creatives

import (
	"encoding/json"
	"net/url"
)

// FindRequest finds creatives matching a selector.
type FindRequest struct {
	RawBody json.RawMessage
}

func (r FindRequest) Method() string    { return "POST" }
func (r FindRequest) Path() string      { return "/creatives/find" }
func (r FindRequest) Body() any         { return r.RawBody }
func (r FindRequest) Query() url.Values { return nil }
