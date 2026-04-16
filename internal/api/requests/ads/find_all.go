package ads

import (
	"encoding/json"
	"net/url"
)

// FindAllRequest finds ads across all campaigns.
type FindAllRequest struct {
	RawBody json.RawMessage
}

func (r FindAllRequest) Method() string    { return "POST" }
func (r FindAllRequest) Path() string      { return "/ads/find" }
func (r FindAllRequest) Body() any         { return r.RawBody }
func (r FindAllRequest) Query() url.Values { return nil }
