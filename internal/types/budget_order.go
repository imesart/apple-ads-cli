package types

// BudgetOrderStatus is the system-controlled status indicator for the budget order.
type BudgetOrderStatus string

const (
	BudgetOrderStatusActive    BudgetOrderStatus = "ACTIVE"
	BudgetOrderStatusInactive  BudgetOrderStatus = "INACTIVE"
	BudgetOrderStatusCancelled BudgetOrderStatus = "CANCELED"
	BudgetOrderStatusCompleted BudgetOrderStatus = "COMPLETED"
	BudgetOrderStatusExhausted BudgetOrderStatus = "EXHAUSTED"
)

// BudgetOrder is the response to requests for budget order details.
type BudgetOrder struct {
	ID                *int64             `json:"id,omitempty"`
	ParentOrgID       *int64             `json:"parentOrgId,omitempty"`
	PrimaryBuyerEmail *string            `json:"primaryBuyerEmail,omitempty"`
	PrimaryBuyerName  *string            `json:"primaryBuyerName,omitempty"`
	BillingEmail      *string            `json:"billingEmail,omitempty"`
	Name              *string            `json:"name,omitempty"`
	ClientName        *string            `json:"clientName,omitempty"`
	Budget            *Money             `json:"budget,omitempty"`
	StartDate         *string            `json:"startDate,omitempty"`
	EndDate           *string            `json:"endDate,omitempty"`
	OrderNumber       *string            `json:"orderNumber,omitempty"`
	SupplySources     []SupplySource     `json:"supplySources,omitempty"`
	Status            *BudgetOrderStatus `json:"status,omitempty"`
}

// BudgetOrderInfo is the parent object response to a request for budget order details.
type BudgetOrderInfo struct {
	Bo *BudgetOrder `json:"bo,omitempty"`
}

// BudgetOrderCreateBo contains the details of a budget order to create.
type BudgetOrderCreateBo struct {
	PrimaryBuyerEmail *string        `json:"primaryBuyerEmail,omitempty"`
	PrimaryBuyerName  *string        `json:"primaryBuyerName,omitempty"`
	BillingEmail      *string        `json:"billingEmail,omitempty"`
	Name              *string        `json:"name,omitempty"`
	ClientName        *string        `json:"clientName,omitempty"`
	Budget            *Money         `json:"budget,omitempty"`
	StartDate         *string        `json:"startDate,omitempty"`
	EndDate           *string        `json:"endDate,omitempty"`
	OrderNumber       *string        `json:"orderNumber,omitempty"`
	SupplySources     []SupplySource `json:"supplySources,omitempty"`
}

// BudgetOrderCreate is the request to create a budget order.
type BudgetOrderCreate struct {
	OrgIDs []int64             `json:"orgIds"`
	Bo     BudgetOrderCreateBo `json:"bo"`
}

// BudgetOrderUpdateBo contains the details of a budget order to update.
type BudgetOrderUpdateBo struct {
	PrimaryBuyerEmail *string `json:"primaryBuyerEmail,omitempty"`
	PrimaryBuyerName  *string `json:"primaryBuyerName,omitempty"`
	BillingEmail      *string `json:"billingEmail,omitempty"`
	Name              *string `json:"name,omitempty"`
	ClientName        *string `json:"clientName,omitempty"`
	Budget            *Money  `json:"budget,omitempty"`
	StartDate         *string `json:"startDate,omitempty"`
	EndDate           *string `json:"endDate,omitempty"`
	OrderNumber       *string `json:"orderNumber,omitempty"`
}

// BudgetOrderUpdate is the request to update a budget order.
type BudgetOrderUpdate struct {
	OrgIDs []int64             `json:"orgIds"`
	Bo     BudgetOrderUpdateBo `json:"bo"`
}
