package impression_share

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single custom impression share report by ID.
type GetRequest struct {
	ReportID string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/custom-reports/%s", r.ReportID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
