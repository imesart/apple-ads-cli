package keywords

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// UpdateRequest updates targeting keywords in bulk.
type UpdateRequest struct {
	CampaignID string
	AdGroupID  string
	RawBody    json.RawMessage
}

func (r UpdateRequest) Method() string { return "PUT" }
func (r UpdateRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/targetingkeywords/bulk", r.CampaignID, r.AdGroupID)
}
func (r UpdateRequest) Body() any         { return r.RawBody }
func (r UpdateRequest) Query() url.Values { return nil }
