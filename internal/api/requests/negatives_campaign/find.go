package negatives_campaign

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// FindRequest finds campaign negative keywords matching a selector.
type FindRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r FindRequest) Method() string { return "POST" }
func (r FindRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/negativekeywords/find", r.CampaignID)
}
func (r FindRequest) Body() any         { return r.RawBody }
func (r FindRequest) Query() url.Values { return nil }
