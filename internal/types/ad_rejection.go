package types

// AdRejectionReasonLevel is the level at which the system applies an ad rejection reason.
type AdRejectionReasonLevel string

const (
	AdRejectionReasonLevelCustomProductPage        AdRejectionReasonLevel = "CUSTOM_PRODUCT_PAGE"
	AdRejectionReasonLevelCustomProductPageLocale  AdRejectionReasonLevel = "CUSTOM_PRODUCT_PAGE_LOCALE"
	AdRejectionReasonLevelDefaultProductPage       AdRejectionReasonLevel = "DEFAULT_PRODUCT_PAGE"
	AdRejectionReasonLevelDefaultProductPageLocale AdRejectionReasonLevel = "DEFAULT_PRODUCT_PAGE_LOCALE"
)

// ProductPageReasonType is the type of reason for a product page rejection.
type ProductPageReasonType string

const (
	ProductPageReasonTypeRejectionReason ProductPageReasonType = "REJECTION_REASON"
)

// ProductPageReasonCode is the rejection reason code for a product page.
type ProductPageReasonCode string

const (
	ProductPageReasonCodeAppIconGraphicOrAdultThemedContent ProductPageReasonCode = "APP_ICON_GRAPHIC_OR_ADULT_THEMED_CONTENT"
	ProductPageReasonCodeAppIconNotEligible                 ProductPageReasonCode = "APP_ICON_NOT_ELIGIBLE"
	ProductPageReasonCodeAppNameLanguageConflict            ProductPageReasonCode = "APP_NAME_LANGUAGE_CONFLICT"
	ProductPageReasonCodeAppNameGraphicOrAdultThemedContent ProductPageReasonCode = "APP_NAME_GRAPHIC_OR_ADULT_THEMED_CONTENT"
	ProductPageReasonCodeAppNameNotEligible                 ProductPageReasonCode = "APP_NAME_NOT_ELIGIBLE"
	ProductPageReasonCodeAppNotEligibleAtThisTime           ProductPageReasonCode = "APP_NOT_ELIGIBLE_AT_THIS_TIME"
	ProductPageReasonCodeMimicsAppStoreTodayCard            ProductPageReasonCode = "MIMICS_APP_STORE_TODAY_CARD"
	ProductPageReasonCodePPOExperimentAppIconNotEligible    ProductPageReasonCode = "PRODUCT_PAGE_OPTIMIZATION_EXPERIMENT_APP_ICON_NOT_ELIGIBLE"
	ProductPageReasonCodeSubtitleGraphicOrAdultThemed       ProductPageReasonCode = "SUBTITLE_GRAPHIC_OR_ADULT_THEMED_CONTENT"
	ProductPageReasonCodeSubtitleLanguageConflict           ProductPageReasonCode = "SUBTITLE_LANGUAGE_CONFLICT"
	ProductPageReasonCodeSubtitleNotEligible                ProductPageReasonCode = "SUBTITLE_NOT_ELIGIBLE"
	ProductPageReasonCodeSubtitlePricingOffersOrRanking     ProductPageReasonCode = "SUBTITLE_PRICING_OFFERS_OR_RANKING_CLAIMS"
)

// ProductPageReason is the ad creative rejection reason based on a product page.
type ProductPageReason struct {
	ID              *int64                  `json:"id,omitempty"`
	AdamID          *int64                  `json:"adamId,omitempty"`
	ProductPageID   *string                 `json:"productPageId,omitempty"`
	AssetGenID      *string                 `json:"assetGenId,omitempty"`
	LanguageCode    *string                 `json:"languageCode,omitempty"`
	ReasonCode      *ProductPageReasonCode  `json:"reasonCode,omitempty"`
	ReasonType      *ProductPageReasonType  `json:"reasonType,omitempty"`
	ReasonLevel     *AdRejectionReasonLevel `json:"reasonLevel,omitempty"`
	SupplySource    *SupplySource           `json:"supplySource,omitempty"`
	CountryOrRegion *string                 `json:"countryOrRegion,omitempty"`
	Comment         *string                 `json:"comment,omitempty"`
}

// MediaAssetType is the type of creative asset.
type MediaAssetType string

const (
	MediaAssetTypeAppPreview MediaAssetType = "APP_PREVIEW"
	MediaAssetTypeScreenshot MediaAssetType = "SCREENSHOT"
)

// MediaAssetOrientation is the orientation of an asset.
type MediaAssetOrientation string

const (
	MediaAssetOrientationLandscape MediaAssetOrientation = "LANDSCAPE"
	MediaAssetOrientationPortrait  MediaAssetOrientation = "PORTRAIT"
	MediaAssetOrientationUnknown   MediaAssetOrientation = "UNKNOWN"
)

// AppAsset is the app asset associated with an adam ID.
type AppAsset struct {
	AdamID           *string                `json:"adamId,omitempty"`
	AppPreviewDevice map[string]string      `json:"appPreviewDevice,omitempty"`
	AssetGenID       *string                `json:"assetGenId,omitempty"`
	AssetType        *MediaAssetType        `json:"assetType,omitempty"`
	AssetURL         *string                `json:"assetURL,omitempty"`
	AssetVideoURL    *string                `json:"assetVideoUrl,omitempty"`
	Deleted          *bool                  `json:"deleted,omitempty"`
	Orientation      *MediaAssetOrientation `json:"orientation,omitempty"`
	SourceHeight     *int                   `json:"sourceHeight,omitempty"`
	SourceWidth      *int                   `json:"sourceWidth,omitempty"`
}
