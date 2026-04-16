package negatives_campaign

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// CreateRequest creates campaign negative keywords in bulk.
type CreateRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r CreateRequest) Method() string { return "POST" }
func (r CreateRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/negativekeywords/bulk", r.CampaignID)
}
func (r CreateRequest) Body() any         { return r.RawBody }
func (r CreateRequest) Query() url.Values { return nil }
