package model

type GeoLocation struct {
	Latitude    float64 `bson:"latitude"`
	Longitude   float64 `bson:"longitude"`
	CountryCode string  `bson:"countryCode"`
	Country     string  `bson:"country"`
	City        string  `bson:"city"`
	Type        string  `bson:"type"`
}
