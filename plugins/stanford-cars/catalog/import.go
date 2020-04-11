package catalog

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	"github.com/qor/media/media_library"

	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
"image_type","image_relpath","name"
TRAIN,./cars_train/02443.jpg,HUMMER H3T Crew Cab 2010
TRAIN,./cars_train/02444.jpg,Ford F-150 Regular Cab 2012
TRAIN,./cars_train/02445.jpg,Buick Rainier SUV 2007
*/

type imageSrc struct {
	URL      string
	Size     int64
	File     string
	Checksum string
	Type     string
}

func ImportFromURL(cfg *config.Config) error {
	fmt.Printf("Import csv from %s\n", cfg.CatalogURL)
	file, size, checksum, err := utils.OpenFileByURL(cfg.CatalogURL)
	if err != nil {
		return err
	}
	fmt.Printf("Inspect remote csv for '%s', stored at '%s', size='%d', checksum='%s'\n", cfg.CatalogURL, file.Name(), size, checksum)

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.LazyQuotes = true
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}

	t := throttler.New(32, len(data))

	for _, row := range data {

		go func(row []string) error {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)

			vehicle := models.Vehicle{}
			vehicle.Source = "stanford-cars"

			name := row[2]
			nameParts := strings.Split(name, " ")

			vehicle.Year = nameParts[len(nameParts)-1]
			vehicle.Manufacturer = nameParts[0]
			model := strings.Replace(name, nameParts[0], "", -1)
			model = strings.Replace(model, vehicle.Year, "", -1)
			vehicle.Modl = model

			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("gid = ?", vehicle.Gid).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping gid=%s as already exists\n", vehicle.Gid)
					return nil
				}
			}

			file, size, checksum, err := utils.OpenFileByURL(row[1])
			if err != nil {
				fmt.Printf("open file failure, got err %v", err)
				file.Close()
				return err
			}

			if size == 0 {
				file.Close()
				return nil
			}

			image := models.VehicleImage{Title: name, SelectedType: "image", Checksum: checksum}

			log.Println("----> Scanning file: ", file.Name())
			image.File.Scan(file)

			if !cfg.DryMode {
				if err := cfg.DB.Create(&image).Error; err != nil {
					log.Printf("create variation_image (%v) failure, got err %v\n", image, err)
					return err
				}
			}

			vehicle.Images.Files = append(vehicle.Images.Files, media_library.File{
				ID:  json.Number(fmt.Sprint(image.ID)),
				Url: image.File.URL(),
			})

			if len(vehicle.MainImage.Files) == 0 {
				vehicle.MainImage.Files = []media_library.File{{
					ID:  json.Number(fmt.Sprint(image.ID)),
					Url: image.File.URL(),
				}}
			}
			file.Close()

			pp.Println(vehicle)
			if !cfg.DryMode {
				if err := cfg.DB.Create(&vehicle).Error; err != nil {
					log.Printf("create product (%v) failure, got err %v", vehicle, err)
					return err
				}
			}
			return nil

		}(row)
		t.Throttle()

	}

	// throttler errors iteration
	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			log.Printf("error #%d: %s", i, err)
		}
		log.Fatal(t.Err())
	}

	return nil
}
