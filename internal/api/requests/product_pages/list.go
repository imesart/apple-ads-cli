package product_pages

import (
	"fmt"
	"net/url"
	"strconv"
)

// ListRequest lists custom product pages for an app.
type ListRequest struct {
	AdamID string
	Name   string
	State  string
	Limit  int
	Offset int
}

func (r ListRequest) Method() string { return "GET" }
func (r ListRequest) Path() string {
	return fmt.Sprintf("/apps/%s/product-pages", r.AdamID)
}
func (r ListRequest) Body() any { return nil }

func (r ListRequest) Query() url.Values {
	v := url.Values{}
	if r.Name != "" {
		v.Set("name", r.Name)
	}
	if r.State != "" {
		v.Set("state", r.State)
	}
	if r.Limit > 0 {
		v.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		v.Set("offset", strconv.Itoa(r.Offset))
	}
	return v
}
