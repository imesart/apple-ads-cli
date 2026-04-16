package ad_rejections

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves rejection reasons by ID.
type GetRequest struct {
	ID string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/product-page-reasons/%s", r.ID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
