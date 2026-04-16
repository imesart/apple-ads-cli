package types

// AppInfo is the response to an app search request.
type AppInfo struct {
	AdamID               int64    `json:"adamId"`
	AppName              string   `json:"appName"`
	DeveloperName        string   `json:"developerName"`
	CountryOrRegionCodes []string `json:"countryOrRegionCodes"`
}

// AppDetails is the media detail object for an app.
type AppDetails struct {
	ID                   string        `json:"id"`
	AdamID               int64         `json:"adamId"`
	AppName              string        `json:"appName"`
	ArtistName           string        `json:"artistName"`
	AvailableStorefronts []string      `json:"availableStorefronts"`
	DeviceClasses        []DeviceClass `json:"deviceClasses"`
	IconPictureURL       string        `json:"iconPictureUrl"`
	IsPreOrder           bool          `json:"isPreOrder"`
	PrimaryGenre         string        `json:"primaryGenre"`
	SecondaryGenre       string        `json:"secondaryGenre"`
	PrimaryLanguage      string        `json:"primaryLanguage"`
}

// EligibilityState is the state of the app eligibility review process.
type EligibilityState string

const (
	EligibilityStateEligible   EligibilityState = "ELIGIBLE"
	EligibilityStateIneligible EligibilityState = "INELIGIBLE"
)

// EligibilityRecord contains app eligibility parameters.
type EligibilityRecord struct {
	AdamID          int64            `json:"adamId"`
	CountryOrRegion string           `json:"countryOrRegion"`
	DeviceClass     DeviceClass      `json:"deviceClass"`
	MinAge          int              `json:"minAge"`
	State           EligibilityState `json:"state"`
	SupplySource    SupplySource     `json:"supplySource"`
}
