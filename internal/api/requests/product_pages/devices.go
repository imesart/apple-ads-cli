package product_pages

import "net/url"

// DevicesRequest retrieves app preview device sizes for creative mappings.
type DevicesRequest struct{}

func (r DevicesRequest) Method() string    { return "GET" }
func (r DevicesRequest) Path() string      { return "/creativeappmappings/devices" }
func (r DevicesRequest) Body() any         { return nil }
func (r DevicesRequest) Query() url.Values { return nil }
