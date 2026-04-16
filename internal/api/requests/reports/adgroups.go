package reports

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// AdGroupsRequest generates an ad group-level report for a campaign.
type AdGroupsRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r AdGroupsRequest) Method() string { return "POST" }
func (r AdGroupsRequest) Path() string {
	return fmt.Sprintf("/reports/campaigns/%s/adgroups", r.CampaignID)
}
func (r AdGroupsRequest) Body() any         { return r.RawBody }
func (r AdGroupsRequest) Query() url.Values { return nil }
