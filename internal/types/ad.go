package types

// AdStatus is the user-controlled status of the ad.
type AdStatus string

const (
	AdStatusEnabled AdStatus = "ENABLED"
	AdStatusPaused  AdStatus = "PAUSED"
)

// AdDisplayStatus is the display status that derives from the ad's serving status.
type AdDisplayStatus string

const (
	AdDisplayStatusActive  AdDisplayStatus = "ACTIVE"
	AdDisplayStatusInvalid AdDisplayStatus = "INVALID"
	AdDisplayStatusOnHold  AdDisplayStatus = "ON_HOLD"
	AdDisplayStatusPaused  AdDisplayStatus = "PAUSED"
	AdDisplayStatusRemoved AdDisplayStatus = "REMOVED"
)

// AdServingStatus is the status of whether the ad is serving.
type AdServingStatus string

const (
	AdServingStatusRunning    AdServingStatus = "RUNNING"
	AdServingStatusNotRunning AdServingStatus = "NOT_RUNNING"
)

// AdServingStateReason is a reason the system provides when an ad isn't running.
type AdServingStateReason string

const (
	AdServingStateReasonAdApprovalPending             AdServingStateReason = "AD_APPROVAL_PENDING"
	AdServingStateReasonAdApprovalRejected            AdServingStateReason = "AD_APPROVAL_REJECTED"
	AdServingStateReasonAdProcessingInProgress        AdServingStateReason = "AD_PROCESSING_IN_PROGRESS"
	AdServingStateReasonCreativeSetInvalid            AdServingStateReason = "CREATIVE_SET_INVALID"
	AdServingStateReasonCreativeSetUnsupported        AdServingStateReason = "CREATIVE_SET_UNSUPPORTED"
	AdServingStateReasonDeletedByUser                 AdServingStateReason = "DELETED_BY_USER"
	AdServingStateReasonPausedByUser                  AdServingStateReason = "PAUSED_BY_USER"
	AdServingStateReasonPausedBySystem                AdServingStateReason = "PAUSED_BY_SYSTEM"
	AdServingStateReasonProductPageDeleted            AdServingStateReason = "PRODUCT_PAGE_DELETED"
	AdServingStateReasonProductPageHidden             AdServingStateReason = "PRODUCT_PAGE_HIDDEN"
	AdServingStateReasonProductPageIncompatible       AdServingStateReason = "PRODUCT_PAGE_INCOMPATIBLE"
	AdServingStateReasonProductPageInsufficientAssets AdServingStateReason = "PRODUCT_PAGE_INSUFFICIENT_ASSETS"
)

// Ad is the response to ad requests.
type Ad struct {
	ID                  *int64                 `json:"id,omitempty"`
	OrgID               *int64                 `json:"orgId,omitempty"`
	AdGroupID           *int64                 `json:"adGroupId,omitempty"`
	CampaignID          *int64                 `json:"campaignId,omitempty"`
	CreativeID          *int64                 `json:"creativeId,omitempty"`
	Name                *string                `json:"name,omitempty"`
	CreativeType        *CreativeKind          `json:"creativeType,omitempty"`
	Deleted             *bool                  `json:"deleted,omitempty"`
	Status              *AdStatus              `json:"status,omitempty"`
	ServingStatus       *AdServingStatus       `json:"servingStatus,omitempty"`
	ServingStateReasons []AdServingStateReason `json:"servingStateReasons,omitempty"`
	CreationTime        *string                `json:"creationTime,omitempty"`
	ModificationTime    *string                `json:"modificationTime,omitempty"`
}

// AdCreate is the request to create an ad and assign a creative to an ad group.
type AdCreate struct {
	ID                  *int64                 `json:"id,omitempty"`
	OrgID               *int64                 `json:"orgId,omitempty"`
	AdGroupID           *int64                 `json:"adGroupId,omitempty"`
	CampaignID          *int64                 `json:"campaignId,omitempty"`
	CreativeID          int64                  `json:"creativeId"`
	Name                string                 `json:"name"`
	CreativeType        *CreativeKind          `json:"creativeType,omitempty"`
	Deleted             *bool                  `json:"deleted,omitempty"`
	Status              AdStatus               `json:"status"`
	ServingStatus       *AdServingStatus       `json:"servingStatus,omitempty"`
	ServingStateReasons []AdServingStateReason `json:"servingStateReasons,omitempty"`
	CreationTime        *string                `json:"creationTime,omitempty"`
	ModificationTime    *string                `json:"modificationTime,omitempty"`
}

// AdUpdate is the request to update an ad.
type AdUpdate struct {
	ID                  *int64                 `json:"id,omitempty"`
	OrgID               *int64                 `json:"orgId,omitempty"`
	AdGroupID           *int64                 `json:"adGroupId,omitempty"`
	CampaignID          *int64                 `json:"campaignId,omitempty"`
	CreativeID          *int64                 `json:"creativeId,omitempty"`
	Name                *string                `json:"name,omitempty"`
	CreativeType        *CreativeKind          `json:"creativeType,omitempty"`
	Deleted             *bool                  `json:"deleted,omitempty"`
	Status              *AdStatus              `json:"status,omitempty"`
	ServingStatus       *AdServingStatus       `json:"servingStatus,omitempty"`
	ServingStateReasons []AdServingStateReason `json:"servingStateReasons,omitempty"`
	CreationTime        *string                `json:"creationTime,omitempty"`
	ModificationTime    *string                `json:"modificationTime,omitempty"`
}
