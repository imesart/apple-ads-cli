package types

// Pagination defines a range and limit of the number of returned records.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PageDetail contains the number of items that return in a page.
type PageDetail struct {
	TotalResults int `json:"totalResults"`
	StartIndex   int `json:"startIndex"`
	ItemsPerPage int `json:"itemsPerPage"`
}
