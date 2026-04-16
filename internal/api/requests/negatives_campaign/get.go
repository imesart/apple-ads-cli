package negatives_campaign

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single campaign negative keyword by ID.
type GetRequest struct {
	CampaignID string
	KeywordID  string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/negativekeywords/%s", r.CampaignID, r.KeywordID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
