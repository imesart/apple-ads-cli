package ad_rejections

import (
	"encoding/json"
	"net/url"
)

// FindRequest finds ad creative rejection reasons.
type FindRequest struct {
	RawBody json.RawMessage
}

func (r FindRequest) Method() string    { return "POST" }
func (r FindRequest) Path() string      { return "/product-page-reasons/find" }
func (r FindRequest) Body() any         { return r.RawBody }
func (r FindRequest) Query() url.Values { return nil }
