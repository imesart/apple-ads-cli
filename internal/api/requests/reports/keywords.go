package reports

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// KeywordsRequest generates a keyword-level report.
// When AdGroupID is non-zero, reports at the ad group level;
// otherwise reports at the campaign level.
type KeywordsRequest struct {
	CampaignID string
	AdGroupID  string
	RawBody    json.RawMessage
}

func (r KeywordsRequest) Method() string { return "POST" }
func (r KeywordsRequest) Path() string {
	if r.AdGroupID != "" {
		return fmt.Sprintf("/reports/campaigns/%s/adgroups/%s/keywords", r.CampaignID, r.AdGroupID)
	}
	return fmt.Sprintf("/reports/campaigns/%s/keywords", r.CampaignID)
}
func (r KeywordsRequest) Body() any         { return r.RawBody }
func (r KeywordsRequest) Query() url.Values { return nil }
