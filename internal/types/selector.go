package types

// Operator defines the comparison operators available for conditions.
type Operator string

const (
	OperatorBetween     Operator = "BETWEEN"
	OperatorContains    Operator = "CONTAINS"
	OperatorContainsAll Operator = "CONTAINS_ALL"
	OperatorContainsAny Operator = "CONTAINS_ANY"
	OperatorEndsWith    Operator = "ENDSWITH"
	OperatorEquals      Operator = "EQUALS"
	OperatorGreaterThan Operator = "GREATER_THAN"
	OperatorIn          Operator = "IN"
	OperatorLessThan    Operator = "LESS_THAN"
	OperatorStartsWith  Operator = "STARTSWITH"
)

// SortOrder defines the ordering direction for sorted results.
type SortOrder string

const (
	SortOrderAscending  SortOrder = "ASCENDING"
	SortOrderDescending SortOrder = "DESCENDING"
)

// Condition filters a list of records, similar to a WHERE clause in SQL.
type Condition struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator"`
	Values   []any    `json:"values"`
}

// Sorting defines the order of grouped results.
type Sorting struct {
	Field     string    `json:"field"`
	SortOrder SortOrder `json:"sortOrder"`
}

// Selector defines what data the API returns when fetching resources.
type Selector struct {
	Conditions []Condition `json:"conditions,omitempty"`
	Fields     []string    `json:"fields,omitempty"`
	OrderBy    []Sorting   `json:"orderBy,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}
