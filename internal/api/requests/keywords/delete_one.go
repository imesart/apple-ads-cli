package keywords

import (
	"fmt"
	"net/url"
)

// DeleteOneRequest deletes a single targeting keyword by ID.
type DeleteOneRequest struct {
	CampaignID string
	AdGroupID  string
	KeywordID  string
}

func (r DeleteOneRequest) Method() string { return "DELETE" }
func (r DeleteOneRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/targetingkeywords/%s", r.CampaignID, r.AdGroupID, r.KeywordID)
}
func (r DeleteOneRequest) Body() any         { return nil }
func (r DeleteOneRequest) Query() url.Values { return nil }
