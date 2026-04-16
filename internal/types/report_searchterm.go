package types

// SearchTermSource is the source of the keyword to use as a search term.
type SearchTermSource string

const (
	SearchTermSourceAuto     SearchTermSource = "AUTO"
	SearchTermSourceTargeted SearchTermSource = "TARGETED"
)

// ReportingSearchTerm is the response to a request to fetch search term-level reports.
type ReportingSearchTerm struct {
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
	SearchTermSource     *SearchTermSource     `json:"searchTermSource,omitempty"`
	SearchTermText       *string               `json:"searchTermText,omitempty"`
	DeviceClass          *string               `json:"deviceClass,omitempty"`
	Gender               *string               `json:"gender,omitempty"`
	AgeRange             *string               `json:"ageRange,omitempty"`
	CountryOrRegion      *string               `json:"countryOrRegion,omitempty"`
	AdminArea            *string               `json:"adminArea,omitempty"`
	Locality             *string               `json:"locality,omitempty"`
}
