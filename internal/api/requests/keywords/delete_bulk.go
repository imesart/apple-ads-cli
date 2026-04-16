package keywords

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// DeleteBulkRequest deletes targeting keywords in bulk.
type DeleteBulkRequest struct {
	CampaignID string
	AdGroupID  string
	RawBody    json.RawMessage
}

func (r DeleteBulkRequest) Method() string { return "POST" }
func (r DeleteBulkRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/targetingkeywords/delete/bulk", r.CampaignID, r.AdGroupID)
}
func (r DeleteBulkRequest) Body() any         { return r.RawBody }
func (r DeleteBulkRequest) Query() url.Values { return nil }
