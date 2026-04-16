package campaigns

import (
	"encoding/json"
	"net/url"
)

// FindRequest finds campaigns matching a selector.
type FindRequest struct {
	RawBody json.RawMessage
}

func (r FindRequest) Method() string    { return "POST" }
func (r FindRequest) Path() string      { return "/campaigns/find" }
func (r FindRequest) Body() any         { return r.RawBody }
func (r FindRequest) Query() url.Values { return nil }
