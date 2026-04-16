package keywords

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// FindRequest finds targeting keywords across a campaign's ad groups.
type FindRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r FindRequest) Method() string { return "POST" }
func (r FindRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/targetingkeywords/find", r.CampaignID)
}
func (r FindRequest) Body() any         { return r.RawBody }
func (r FindRequest) Query() url.Values { return nil }
