package types

// AdGroupStatus is the user-controlled status to enable or pause the ad group.
type AdGroupStatus string

const (
	AdGroupStatusEnabled AdGroupStatus = "ENABLED"
	AdGroupStatusPaused  AdGroupStatus = "PAUSED"
)

// AdGroupServingStatus is the serving status of the ad group.
type AdGroupServingStatus string

const (
	AdGroupServingStatusRunning    AdGroupServingStatus = "RUNNING"
	AdGroupServingStatusNotRunning AdGroupServingStatus = "NOT_RUNNING"
)

// AdGroupDisplayStatus is the display status of the ad group.
type AdGroupDisplayStatus string

const (
	AdGroupDisplayStatusRunning        AdGroupDisplayStatus = "RUNNING"
	AdGroupDisplayStatusOnHold         AdGroupDisplayStatus = "ON_HOLD"
	AdGroupDisplayStatusCampaignOnHold AdGroupDisplayStatus = "CAMPAIGN_ON_HOLD"
	AdGroupDisplayStatusPaused         AdGroupDisplayStatus = "PAUSED"
	AdGroupDisplayStatusDeleted        AdGroupDisplayStatus = "DELETED"
)

// AdGroupServingStateReason is a reason that displays when an ad group isn't running.
type AdGroupServingStateReason string

const (
	AdGroupServingStateReasonAdGroupPausedByUser                         AdGroupServingStateReason = "AD_GROUP_PAUSED_BY_USER"
	AdGroupServingStateReasonAppNotSupport                               AdGroupServingStateReason = "APP_NOT_SUPPORT"
	AdGroupServingStateReasonDeletedByUser                               AdGroupServingStateReason = "DELETED_BY_USER"
	AdGroupServingStateReasonAdGroupEndDateReached                       AdGroupServingStateReason = "ADGROUP_END_DATE_REACHED"
	AdGroupServingStateReasonAudienceBelowThreshold                      AdGroupServingStateReason = "AUDIENCE_BELOW_THRESHOLD"
	AdGroupServingStateReasonCampaignEndDateReached                      AdGroupServingStateReason = "CAMPAIGN_END_DATE_REACHED"
	AdGroupServingStateReasonCampaignNotRunning                          AdGroupServingStateReason = "CAMPAIGN_NOT_RUNNING"
	AdGroupServingStateReasonCampaignStartDateInFuture                   AdGroupServingStateReason = "CAMPAIGN_START_DATE_IN_FUTURE"
	AdGroupServingStateReasonStartDateInTheFuture                        AdGroupServingStateReason = "START_DATE_IN_THE_FUTURE"
	AdGroupServingStateReasonNoAvailableAds                              AdGroupServingStateReason = "NO_AVAILABLE_ADS"
	AdGroupServingStateReasonPendingAudienceVerification                 AdGroupServingStateReason = "PENDING_AUDIENCE_VERIFICATION"
	AdGroupServingStateReasonTargetedDeviceClassNotSupportedSupplySource AdGroupServingStateReason = "TARGETED_DEVICE_CLASS_NOT_SUPPORTED_SUPPLY_SOURCE"
)

// PricingModel is the type of pricing model for a bid.
type PricingModel string

const (
	PricingModelCPC PricingModel = "CPC"
	PricingModelCPM PricingModel = "CPM"
)

// AdGroup is the response to ad group requests.
type AdGroup struct {
	ID                     *int64                      `json:"id,omitempty"`
	OrgID                  *int64                      `json:"orgId,omitempty"`
	CampaignID             *int64                      `json:"campaignId,omitempty"`
	Status                 *AdGroupStatus              `json:"status,omitempty"`
	ServingStatus          *AdGroupServingStatus       `json:"servingStatus,omitempty"`
	ServingStateReasons    []AdGroupServingStateReason `json:"servingStateReasons,omitempty"`
	DisplayStatus          *AdGroupDisplayStatus       `json:"displayStatus,omitempty"`
	Name                   string                      `json:"name"`
	PricingModel           PricingModel                `json:"pricingModel"`
	PaymentModel           *PaymentModel               `json:"paymentModel,omitempty"`
	DefaultBidAmount       Money                       `json:"defaultBidAmount"`
	CpaGoal                *Money                      `json:"cpaGoal,omitempty"`
	Deleted                *bool                       `json:"deleted,omitempty"`
	AutomatedKeywordsOptIn *bool                       `json:"automatedKeywordsOptIn,omitempty"`
	ModificationTime       *string                     `json:"modificationTime,omitempty"`
	StartTime              string                      `json:"startTime"`
	EndTime                *string                     `json:"endTime,omitempty"`
	TargetingDimensions    *TargetingDimensions        `json:"targetingDimensions,omitempty"`
}
