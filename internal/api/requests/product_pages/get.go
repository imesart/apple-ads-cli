package product_pages

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single product page by ID.
type GetRequest struct {
	AdamID        string
	ProductPageID string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/apps/%s/product-pages/%s", r.AdamID, r.ProductPageID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
