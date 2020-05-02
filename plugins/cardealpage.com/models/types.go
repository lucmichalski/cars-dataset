package models

type JSONLD struct {
	AtContext                   string                        `json:"@context"`
	AtType                      string                        `json:"@type"`
	Color                       string                        `json:"color"`
	DriveWheelConfiguration     JSONLDDriveWheelConfiguration `json:"driveWheelConfiguration"`
	FuelEfficiency              JSONLDFuelEfficiency          `json:"fuelEfficiency"`
	FuelType                    string                        `json:"fuelType"`
	Image                       JSONLDImage                   `json:"image"`
	ItemCondition               JSONLDItemCondition           `json:"itemCondition"`
	ItemListElement             []JSONLDItemListElement       `json:"itemListElement"`
	Logo                        string                        `json:"logo"`
	Manufacturer                JSONLDManufacturer            `json:"manufacturer"`
	MileageFromOdometer         JSONLDMileageFromOdometer     `json:"mileageFromOdometer"`
	Model                       JSONLDModel                   `json:"model"`
	Name                        string                        `json:"name"`
	NumberOfDoors               int                           `json:"numberOfDoors"`
	Offers                      JSONLDOffers                  `json:"offers"`
	SameAs                      []string                      `json:"sameAs"`
	URL                         string                        `json:"url"`
	VehicleEngine               JSONLDVehicleEngine           `json:"vehicleEngine"`
	VehicleIdentificationNumber string                        `json:"vehicleIdentificationNumber"`
	VehicleInteriorColor        string                        `json:"vehicleInteriorColor"`
	VehicleModelDate            int                           `json:"vehicleModelDate"`
	VehicleSeatingCapacity      string                        `json:"vehicleSeatingCapacity"`
	VehicleTransmission         string                        `json:"vehicleTransmission"`
}

type JSONLDDriveWheelConfiguration struct {
	AtType string `json:"@type"`
	Name   string `json:"name"`
}

type JSONLDFuelEfficiency struct {
	AtType   string `json:"@type"`
	UnitCode string `json:"unitCode"`
	Value    string `json:"value"`
}

type JSONLDImage struct {
	AtType     string `json:"@type"`
	ContentURL string `json:"contentUrl"`
}

type JSONLDItemCondition struct {
	AtType string `json:"@type"`
	Name   string `json:"name"`
}

type JSONLDItemListElement struct {
	AtType   string                    `json:"@type"`
	Item     JSONLDItemListElementItem `json:"item"`
	Position int                       `json:"position"`
}

type JSONLDItemListElementItem struct {
	AtID string `json:"@id"`
	Name string `json:"name"`
}

type JSONLDManufacturer struct {
	AtType string `json:"@type"`
	Name   string `json:"name"`
}

type JSONLDMileageFromOdometer struct {
	AtType   string `json:"@type"`
	UnitCode string `json:"unitCode"`
	Value    int    `json:"value"`
}

type JSONLDModel struct {
	AtType      string `json:"@type"`
	IsVariantOf string `json:"isVariantOf"`
	Name        string `json:"name"`
}

type JSONLDOffers struct {
	AtType        string             `json:"@type"`
	Price         int                `json:"price"`
	PriceCurrency string             `json:"priceCurrency"`
	Seller        JSONLDOffersSeller `json:"seller"`
}

type JSONLDOffersSeller struct {
	AtType          string                            `json:"@type"`
	Address         JSONLDOffersSellerAddress         `json:"address"`
	AggregateRating JSONLDOffersSellerAggregateRating `json:"aggregateRating"`
	Name            string                            `json:"name"`
	Telephone       string                            `json:"telephone"`
}

type JSONLDOffersSellerAddress struct {
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
	StreetAddress   string `json:"streetAddress"`
}

type JSONLDOffersSellerAggregateRating struct {
	AtType      string  `json:"@type"`
	RatingValue float64 `json:"ratingValue"`
	ReviewCount int     `json:"reviewCount"`
}

type JSONLDVehicleEngine struct {
	AtType string `json:"@type"`
	Name   string `json:"name"`
}
