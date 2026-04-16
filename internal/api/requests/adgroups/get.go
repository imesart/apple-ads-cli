package adgroups

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single ad group by ID.
type GetRequest struct {
	CampaignID string
	AdGroupID  string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s", r.CampaignID, r.AdGroupID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
