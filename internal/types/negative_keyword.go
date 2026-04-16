package types

// NegativeKeyword contains negative keyword parameters for requests and responses.
type NegativeKeyword struct {
	ID               *int64           `json:"id,omitempty"`
	CampaignID       *int64           `json:"campaignId,omitempty"`
	AdGroupID        *int64           `json:"adGroupId,omitempty"`
	Text             string           `json:"text"`
	MatchType        KeywordMatchType `json:"matchType"`
	Status           *KeywordStatus   `json:"status,omitempty"`
	Deleted          *bool            `json:"deleted,omitempty"`
	ModificationTime *string          `json:"modificationTime,omitempty"`
}
