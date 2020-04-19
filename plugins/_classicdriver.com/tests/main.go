package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/classicdriver.com/crawler"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AllowedDomains: []string{"www.classicdriver.com", "classicdriver.com"},
		URLs: []string{
			// "https://www.classicdriver.com/en/car/ferrari/308/1974/388412",
			"https://www.classicdriver.com/en/sitemap.xml?page=1",
		},
	//	CacheDir:        "../../../shared/data",
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
		DryMode:         true,
		IsDebug:         true,
		IsSitemapIndex:  false,
	}

	err := crawler.Extract(cfg)
	if err != nil {
		log.Fatal(err)
	}

}
