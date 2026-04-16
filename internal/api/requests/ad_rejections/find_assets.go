package ad_rejections

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// FindAssetsRequest finds app assets for rejection analysis.
type FindAssetsRequest struct {
	AdamID  string
	RawBody json.RawMessage
}

func (r FindAssetsRequest) Method() string { return "POST" }
func (r FindAssetsRequest) Path() string {
	return fmt.Sprintf("/apps/%s/assets/find", r.AdamID)
}
func (r FindAssetsRequest) Body() any         { return r.RawBody }
func (r FindAssetsRequest) Query() url.Values { return nil }
