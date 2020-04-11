
package main

import (
	"log"

	"github.com/lucmichalski/cars-contrib/carvana-kaggle/catalog"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

func main() {

	cfg := &config.Config{
		AnalyzerURL: 	"http://localhost:9003/crop?url=%s",
		CatalogURL: 	"../../../shared/datasets/carvana-kaggle/metadata.csv",
		ImageDirs:       []string{"../../../shared/datasets/carvana-kaggle/train_hq", "../../../shared/datasets/carvana-kaggle/test_hq"},
		DryMode:         true,
		IsDebug:         true,
	}

	err := catalog.ImportFromURL(cfg)
	if err != nil {
		log.Fatal(err)
	}

}
