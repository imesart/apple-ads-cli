package ads

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// FindRequest finds ads in a campaign ad group matching a selector.
type FindRequest struct {
	CampaignID string
	AdGroupID  string
	RawBody    json.RawMessage
}

func (r FindRequest) Method() string { return "POST" }
func (r FindRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/ads/find", r.CampaignID, r.AdGroupID)
}
func (r FindRequest) Body() any         { return r.RawBody }
func (r FindRequest) Query() url.Values { return nil }
