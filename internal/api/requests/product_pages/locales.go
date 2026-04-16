package product_pages

import (
	"fmt"
	"net/url"
)

// LocalesRequest retrieves locale details for a product page.
type LocalesRequest struct {
	AdamID        string
	ProductPageID string
}

func (r LocalesRequest) Method() string { return "GET" }
func (r LocalesRequest) Path() string {
	return fmt.Sprintf("/apps/%s/product-pages/%s/locale-details", r.AdamID, r.ProductPageID)
}
func (r LocalesRequest) Body() any         { return nil }
func (r LocalesRequest) Query() url.Values { return nil }
