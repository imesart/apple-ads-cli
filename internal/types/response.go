package types

// DataResponse wraps a single data object in an API response.
type DataResponse[T any] struct {
	Data T `json:"data"`
}

// ListResponse wraps a list of data objects with pagination in an API response.
type ListResponse[T any] struct {
	Data       []T         `json:"data"`
	Pagination *PageDetail `json:"pagination,omitempty"`
}

// ErrorItem contains the details of an individual API error.
type ErrorItem struct {
	Field       *string `json:"field,omitempty"`
	Message     *string `json:"message,omitempty"`
	MessageCode *string `json:"messageCode,omitempty"`
}

// ErrorResponse contains the API's nested error envelope.
type ErrorResponse struct {
	Error struct {
		Errors []ErrorItem `json:"errors"`
	} `json:"error"`
}

// ReportingResponse wraps reporting data with grandTotals and rows.
type ReportingResponse[Insights any, Metadata any] struct {
	ReportingDataResponse *ReportingData[Insights, Metadata] `json:"reportingDataResponse,omitempty"`
}

// ReportingData contains the total metrics and rows for a report.
type ReportingData[Insights any, Metadata any] struct {
	GrandTotals *GrandTotalsRow           `json:"grandTotals,omitempty"`
	Row         []Row[Insights, Metadata] `json:"row,omitempty"`
}
