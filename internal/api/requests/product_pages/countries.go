package product_pages

import "net/url"

// CountriesRequest retrieves supported countries or regions.
type CountriesRequest struct {
	CountriesOrRegions string
}

func (r CountriesRequest) Method() string { return "GET" }
func (r CountriesRequest) Path() string   { return "/countries-or-regions" }
func (r CountriesRequest) Body() any      { return nil }
func (r CountriesRequest) Query() url.Values {
	if r.CountriesOrRegions == "" {
		return nil
	}
	v := url.Values{}
	v.Set("countriesOrRegions", r.CountriesOrRegions)
	return v
}
