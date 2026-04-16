package reports

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// SearchTermsRequest generates a search term-level report.
// When AdGroupID is non-zero, reports at the ad group level;
// otherwise reports at the campaign level.
type SearchTermsRequest struct {
	CampaignID string
	AdGroupID  string
	RawBody    json.RawMessage
}

func (r SearchTermsRequest) Method() string { return "POST" }
func (r SearchTermsRequest) Path() string {
	if r.AdGroupID != "" {
		return fmt.Sprintf("/reports/campaigns/%s/adgroups/%s/searchterms", r.CampaignID, r.AdGroupID)
	}
	return fmt.Sprintf("/reports/campaigns/%s/searchterms", r.CampaignID)
}
func (r SearchTermsRequest) Body() any         { return r.RawBody }
func (r SearchTermsRequest) Query() url.Values { return nil }
