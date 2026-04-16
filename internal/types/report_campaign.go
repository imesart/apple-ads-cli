package types

// CampaignAppDetail contains the app data from campaign-level reports.
type CampaignAppDetail struct {
	AdamID  *int64  `json:"adamId,omitempty"`
	AppName *string `json:"appName,omitempty"`
}

// ReportingCampaign is the response to a request to fetch campaign-level reports.
type ReportingCampaign struct {
	AdChannelType                      *CampaignAdChannelType                                  `json:"adChannelType,omitempty"`
	App                                *CampaignAppDetail                                      `json:"app,omitempty"`
	CampaignID                         *int64                                                  `json:"campaignId,omitempty"`
	CampaignName                       *string                                                 `json:"campaignName,omitempty"`
	CampaignStatus                     *CampaignStatus                                         `json:"campaignStatus,omitempty"`
	CountriesOrRegions                 []string                                                `json:"countriesOrRegions,omitempty"`
	CountryOrRegionServingStateReasons map[string][]CampaignCountryOrRegionsServingStateReason `json:"countryOrRegionServingStateReasons,omitempty"`
	DailyBudget                        *Money                                                  `json:"dailyBudget,omitempty"`
	Deleted                            *bool                                                   `json:"deleted,omitempty"`
	DisplayStatus                      *CampaignDisplayStatus                                  `json:"displayStatus,omitempty"`
	ModificationTime                   *string                                                 `json:"modificationTime,omitempty"`
	OrgID                              *int64                                                  `json:"orgId,omitempty"`
	ServingStateReasons                []CampaignServingStateReason                            `json:"servingStateReasons,omitempty"`
	ServingStatus                      *CampaignServingStatus                                  `json:"servingStatus,omitempty"`
	SupplySources                      []SupplySource                                          `json:"supplySources,omitempty"`
	TotalBudget                        *Money                                                  `json:"totalBudget,omitempty"`
	DeviceClass                        *string                                                 `json:"deviceClass,omitempty"`
	Gender                             *string                                                 `json:"gender,omitempty"`
	AgeRange                           *string                                                 `json:"ageRange,omitempty"`
	CountryOrRegion                    *string                                                 `json:"countryOrRegion,omitempty"`
	AdminArea                          *string                                                 `json:"adminArea,omitempty"`
	Locality                           *string                                                 `json:"locality,omitempty"`
}
