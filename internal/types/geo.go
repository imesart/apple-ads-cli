package types

// GeolocationEntity is the type of geography for targeting locations.
type GeolocationEntity string

const (
	GeolocationEntityCountry   GeolocationEntity = "Country"
	GeolocationEntityAdminArea GeolocationEntity = "AdminArea"
	GeolocationEntityLocality  GeolocationEntity = "Locality"
)

// GeolocationRequest is the geosearch request object.
type GeolocationRequest struct {
	ID     string            `json:"id"`
	Entity GeolocationEntity `json:"entity"`
}

// GeolocationSearchEntity contains geolocation details including the geoidentifier and entity type.
type GeolocationSearchEntity struct {
	ID              string            `json:"id"`
	Entity          GeolocationEntity `json:"entity"`
	DisplayName     string            `json:"displayName"`
	CountryOrRegion *string           `json:"countryOrRegion,omitempty"`
	AdminArea       *string           `json:"adminArea,omitempty"`
	Locality        *string           `json:"locality,omitempty"`
}
