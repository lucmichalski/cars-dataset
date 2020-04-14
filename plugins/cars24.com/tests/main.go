package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/cars24.com/crawler"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AllowedDomains: []string{"www.cars24.com", "cars24.com"},
		URLs: []string{
			"https://www.cars24.com/buy-used-Tata-Tiago-2017-cars-Noida-1009527588/",
		},
	//	CacheDir:        "../../../shared/data",
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
