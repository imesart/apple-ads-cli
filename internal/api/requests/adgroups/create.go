package adgroups

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// CreateRequest creates a new ad group in a campaign.
type CreateRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r CreateRequest) Method() string { return "POST" }
func (r CreateRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups", r.CampaignID)
}
func (r CreateRequest) Body() any         { return r.RawBody }
func (r CreateRequest) Query() url.Values { return nil }
