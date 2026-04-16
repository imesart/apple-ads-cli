package ads

import (
	"fmt"
	"net/url"
)

// DeleteRequest deletes an ad by ID.
type DeleteRequest struct {
	CampaignID string
	AdGroupID  string
	AdID       string
}

func (r DeleteRequest) Method() string { return "DELETE" }
func (r DeleteRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/ads/%s", r.CampaignID, r.AdGroupID, r.AdID)
}
func (r DeleteRequest) Body() any         { return nil }
func (r DeleteRequest) Query() url.Values { return nil }
