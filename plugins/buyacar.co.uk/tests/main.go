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

	"github.com/lucmichalski/cars-contrib/buyacar.co.uk/crawler"
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
		AllowedDomains: []string{"www.buyacar.co.uk", "buyacar.co.uk"},
		URLs: []string{
			"https://www.buyacar.co.uk/abarth/595/595-hatchback/1-4-t-jet-180-competizione-3dr-75464/deal-3112001",
		},
		DB: 			 DB,
		CacheDir:        "../../../shared/data",
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
		DryMode:         true,
		IsDebug:         true,
	}

	err = crawler.Extract(cfg)
	if err != nil {
		log.Fatal(err)
	}

}
