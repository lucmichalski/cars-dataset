package main

import (
	"log"
	"fmt"
	"os"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qor/media"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"

	"github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/crawler"
)

func main() {

	DB, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4,utf8&parseTime=True", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE")))
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// callback for images and validation
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)

	// migrate tables
	DB.AutoMigrate(&models.Vehicle{})
	DB.AutoMigrate(&models.VehicleImage{})

	cfg := &config.Config{
		AllowedDomains: []string{"motorcycles.autotrader.com"},
		URLs: []string{
			"https://motorcycles.autotrader.com/motorcycles/2019/bmw/c400x/200865678",
			// "https://www.motorcycles.autotrader.com/cars-for-sale/vehicledetails.xhtml?listingId=523174395&referrer=%2Fcars-for-sale%2Fsearchresults.xhtml%3FlistingTypes%3DNEW%26startYear%3D2018%26sortBy%3DderivedpriceDESC%26incremental%3Dall%26firstRecord%3D0%26marketExtension%3Dinclude%26endYear%3D2021%26makeCodeList%3DBMW%26isNewSearch%3Dtrue&listingTypes=NEW&startYear=2018&numRecords=25&firstRecord=0&endYear=2021&makeCodeList=BMW&clickType=spotlight",
		},
		DB: 			 DB,
		CacheDir:        "../../../shared/data",
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
		DryMode:         false,
		IsDebug:         true,
	}

	err = crawler.Extract(cfg)
	if err != nil {
		log.Fatal(err)
	}

}

