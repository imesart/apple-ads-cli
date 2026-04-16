package types

// Money represents budget amounts in campaigns.
// The amount is a string that can contain up to two decimal digits.
type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}
