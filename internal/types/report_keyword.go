package types

// KeywordBidRecommendation contains the suggested bid amount for a keyword.
type KeywordBidRecommendation struct {
	SuggestedBidAmount *Money `json:"suggestedBidAmount,omitempty"`
}

// KeywordInsights contains bid recommendations for a keyword.
type KeywordInsights struct {
	BidRecommendation *KeywordBidRecommendation `json:"bidRecommendation,omitempty"`
}

// ReportingKeyword is the response to a request to fetch keyword-level reports.
type ReportingKeyword struct {
	AdGroupDeleted       *bool                 `json:"adGroupDeleted,omitempty"`
	AdGroupID            *int64                `json:"adGroupId,omitempty"`
	AdGroupName          *string               `json:"adGroupName,omitempty"`
	BidAmount            *Money                `json:"bidAmount,omitempty"`
	CampaignID           *int64                `json:"campaignId,omitempty"`
	Deleted              *bool                 `json:"deleted,omitempty"`
	KeywordID            *int64                `json:"keywordId,omitempty"`
	Keyword              *string               `json:"keyword,omitempty"`
	KeywordStatus        *KeywordStatus        `json:"keywordStatus,omitempty"`
	KeywordDisplayStatus *KeywordDisplayStatus `json:"keywordDisplayStatus,omitempty"`
	MatchType            *KeywordMatchType     `json:"matchType,omitempty"`
	OrgID                *int64                `json:"orgId,omitempty"`
	ModificationTime     *string               `json:"modificationTime,omitempty"`
	DeviceClass          *string               `json:"deviceClass,omitempty"`
	Gender               *string               `json:"gender,omitempty"`
	AgeRange             *string               `json:"ageRange,omitempty"`
	CountryOrRegion      *string               `json:"countryOrRegion,omitempty"`
	AdminArea            *string               `json:"adminArea,omitempty"`
	Locality             *string               `json:"locality,omitempty"`
}
