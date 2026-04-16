package campaigns

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single campaign by ID.
type GetRequest struct {
	CampaignID string
}

func (r GetRequest) Method() string    { return "GET" }
func (r GetRequest) Path() string      { return fmt.Sprintf("/campaigns/%s", r.CampaignID) }
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
