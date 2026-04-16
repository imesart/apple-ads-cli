package creatives

import (
	"net/url"
	"strconv"
)

// ListRequest lists all creatives with pagination.
type ListRequest struct {
	Limit  int
	Offset int
}

func (r ListRequest) Method() string { return "GET" }
func (r ListRequest) Path() string   { return "/creatives" }
func (r ListRequest) Body() any      { return nil }

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
