package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/stanford-cars/catalog"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
		CatalogURL:  "../../../shared/datasets/stanford-cars/data/yolo_cars_data.csv",
		DryMode:     true,
		IsDebug:     true,
	}

	err := catalog.ImportFromURL(cfg)
	if err != nil {
		log.Fatal(err)
	}

}
