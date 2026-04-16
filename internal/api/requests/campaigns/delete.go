package campaigns

import (
	"fmt"
	"net/url"
)

// DeleteRequest deletes a campaign by ID.
type DeleteRequest struct {
	CampaignID string
}

func (r DeleteRequest) Method() string    { return "DELETE" }
func (r DeleteRequest) Path() string      { return fmt.Sprintf("/campaigns/%s", r.CampaignID) }
func (r DeleteRequest) Body() any         { return nil }
func (r DeleteRequest) Query() url.Values { return nil }
