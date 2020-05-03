package models

type JSONLD struct {
	AtContext       string                `json:"@context"`
	AtID            string                `json:"@id"`
	AtType          string                `json:"@type"`
	Breadcrumb      JSONLDBreadcrumb      `json:"breadcrumb"`
	MainEntity      JSONLDMainEntity      `json:"mainEntity"`
	Name            string                `json:"name"`
	PotentialAction JSONLDPotentialAction `json:"potentialAction"`
	Publisher       JSONLDPublisher       `json:"publisher"`
	SameAs          []string              `json:"sameAs"`
	URL             string                `json:"url"`
}

type JSONLDBreadcrumb struct {
	AtType          string                            `json:"@type"`
	ItemListElement []JSONLDBreadcrumbItemListElement `json:"itemListElement"`
}

type JSONLDBreadcrumbItemListElement struct {
	AtType   string                              `json:"@type"`
	Item     JSONLDBreadcrumbItemListElementItem `json:"item"`
	Position int                                 `json:"position"`
}

type JSONLDBreadcrumbItemListElementItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type JSONLDMainEntity struct {
	AtID             string                         `json:"@id"`
	AtType           string                         `json:"@type"`
	Brand            JSONLDMainEntityBrand          `json:"brand"`
	FuelEfficiency   JSONLDMainEntityFuelEfficiency `json:"fuelEfficiency"`
	Image            interface{}                    `json:"image"`
	Model            string                         `json:"model"`
	Name             string                         `json:"name"`
	Offers           JSONLDMainEntityOffers         `json:"offers"`
	URL              string                         `json:"url"`
	VehicleModelDate string                         `json:"vehicleModelDate"`
}

type JSONLDMainEntityBrand struct {
	AtID   string `json:"@id"`
	AtType string `json:"@type"`
	Name   string `json:"name"`
}

type JSONLDMainEntityFuelEfficiency struct {
	AtType   string `json:"@type"`
	MaxValue int    `json:"maxValue"`
	MinValue int    `json:"minValue"`
	UnitText string `json:"unitText"`
}

type JSONLDMainEntityOffers struct {
	AtType        string `json:"@type"`
	Availability  string `json:"availability"`
	HighPrice     int    `json:"highPrice"`
	LowPrice      int    `json:"lowPrice"`
	PriceCurrency string `json:"priceCurrency"`
}

type JSONLDPotentialAction struct {
	AtType     string                          `json:"@type"`
	QueryInput JSONLDPotentialActionQueryInput `json:"query-input"`
	Target     JSONLDPotentialActionTarget     `json:"target"`
}

type JSONLDPotentialActionQueryInput struct {
	AtType        string `json:"@type"`
	ValueName     string `json:"valueName"`
	ValueRequired string `json:"valueRequired"`
}

type JSONLDPotentialActionTarget struct {
	AtType      string `json:"@type"`
	URLTemplate string `json:"urlTemplate"`
}

type JSONLDPublisher struct {
	AtContext string               `json:"@context"`
	AtID      string               `json:"@id"`
	AtType    string               `json:"@type"`
	Brand     JSONLDPublisherBrand `json:"brand"`
	Logo      JSONLDPublisherLogo  `json:"logo"`
	Name      string               `json:"name"`
	URL       string               `json:"url"`
}

type JSONLDPublisherBrand struct {
	AtType string `json:"@type"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}

type JSONLDPublisherLogo struct {
	AtID   string `json:"@id"`
	AtType string `json:"@type"`
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}
