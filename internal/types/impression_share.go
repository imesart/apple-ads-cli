package types

// CustomReportState is the state of an Impression Share report.
type CustomReportState string

const (
	CustomReportStateQueued    CustomReportState = "QUEUED"
	CustomReportStatePending   CustomReportState = "PENDING"
	CustomReportStateCompleted CustomReportState = "COMPLETED"
	CustomReportStateFailed    CustomReportState = "FAILED"
)

// CustomReportDimension is a dimension for an Impression Share report.
type CustomReportDimension string

const (
	CustomReportDimensionAdamID          CustomReportDimension = "adamId"
	CustomReportDimensionAppName         CustomReportDimension = "appName"
	CustomReportDimensionCountryOrRegion CustomReportDimension = "countryOrRegion"
	CustomReportDimensionSearchTerm      CustomReportDimension = "searchTerm"
)

// CustomReportMetric is a metric for an Impression Share report.
type CustomReportMetric string

const (
	CustomReportMetricLowImpressionShare  CustomReportMetric = "lowImpressionShare"
	CustomReportMetricHighImpressionShare CustomReportMetric = "highImpressionShare"
	CustomReportMetricRank                CustomReportMetric = "rank"
	CustomReportMetricSearchPopularity    CustomReportMetric = "searchPopularity"
)

// CustomReportDateRange is the date range of an Impression Share report request.
type CustomReportDateRange string

const (
	CustomReportDateRangeCustom     CustomReportDateRange = "CUSTOM"
	CustomReportDateRangeLastWeek   CustomReportDateRange = "LAST_WEEK"
	CustomReportDateRangeLast2Weeks CustomReportDateRange = "LAST_2_WEEKS"
	CustomReportDateRangeLast4Weeks CustomReportDateRange = "LAST_4_WEEKS"
)

// CustomReportGranularity is the granularity of an Impression Share report.
type CustomReportGranularity string

const (
	CustomReportGranularityDaily  CustomReportGranularity = "DAILY"
	CustomReportGranularityWeekly CustomReportGranularity = "WEEKLY"
)

// SovCondition is a condition for filtering Impression Share reports.
// Only the IN operator is supported.
type SovCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Values   []any  `json:"values"`
}

// CustomReportSelector is a selector to filter Impression Share report results.
type CustomReportSelector struct {
	Conditions []SovCondition `json:"conditions,omitempty"`
}

// CustomReport is a container for Impression Share report metrics.
type CustomReport struct {
	ID               *int64                   `json:"id,omitempty"`
	Name             *string                  `json:"name,omitempty"`
	State            *CustomReportState       `json:"state,omitempty"`
	StartTime        *string                  `json:"startTime,omitempty"`
	EndTime          *string                  `json:"endTime,omitempty"`
	DateRange        *CustomReportDateRange   `json:"dateRange,omitempty"`
	CreationTime     *string                  `json:"creationTime,omitempty"`
	ModificationTime *string                  `json:"modificationTime,omitempty"`
	Granularity      *CustomReportGranularity `json:"granularity,omitempty"`
	Dimensions       []CustomReportDimension  `json:"dimensions,omitempty"`
	Metrics          []CustomReportMetric     `json:"metrics,omitempty"`
	Selector         *CustomReportSelector    `json:"selector,omitempty"`
	DownloadURI      *string                  `json:"downloadUri,omitempty"`
}

// CustomReportRequest is the Impression Share report request body.
type CustomReportRequest struct {
	Name        string                   `json:"name"`
	StartTime   *string                  `json:"startTime,omitempty"`
	EndTime     *string                  `json:"endTime,omitempty"`
	DateRange   *CustomReportDateRange   `json:"dateRange,omitempty"`
	Granularity *CustomReportGranularity `json:"granularity,omitempty"`
	Selector    *CustomReportSelector    `json:"selector,omitempty"`
}
