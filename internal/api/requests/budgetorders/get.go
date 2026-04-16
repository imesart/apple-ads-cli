package budgetorders

import (
	"fmt"
	"net/url"
)

// GetRequest retrieves a single budget order by ID.
type GetRequest struct {
	BudgetOrderID string
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string {
	return fmt.Sprintf("/budgetorders/%s", r.BudgetOrderID)
}
func (r GetRequest) Body() any         { return nil }
func (r GetRequest) Query() url.Values { return nil }
