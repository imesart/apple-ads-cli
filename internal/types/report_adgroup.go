package types

// ReportingAdGroup is the response to a request to fetch ad group-level reports.
type ReportingAdGroup struct {
	AdGroupDisplayStatus       *AdGroupDisplayStatus       `json:"adGroupDisplayStatus,omitempty"`
	AdGroupID                  *int64                      `json:"adGroupId,omitempty"`
	AdGroupName                *string                     `json:"adGroupName,omitempty"`
	AdGroupServingStateReasons []AdGroupServingStateReason `json:"adGroupServingStateReasons,omitempty"`
	AdGroupServingStatus       *AdGroupServingStatus       `json:"adGroupServingStatus,omitempty"`
	AdGroupStatus              *AdGroupStatus              `json:"adGroupStatus,omitempty"`
	AutomatedKeywordsOptIn     *bool                       `json:"automatedKeywordsOptIn,omitempty"`
	CampaignID                 *int64                      `json:"campaignId,omitempty"`
	CpaGoal                    *Money                      `json:"cpaGoal,omitempty"`
	DefaultBidAmount           *Money                      `json:"defaultBidAmount,omitempty"`
	Deleted                    *bool                       `json:"deleted,omitempty"`
	StartTime                  *string                     `json:"startTime,omitempty"`
	EndTime                    *string                     `json:"endTime,omitempty"`
	ModificationTime           *string                     `json:"modificationTime,omitempty"`
	OrgID                      *int64                      `json:"orgId,omitempty"`
	DeviceClass                *string                     `json:"deviceClass,omitempty"`
	Gender                     *string                     `json:"gender,omitempty"`
	AgeRange                   *string                     `json:"ageRange,omitempty"`
	CountryOrRegion            *string                     `json:"countryOrRegion,omitempty"`
	AdminArea                  *string                     `json:"adminArea,omitempty"`
	Locality                   *string                     `json:"locality,omitempty"`
}
