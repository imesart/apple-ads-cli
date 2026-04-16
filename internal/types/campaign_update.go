package types

// CampaignUpdateProperties contains the list of campaign fields that are updatable.
type CampaignUpdateProperties struct {
	Name               *string            `json:"name,omitempty"`
	Status             *CampaignStatus    `json:"status,omitempty"`
	BudgetAmount       *Money             `json:"budgetAmount,omitempty"`
	BudgetOrders       []int64            `json:"budgetOrders,omitempty"`
	DailyBudgetAmount  *Money             `json:"dailyBudgetAmount,omitempty"`
	CountriesOrRegions []string           `json:"countriesOrRegions,omitempty"`
	LOCInvoiceDetails  *LOCInvoiceDetails `json:"locInvoiceDetails,omitempty"`
	StartTime          *string            `json:"startTime,omitempty"`
	EndTime            *string            `json:"endTime,omitempty"`
}

// CampaignUpdate is the payload to update a campaign.
// The API expects the campaign properties wrapped in a "campaign" key.
type CampaignUpdate struct {
	ClearGeoTargetingOnCountryOrRegionChange *bool                     `json:"clearGeoTargetingOnCountryOrRegionChange,omitempty"`
	Campaign                                 *CampaignUpdateProperties `json:"campaign,omitempty"`
}
