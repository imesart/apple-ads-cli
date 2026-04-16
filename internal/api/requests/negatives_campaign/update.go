package negatives_campaign

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// UpdateRequest updates campaign negative keywords in bulk.
type UpdateRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r UpdateRequest) Method() string { return "PUT" }
func (r UpdateRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/negativekeywords/bulk", r.CampaignID)
}
func (r UpdateRequest) Body() any         { return r.RawBody }
func (r UpdateRequest) Query() url.Values { return nil }
