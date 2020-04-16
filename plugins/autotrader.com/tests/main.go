package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/autotrader.com/crawler"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AllowedDomains: []string{"www.autotrader.com", "autotrader.com", "motorcycles.autotrader.com"},
		URLs: []string{
			"https://motorcycles.autotrader.com/motorcycles/2019/bmw/c400x/200865678",
			// "https://www.autotrader.com/cars-for-sale/vehicledetails.xhtml?listingId=523174395&referrer=%2Fcars-for-sale%2Fsearchresults.xhtml%3FlistingTypes%3DNEW%26startYear%3D2018%26sortBy%3DderivedpriceDESC%26incremental%3Dall%26firstRecord%3D0%26marketExtension%3Dinclude%26endYear%3D2021%26makeCodeList%3DBMW%26isNewSearch%3Dtrue&listingTypes=NEW&startYear=2018&numRecords=25&firstRecord=0&endYear=2021&makeCodeList=BMW&clickType=spotlight",
		},
		CacheDir:        "../../../shared/data",
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
		DryMode:         true,
		IsDebug:         true,
	}

	err := crawler.Extract(cfg)
	if err != nil {
		log.Fatal(err)
	}

}
