package types

// ReportingGranularity is the report data granularity.
type ReportingGranularity string

const (
	ReportingGranularityHourly  ReportingGranularity = "HOURLY"
	ReportingGranularityDaily   ReportingGranularity = "DAILY"
	ReportingGranularityWeekly  ReportingGranularity = "WEEKLY"
	ReportingGranularityMonthly ReportingGranularity = "MONTHLY"
)

// ReportingGroupBy is a dimension used to group report responses.
type ReportingGroupBy string

const (
	ReportingGroupByAdminArea       ReportingGroupBy = "adminArea"
	ReportingGroupByAgeRange        ReportingGroupBy = "ageRange"
	ReportingGroupByCountryCode     ReportingGroupBy = "countryCode"
	ReportingGroupByCountryOrRegion ReportingGroupBy = "countryOrRegion"
	ReportingGroupByDeviceClass     ReportingGroupBy = "deviceClass"
	ReportingGroupByGender          ReportingGroupBy = "gender"
	ReportingGroupByLocality        ReportingGroupBy = "locality"
)

// ReportingRequest is the report request body.
type ReportingRequest struct {
	StartTime                  *string               `json:"startTime,omitempty"`
	EndTime                    *string               `json:"endTime,omitempty"`
	TimeZone                   *string               `json:"timeZone,omitempty"`
	Granularity                *ReportingGranularity `json:"granularity,omitempty"`
	GroupBy                    []ReportingGroupBy    `json:"groupBy,omitempty"`
	Selector                   *Selector             `json:"selector,omitempty"`
	ReturnGrandTotals          *bool                 `json:"returnGrandTotals,omitempty"`
	ReturnRecordsWithNoMetrics *bool                 `json:"returnRecordsWithNoMetrics,omitempty"`
	ReturnRowTotals            *bool                 `json:"returnRowTotals,omitempty"`
}

// SpendRow contains reporting response metrics.
type SpendRow struct {
	AvgCPM               *Money   `json:"avgCPM,omitempty"`
	AvgCPT               *Money   `json:"avgCPT,omitempty"`
	Impressions          *int     `json:"impressions,omitempty"`
	LocalSpend           *Money   `json:"localSpend,omitempty"`
	TapInstallCPI        *Money   `json:"tapInstallCPI,omitempty"`
	TapInstallRate       *float64 `json:"tapInstallRate,omitempty"`
	TapInstalls          *int     `json:"tapInstalls,omitempty"`
	TapNewDownloads      *int     `json:"tapNewDownloads,omitempty"`
	TapRedownloads       *int     `json:"tapRedownloads,omitempty"`
	TapPreOrdersPlaced   *int     `json:"tapPreOrdersPlaced,omitempty"`
	Taps                 *int     `json:"taps,omitempty"`
	TotalAvgCPI          *Money   `json:"totalAvgCPI,omitempty"`
	TotalInstallRate     *float64 `json:"totalInstallRate,omitempty"`
	TotalInstalls        *int     `json:"totalInstalls,omitempty"`
	TotalNewDownloads    *int     `json:"totalNewDownloads,omitempty"`
	TotalRedownloads     *int     `json:"totalRedownloads,omitempty"`
	TotalPreOrdersPlaced *int     `json:"totalPreOrdersPlaced,omitempty"`
	TTR                  *float64 `json:"ttr,omitempty"`
	ViewInstalls         *int     `json:"viewInstalls,omitempty"`
	ViewNewDownloads     *int     `json:"viewNewDownloads,omitempty"`
	ViewRedownloads      *int     `json:"viewRedownloads,omitempty"`
	ViewPreOrdersPlaced  *int     `json:"viewPreOrdersPlaced,omitempty"`
}

// SpendRowExtended contains reporting response metrics with a date.
type SpendRowExtended struct {
	AvgCPM               *Money   `json:"avgCPM,omitempty"`
	AvgCPT               *Money   `json:"avgCPT,omitempty"`
	Date                 *string  `json:"date,omitempty"`
	Impressions          *int     `json:"impressions,omitempty"`
	LocalSpend           *Money   `json:"localSpend,omitempty"`
	TapInstallCPI        *Money   `json:"tapInstallCPI,omitempty"`
	TapInstallRate       *float64 `json:"tapInstallRate,omitempty"`
	TapInstalls          *int     `json:"tapInstalls,omitempty"`
	TapNewDownloads      *int     `json:"tapNewDownloads,omitempty"`
	TapRedownloads       *int     `json:"tapRedownloads,omitempty"`
	TapPreOrdersPlaced   *int     `json:"tapPreOrdersPlaced,omitempty"`
	Taps                 *int     `json:"taps,omitempty"`
	TotalAvgCPI          *Money   `json:"totalAvgCPI,omitempty"`
	TotalInstallRate     *float64 `json:"totalInstallRate,omitempty"`
	TotalInstalls        *int     `json:"totalInstalls,omitempty"`
	TotalNewDownloads    *int     `json:"totalNewDownloads,omitempty"`
	TotalRedownloads     *int     `json:"totalRedownloads,omitempty"`
	TotalPreOrdersPlaced *int     `json:"totalPreOrdersPlaced,omitempty"`
	TTR                  *float64 `json:"ttr,omitempty"`
	ViewInstalls         *int     `json:"viewInstalls,omitempty"`
	ViewNewDownloads     *int     `json:"viewNewDownloads,omitempty"`
	ViewRedownloads      *int     `json:"viewRedownloads,omitempty"`
	ViewPreOrdersPlaced  *int     `json:"viewPreOrdersPlaced,omitempty"`
}

// GrandTotalsRow is the summary of cumulative metrics.
type GrandTotalsRow struct {
	Other *bool     `json:"other,omitempty"`
	Total *SpendRow `json:"total,omitempty"`
}

// Row contains report metrics by time granularity.
type Row[Insights any, Metadata any] struct {
	Insights    *Insights          `json:"insights,omitempty"`
	Metadata    *Metadata          `json:"metadata,omitempty"`
	Other       *bool              `json:"other,omitempty"`
	Total       *SpendRow          `json:"total,omitempty"`
	Granularity []SpendRowExtended `json:"granularity,omitempty"`
}
