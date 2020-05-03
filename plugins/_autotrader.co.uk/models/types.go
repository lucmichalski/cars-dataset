package models

type VehicleGtm struct {
	ProductBrand        string `json:"ProductBrand"`
	ProductDistance     int    `json:"ProductDistance"`
	ProductFuel         string `json:"ProductFuel"`
	ProductKilometrage  string `json:"ProductKilometrage"`
	ProductModele       string `json:"ProductModele"`
	ProductPrice        string `json:"ProductPrice"`
	ProductTransmission string `json:"ProductTransmission"`
	ProductYear         string `json:"ProductYear"`
	Event               string `json:"event"`
	ID                  string `json:"id"`
}
