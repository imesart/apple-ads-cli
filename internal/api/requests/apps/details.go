package apps

import (
	"fmt"
	"net/url"
)

// DetailsRequest retrieves app details by Adam ID.
type DetailsRequest struct {
	AdamID string
}

func (r DetailsRequest) Method() string    { return "GET" }
func (r DetailsRequest) Path() string      { return fmt.Sprintf("/apps/%s", r.AdamID) }
func (r DetailsRequest) Body() any         { return nil }
func (r DetailsRequest) Query() url.Values { return nil }
