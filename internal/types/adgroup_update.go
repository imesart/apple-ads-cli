package types

// AdGroupUpdate is the payload to update an ad group.
type AdGroupUpdate struct {
	Status                 *AdGroupStatus       `json:"status,omitempty"`
	Name                   *string              `json:"name,omitempty"`
	DefaultBidAmount       *Money               `json:"defaultBidAmount,omitempty"`
	CpaGoal                *Money               `json:"cpaGoal,omitempty"`
	AutomatedKeywordsOptIn *bool                `json:"automatedKeywordsOptIn,omitempty"`
	StartTime              *string              `json:"startTime,omitempty"`
	EndTime                *string              `json:"endTime,omitempty"`
	TargetingDimensions    *TargetingDimensions `json:"targetingDimensions,omitempty"`
}
