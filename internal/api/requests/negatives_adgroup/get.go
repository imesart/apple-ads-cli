package negatives_adgroup

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single ad group negative keyword by ID.
type GetRequest struct {
	CampaignID string
	AdGroupID  string
	KeywordID  string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups/%s/negativekeywords/%s", r.CampaignID, r.AdGroupID, r.KeywordID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
