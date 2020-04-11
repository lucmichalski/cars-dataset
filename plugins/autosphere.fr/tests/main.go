package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/autosphere.fr/crawler"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AllowedDomains: []string{"www.autosphere.fr", "autosphere.fr"},
		URLs: []string{
			"https://www.autosphere.fr/fiche4/auto-occasion-renault-megane-1-3-tce-140ch-fap-intens-120g-38200-vienne-95322?recobdme=reco",
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
