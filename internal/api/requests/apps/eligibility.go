package apps

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// EligibilityRequest checks app eligibility for advertising.
type EligibilityRequest struct {
	AdamID  string
	RawBody json.RawMessage
}

func (r EligibilityRequest) Method() string { return "POST" }
func (r EligibilityRequest) Path() string {
	return fmt.Sprintf("/apps/%s/eligibilities/find", r.AdamID)
}
func (r EligibilityRequest) Body() any         { return r.RawBody }
func (r EligibilityRequest) Query() url.Values { return nil }
