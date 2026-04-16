package adgroups

import (
	"fmt"
	"net/url"
)

// DeleteRequest deletes an ad group by ID.
type DeleteRequest struct {
	CampaignID string
	AdGroupID  string
}

func (r DeleteRequest) Method() string { return "DELETE" }
func (r DeleteRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s", r.CampaignID, r.AdGroupID)
}
func (r DeleteRequest) Body() any         { return nil }
func (r DeleteRequest) Query() url.Values { return nil }
