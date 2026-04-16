package budgetorders

import (
	"encoding/json"
	"net/url"
)

// CreateRequest creates a new budget order.
type CreateRequest struct {
	RawBody json.RawMessage
}

func (r CreateRequest) Method() string    { return "POST" }
func (r CreateRequest) Path() string      { return "/budgetorders" }
func (r CreateRequest) Body() any         { return r.RawBody }
func (r CreateRequest) Query() url.Values { return nil }
