package budgetorders

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// UpdateRequest updates an existing budget order.
type UpdateRequest struct {
	BudgetOrderID string
	RawBody       json.RawMessage
}

func (r UpdateRequest) Method() string { return "PUT" }
func (r UpdateRequest) Path() string {
	return fmt.Sprintf("/budgetorders/%s", r.BudgetOrderID)
}
func (r UpdateRequest) Body() any         { return r.RawBody }
func (r UpdateRequest) Query() url.Values { return nil }
