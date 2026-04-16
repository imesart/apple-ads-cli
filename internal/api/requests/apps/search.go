package apps

import (
	"net/url"
	"strconv"
)

// SearchRequest searches for iOS apps.
type SearchRequest struct {
	SearchQuery     string
	ReturnOwnedApps bool
	Limit           int
	Offset          int
}

func (r SearchRequest) Method() string { return "GET" }
func (r SearchRequest) Path() string   { return "/search/apps" }
func (r SearchRequest) Body() any      { return nil }

func (r SearchRequest) Query() url.Values {
	v := url.Values{}
	if r.SearchQuery != "" {
		v.Set("query", r.SearchQuery)
	}
	if r.ReturnOwnedApps {
		v.Set("returnOwnedApps", "true")
	}
	if r.Limit > 0 {
		v.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		v.Set("offset", strconv.Itoa(r.Offset))
	}
	return v
}
