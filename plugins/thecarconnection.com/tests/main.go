package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/thecarconnection.com/crawler"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AllowedDomains: []string{"www.thecarconnection.com", "thecarconnection.com"},
		URLs: []string{
			"https://www.thecarconnection.com/overview/jaguar_f-type_2021",
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
