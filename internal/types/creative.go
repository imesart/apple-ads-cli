package types

// CreativeKind is the type of creative.
type CreativeKind string

const (
	CreativeKindCustomProductPage  CreativeKind = "CUSTOM_PRODUCT_PAGE"
	CreativeKindCreativeSet        CreativeKind = "CREATIVE_SET"
	CreativeKindDefaultProductPage CreativeKind = "DEFAULT_PRODUCT_PAGE"
)

// CreativeState is the system state of the creative.
type CreativeState string

const (
	CreativeStateValid   CreativeState = "VALID"
	CreativeStateInvalid CreativeState = "INVALID"
)

// CreativeStateReason is a reason the system provides for the creative's state.
type CreativeStateReason string

const (
	CreativeStateReasonAssetDeleted           CreativeStateReason = "ASSET_DELETED"
	CreativeStateReasonCreativeSetUnsupported CreativeStateReason = "CREATIVE_SET_UNSUPPORTED"
	CreativeStateReasonProductPageDeleted     CreativeStateReason = "PRODUCT_PAGE_DELETED"
	CreativeStateReasonProductPageHidden      CreativeStateReason = "PRODUCT_PAGE_HIDDEN"
)

// Creative is the creative object.
type Creative struct {
	ID               *int64                `json:"id,omitempty"`
	AdamID           *int64                `json:"adamId,omitempty"`
	OrgID            *int64                `json:"orgId,omitempty"`
	Name             *string               `json:"name,omitempty"`
	Type             *CreativeKind         `json:"type,omitempty"`
	State            *CreativeState        `json:"state,omitempty"`
	StateReasons     []CreativeStateReason `json:"stateReasons,omitempty"`
	CreationTime     *string               `json:"creationTime,omitempty"`
	ModificationTime *string               `json:"modificationTime,omitempty"`
}
