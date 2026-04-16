package types

// ProductPageState is the system state of the custom product page.
type ProductPageState string

const (
	ProductPageStateHidden  ProductPageState = "HIDDEN"
	ProductPageStateVisible ProductPageState = "VISIBLE"
)

// ProductPageDetail contains the product page metadata.
type ProductPageDetail struct {
	ID               *string           `json:"id,omitempty"`
	AdamID           *int64            `json:"adamId,omitempty"`
	Name             *string           `json:"name,omitempty"`
	State            *ProductPageState `json:"state,omitempty"`
	DeepLink         *string           `json:"deepLink,omitempty"`
	CreationTime     *string           `json:"creationTime,omitempty"`
	ModificationTime *string           `json:"modificationTime,omitempty"`
}

// ProductPageLocaleDetail contains the product page locale metadata.
type ProductPageLocaleDetail struct {
	AdamID                     *int64         `json:"adamId,omitempty"`
	AppName                    *string        `json:"appName,omitempty"`
	DeviceClasses              []DeviceClass  `json:"deviceClasses,omitempty"`
	Language                   *string        `json:"language,omitempty"`
	LanguageCode               *string        `json:"languageCode,omitempty"`
	ProductPageID              *string        `json:"productPageId,omitempty"`
	PromotionalText            *string        `json:"promotionalText,omitempty"`
	ShortDescription           *string        `json:"shortDescription,omitempty"`
	SubTitle                   *string        `json:"subTitle,omitempty"`
	AppPreviewDeviceWithAssets map[string]any `json:"appPreviewDeviceWithAssets,omitempty"`
}
