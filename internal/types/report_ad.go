package types

// ReportingAd is the response to a request to fetch ad-level reports.
type ReportingAd struct {
	AdID                  *int64                 `json:"adId,omitempty"`
	AdGroupID             *int64                 `json:"adGroupId,omitempty"`
	CampaignID            *int64                 `json:"campaignId,omitempty"`
	CreativeID            *int64                 `json:"creativeId,omitempty"`
	OrgID                 *int64                 `json:"orgId,omitempty"`
	ProductPageID         *string                `json:"productPageId,omitempty"`
	AdName                *string                `json:"adName,omitempty"`
	CreativeType          *CreativeKind          `json:"creativeType,omitempty"`
	Status                *CreativeState         `json:"status,omitempty"`
	AdDisplayStatus       *AdDisplayStatus       `json:"adDisplayStatus,omitempty"`
	AdServingStateReasons []AdServingStateReason `json:"adServingStateReasons,omitempty"`
	Language              *string                `json:"language,omitempty"`
	Deleted               *bool                  `json:"deleted,omitempty"`
	CreationTime          *string                `json:"creationTime,omitempty"`
	ModificationTime      *string                `json:"modificationTime,omitempty"`
	DeviceClass           *string                `json:"deviceClass,omitempty"`
	Gender                *string                `json:"gender,omitempty"`
	AgeRange              *string                `json:"ageRange,omitempty"`
	CountryOrRegion       *string                `json:"countryOrRegion,omitempty"`
	AdminArea             *string                `json:"adminArea,omitempty"`
	Locality              *string                `json:"locality,omitempty"`
}
