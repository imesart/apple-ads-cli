package ads

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single ad by ID.
type GetRequest struct {
	CampaignID string
	AdGroupID  string
	AdID       string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/ads/%s", r.CampaignID, r.AdGroupID, r.AdID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
