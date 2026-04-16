package types

// KeywordMatchType is an automated keyword and bidding strategy.
type KeywordMatchType string

const (
	KeywordMatchTypeAuto  KeywordMatchType = "AUTO"
	KeywordMatchTypeBroad KeywordMatchType = "BROAD"
	KeywordMatchTypeExact KeywordMatchType = "EXACT"
)

// KeywordStatus is the user-controlled status of a keyword.
type KeywordStatus string

const (
	KeywordStatusActive KeywordStatus = "ACTIVE"
	KeywordStatusPaused KeywordStatus = "PAUSED"
)

// KeywordDisplayStatus is the state of the keyword display operation.
type KeywordDisplayStatus string

const (
	KeywordDisplayStatusAdGroupOnHold  KeywordDisplayStatus = "AD_GROUP_ON_HOLD"
	KeywordDisplayStatusCampaignOnHold KeywordDisplayStatus = "CAMPAIGN_ON_HOLD"
	KeywordDisplayStatusDeleted        KeywordDisplayStatus = "DELETED"
	KeywordDisplayStatusPaused         KeywordDisplayStatus = "PAUSED"
	KeywordDisplayStatusRunning        KeywordDisplayStatus = "RUNNING"
)

// Keyword contains targeting keyword parameters.
type Keyword struct {
	ID               *int64           `json:"id,omitempty"`
	CampaignID       *int64           `json:"campaignId,omitempty"`
	AdGroupID        *int64           `json:"adGroupId,omitempty"`
	Text             string           `json:"text"`
	MatchType        KeywordMatchType `json:"matchType"`
	Status           *KeywordStatus   `json:"status,omitempty"`
	BidAmount        *Money           `json:"bidAmount,omitempty"`
	Deleted          *bool            `json:"deleted,omitempty"`
	CreationTime     *string          `json:"creationTime,omitempty"`
	ModificationTime *string          `json:"modificationTime,omitempty"`
}

// KeywordUpdate contains keyword parameters for update requests.
type KeywordUpdate struct {
	ID               *int64            `json:"id,omitempty"`
	AdGroupID        *int64            `json:"adGroupId,omitempty"`
	Status           *KeywordStatus    `json:"status,omitempty"`
	Text             *string           `json:"text,omitempty"`
	MatchType        *KeywordMatchType `json:"matchType,omitempty"`
	BidAmount        *Money            `json:"bidAmount,omitempty"`
	Deleted          *bool             `json:"deleted,omitempty"`
	ModificationTime *string           `json:"modificationTime,omitempty"`
}
