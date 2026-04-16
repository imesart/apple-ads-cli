package reports

import (
	"encoding/json"
	"net/url"
)

// CampaignsRequest generates a campaign-level report.
type CampaignsRequest struct {
	RawBody json.RawMessage
}

func (r CampaignsRequest) Method() string    { return "POST" }
func (r CampaignsRequest) Path() string      { return "/reports/campaigns" }
func (r CampaignsRequest) Body() any         { return r.RawBody }
func (r CampaignsRequest) Query() url.Values { return nil }
