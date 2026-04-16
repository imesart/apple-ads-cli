package ads

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// CreateRequest creates a new ad in an ad group.
type CreateRequest struct {
	CampaignID string
	AdGroupID  string
	RawBody    json.RawMessage
}

func (r CreateRequest) Method() string { return "POST" }
func (r CreateRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/ads", r.CampaignID, r.AdGroupID)
}
func (r CreateRequest) Body() any         { return r.RawBody }
func (r CreateRequest) Query() url.Values { return nil }
