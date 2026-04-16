package geo

import (
	"net/url"
	"strconv"
)

// GetRequest retrieves geolocation details for a specific geo identifier.
type GetRequest struct {
	ID     string
	Entity string
	Limit  int
	Offset int
}

func (r GetRequest) Method() string { return "POST" }
func (r GetRequest) Path() string   { return "/search/geo" }
func (r GetRequest) Body() any {
	return []map[string]string{{
		"id":     r.ID,
		"entity": geoEntityBodyValue(r.Entity),
	}}
}

func (r GetRequest) Query() url.Values {
	v := url.Values{}
	if r.Limit > 0 {
		v.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		v.Set("offset", strconv.Itoa(r.Offset))
	}
	return v
}

func geoEntityBodyValue(entity string) string {
	switch entity {
	case "Country":
		return "country"
	case "AdminArea":
		return "adminArea"
	case "Locality":
		return "locality"
	default:
		return entity
	}
}
