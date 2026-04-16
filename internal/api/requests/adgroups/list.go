package adgroups

import (
	"fmt"
	"net/url"
	"strconv"
)

// ListRequest lists all ad groups in a campaign with pagination.
type ListRequest struct {
	CampaignID string
	Limit      int
	Offset     int
}

func (r ListRequest) Method() string { return "GET" }
func (r ListRequest) Path() string {
	return fmt.Sprintf("/campaigns/%s/adgroups", r.CampaignID)
}
func (r ListRequest) Body() any { return nil }

func (r ListRequest) Query() url.Values {
	v := url.Values{}
	if r.Limit > 0 {
		v.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		v.Set("offset", strconv.Itoa(r.Offset))
	}
	return v
}
