package types

// TargetingDimensions contains optional criteria to narrow the audience that views ads.
type TargetingDimensions struct {
	Age            *AgeCriteria           `json:"age,omitempty"`
	Gender         *GenderCriteria        `json:"gender,omitempty"`
	AppCategories  *AppCategoryCriteria   `json:"appCategories,omitempty"`
	AppDownloaders *AppDownloaderCriteria `json:"appDownloaders,omitempty"`
	Country        *CountryCriteria       `json:"country,omitempty"`
	AdminArea      *AdminAreaCriteria     `json:"adminArea,omitempty"`
	Locality       *LocalityCriteria      `json:"locality,omitempty"`
	Daypart        *DaypartCriteria       `json:"daypart,omitempty"`
	DeviceClass    *DeviceClassCriteria   `json:"deviceClass,omitempty"`
}

// AgeRange defines the minimum and maximum age for targeting.
type AgeRange struct {
	MinAge *int `json:"minAge,omitempty"`
	MaxAge *int `json:"maxAge,omitempty"`
}

// AgeCriteria targets users by age demographic.
type AgeCriteria struct {
	Included []AgeRange `json:"included,omitempty"`
	Excluded []AgeRange `json:"excluded,omitempty"`
}

// GenderCriteria targets users by gender demographic.
type GenderCriteria struct {
	Included []Gender `json:"included,omitempty"`
	Excluded []Gender `json:"excluded,omitempty"`
}

// DeviceClassCriteria targets users by device type.
type DeviceClassCriteria struct {
	Included []DeviceClass `json:"included,omitempty"`
	Excluded []DeviceClass `json:"excluded,omitempty"`
}

// CountryCriteria targets users by country or region using ISO alpha-2 country codes.
type CountryCriteria struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// AdminAreaCriteria targets users by administrative area (state or equivalent).
type AdminAreaCriteria struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// LocalityCriteria targets users by locality (city or equivalent).
type LocalityCriteria struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// AppCategoryCriteria targets users by app category.
// A value of 100 indicates targeting apps with the same category as your app.
type AppCategoryCriteria struct {
	Included []int `json:"included,omitempty"`
	Excluded []int `json:"excluded,omitempty"`
}

// AppDownloaderCriteria targets users based on app downloads.
// Use the adamId of the app you're promoting as an included or excluded value.
type AppDownloaderCriteria struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// DaypartCriteria targets users by a specific time of day.
type DaypartCriteria struct {
	UserTime *DaypartDetail `json:"userTime,omitempty"`
}

// DaypartDetail defines hours of the week for daypart targeting.
// Numbers 0-167 represent hours of the week beginning at Sunday 12:00 midnight.
type DaypartDetail struct {
	Included []int `json:"included,omitempty"`
	Excluded []int `json:"excluded,omitempty"`
}
