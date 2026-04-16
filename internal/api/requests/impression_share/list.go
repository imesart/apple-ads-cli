package impression_share

import (
	"net/url"
	"strconv"
)

// ListRequest lists all custom impression share reports with pagination.
type ListRequest struct {
	Field     string
	SortOrder string
	Limit     int
	Offset    int
}

func (r ListRequest) Method() string { return "GET" }
func (r ListRequest) Path() string   { return "/custom-reports" }
func (r ListRequest) Body() any      { return nil }

func (r ListRequest) Query() url.Values {
	v := url.Values{}
	if r.Field != "" {
		v.Set("field", r.Field)
	}
	if r.Limit > 0 {
		v.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		v.Set("offset", strconv.Itoa(r.Offset))
	}
	if r.SortOrder != "" {
		v.Set("sortOrder", r.SortOrder)
	}
	return v
}
