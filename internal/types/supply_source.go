package types

// SupplySource is the ad placement for a campaign.
type SupplySource string

const (
	SupplySourceAppStoreProductPagesBrowse SupplySource = "APPSTORE_PRODUCT_PAGES_BROWSE"
	SupplySourceAppStoreSearchResults      SupplySource = "APPSTORE_SEARCH_RESULTS"
	SupplySourceAppStoreSearchTab          SupplySource = "APPSTORE_SEARCH_TAB"
	SupplySourceAppStoreTodayTab           SupplySource = "APPSTORE_TODAY_TAB"
)
