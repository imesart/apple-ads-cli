package apps

import (
	"fmt"
	"net/url"
)

// LocalizedRequest retrieves localized app details by Adam ID.
type LocalizedRequest struct {
	AdamID string
}

func (r LocalizedRequest) Method() string { return "GET" }
func (r LocalizedRequest) Path() string {
	return fmt.Sprintf("/apps/%s/locale-details", r.AdamID)
}
func (r LocalizedRequest) Body() any         { return nil }
func (r LocalizedRequest) Query() url.Values { return nil }
