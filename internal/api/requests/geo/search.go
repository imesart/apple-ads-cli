package geo

import (
	"net/url"
	"strconv"
)

// SearchRequest searches for geolocations.
type SearchRequest struct {
	SearchQuery string
	Entity      string
	CountryCode string
	Limit       int
	Offset      int
}

func (r SearchRequest) Method() string { return "GET" }
func (r SearchRequest) Path() string   { return "/search/geo" }
func (r SearchRequest) Body() any      { return nil }

func (r SearchRequest) Query() url.Values {
	v := url.Values{}
	if r.SearchQuery != "" {
		v.Set("query", r.SearchQuery)
	}
	if r.Entity != "" {
		v.Set("entity", r.Entity)
	}
	if r.CountryCode != "" {
		v.Set("countrycode", r.CountryCode)
	}
	if r.Limit > 0 {
		v.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		v.Set("offset", strconv.Itoa(r.Offset))
	}
	return v
}
