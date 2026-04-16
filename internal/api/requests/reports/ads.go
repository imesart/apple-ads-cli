package reports

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// AdsRequest generates an ad-level report for a campaign.
type AdsRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r AdsRequest) Method() string { return "POST" }
func (r AdsRequest) Path() string {
	return fmt.Sprintf("/reports/campaigns/%s/ads", r.CampaignID)
}
func (r AdsRequest) Body() any         { return r.RawBody }
func (r AdsRequest) Query() url.Values { return nil }
