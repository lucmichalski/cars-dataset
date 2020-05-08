package main

import (
	"context"
	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
	"strings"
	"time"
)

// voiture-occasion-renault-kadjar-1-5-blue-dci-115ch-black--FVO16533323.htm

func handleStandalone (url string, collection *mongo.Collection) {
	var a Announce

	domain := "www.autoreflex.com"

	c := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.MaxDepth(1),
	)

	// Annouce URL
	a.URL = "http://" + domain + "/" + url

	// Announce ID
	c.OnHTML("section[star-id]", func(e *colly.HTMLElement) {
		announceId := e.Attr("star-id")
		//fmt.Println(announceId)
		a.AnnounceId = announceId
	})

	// Announce title
	c.OnHTML("div.header", func(e *colly.HTMLElement) {
		title := e.ChildText("h1")
		//log.Println(title)
		a.Title = title
	})

	// Price
	c.OnHTML("div.prix-container", func(e *colly.HTMLElement) {
		price := e.ChildText("div.prix")
		price = strings.TrimSuffix(price, "€")
		price = strings.Replace(price, " ", "", -1)
		n, _ := strconv.Atoi(price)
		//log.Println(n)
		a.Price = n
	})

	// Seller and phone
	c.OnHTML("div.contact", func(e *colly.HTMLElement) {
		/*garageName := e.ChildText("a[title]")
		log.Println(garageName)*/

		phone := e.ChildAttr("a[data-seller-telephone]","data-seller-telephone")
		//log.Println(phone)
		a.Phone = phone
	})

	// Seller name and address
	c.OnHTML("div.small-12.large-5.columns.mg-bottom", func (e *colly.HTMLElement) {
		garageName := e.ChildText("h3")
		//log.Println(garageName)
		a.GarageName = garageName

		garageAddress := e.ChildText("p")
		// Removes <br> tag
		garageAddress = strings.Replace(garageAddress, "\n", "", -1)
		// Removes the 56 spaces before the postal code (no idea why they've put 56 blanks here)
		garageAddress = strings.Replace(garageAddress, "                                                        ", " ", -1)
		//log.Println(garageAddress)
		a.GarageAddress = garageAddress
	})

	// Other attributes
	c.OnHTML("div.specs", func(e *colly.HTMLElement) {
		//km := e.ChildText("li")
		attr := e.ChildTexts("li")

		// After testing with numerous standalone pages, it seems that the order of the attributes doesn't change, so
		// even if it's dirty, we can rely on table indexes to extract attributes for this short example
		a.CarMileage = attr[0] // KM
		a.CarFirstRegistration = attr[1]
		a.AnnouncePostalCode = attr[2]
		a.CarEngine = attr[3]
		a.Transmission = attr[4]
		a.FiscalPower = attr[5]

		/*for i, j := range attr {
			log.Print(i, " ", j, " | ")
		}*/
	})


	if err := c.Visit("http://" + domain + "/" + url); err != nil {
		log.Printf("Error when visiting %s", "http://" + domain + "/" + url)
	} else {
		//log.Println(a)
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := collection.InsertOne(ctx, a)
		if err != nil {
			log.Panic(err)
		} else {
			//log.Println(res)
		}
	}

	//log.Println(url)
}
