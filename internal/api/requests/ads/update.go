package ads

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// UpdateRequest updates an existing ad.
type UpdateRequest struct {
	CampaignID string
	AdGroupID  string
	AdID       string
	RawBody    json.RawMessage
}

func (r UpdateRequest) Method() string { return "PUT" }
func (r UpdateRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/ads/%s", r.CampaignID, r.AdGroupID, r.AdID)
}
func (r UpdateRequest) Body() any         { return r.RawBody }
func (r UpdateRequest) Query() url.Values { return nil }
