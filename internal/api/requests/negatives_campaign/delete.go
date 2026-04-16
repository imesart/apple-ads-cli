package negatives_campaign

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// DeleteBulkRequest deletes campaign negative keywords in bulk.
type DeleteBulkRequest struct {
	CampaignID string
	RawBody    json.RawMessage
}

func (r DeleteBulkRequest) Method() string { return "POST" }
func (r DeleteBulkRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/negativekeywords/delete/bulk", r.CampaignID)
}
func (r DeleteBulkRequest) Body() any         { return r.RawBody }
func (r DeleteBulkRequest) Query() url.Values { return nil }
