package models

type Car struct {
	Description string    `json:"description"`
	ID          string    `json:"id"`
	Images      CarImages `json:"images"`
	IsExt       bool      `json:"isExt"`
	IsInt       bool      `json:"isInt"`
	Title       string    `json:"title"`
}

type CarImages struct {
	Huge   CarImagesHuge   `json:"huge"`
	Large  CarImagesLarge  `json:"large"`
	Medium CarImagesMedium `json:"medium"`
	Small  CarImagesSmall  `json:"small"`
	Thumb  CarImagesThumb  `json:"thumb"`
}

type CarImagesHuge struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type CarImagesLarge struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type CarImagesMedium struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type CarImagesSmall struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type CarImagesThumb struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type JSONLD struct {
	AtContext  string           `json:"@context"`
	AtID       string           `json:"@id"`
	AtType     string           `json:"@type"`
	Breadcrumb JSONLDBreadcrumb `json:"breadcrumb"`
	IsPartOf   JSONLDIsPartOf   `json:"isPartOf"`
	MainEntity JSONLDMainEntity `json:"mainEntity"`
	URL        string           `json:"url"`
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
	AtID string `json:"@id"`
	Name string `json:"name"`
}

type JSONLDIsPartOf struct {
	AtID   string `json:"@id"`
	AtType string `json:"@type"`
}

type JSONLDMainEntity struct {
	AtID             string                         `json:"@id"`
	AtType           string                         `json:"@type"`
	Brand            JSONLDMainEntityBrand          `json:"brand"`
	FuelEfficiency   JSONLDMainEntityFuelEfficiency `json:"fuelEfficiency"`
	Image            JSONLDMainEntityImage          `json:"image"`
	Model            string                         `json:"model"`
	Name             string                         `json:"name"`
	Offers           JSONLDMainEntityOffers         `json:"offers"`
	Review           JSONLDMainEntityReview         `json:"review"`
	URL              string                         `json:"url"`
	VehicleModelDate int                            `json:"vehicleModelDate"`
}

type JSONLDMainEntityBrand struct {
	AtID   string `json:"@id"`
	AtType string `json:"@type"`
}

type JSONLDMainEntityFuelEfficiency struct {
	AtType   string `json:"@type"`
	MaxValue int    `json:"maxValue"`
	MinValue int    `json:"minValue"`
	UnitText string `json:"unitText"`
}

type JSONLDMainEntityImage struct {
	AtID        string `json:"@id"`
	AtType      string `json:"@type"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
	Height      int    `json:"height"`
	UploadDate  string `json:"uploadDate"`
	URL         string `json:"url"`
	Width       int    `json:"width"`
}

type JSONLDMainEntityOffers struct {
	AtType        string `json:"@type"`
	Availability  string `json:"availability"`
	HighPrice     int    `json:"highPrice"`
	LowPrice      int    `json:"lowPrice"`
	PriceCurrency string `json:"priceCurrency"`
}

type JSONLDMainEntityReview struct {
	AtID           string                             `json:"@id"`
	AtType         string                             `json:"@type"`
	ArticleSection string                             `json:"articleSection"`
	Author         JSONLDMainEntityReviewAuthor       `json:"author"`
	DateModified   string                             `json:"dateModified"`
	DatePublished  string                             `json:"datePublished"`
	Description    string                             `json:"description"`
	Headline       string                             `json:"headline"`
	Image          JSONLDMainEntityReviewImage        `json:"image"`
	Keywords       string                             `json:"keywords"`
	Publisher      JSONLDMainEntityReviewPublisher    `json:"publisher"`
	ReviewRating   JSONLDMainEntityReviewReviewRating `json:"reviewRating"`
	URL            string                             `json:"url"`
}

type JSONLDMainEntityReviewAuthor struct {
	AtType   string `json:"@type"`
	Image    string `json:"image"`
	JobTitle string `json:"jobTitle"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}

type JSONLDMainEntityReviewImage struct {
	AtID        string `json:"@id"`
	AtType      string `json:"@type"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
	Height      int    `json:"height"`
	UploadDate  string `json:"uploadDate"`
	URL         string `json:"url"`
	Width       int    `json:"width"`
}

type JSONLDMainEntityReviewPublisher struct {
	AtID   string `json:"@id"`
	AtType string `json:"@type"`
}

type JSONLDMainEntityReviewReviewRating struct {
	AtType      string  `json:"@type"`
	BestRating  int     `json:"bestRating"`
	RatingValue float64 `json:"ratingValue"`
	WorstRating int     `json:"worstRating"`
}
