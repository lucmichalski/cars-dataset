package main

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qor/media"
	"github.com/qor/validations"
	"github.com/k0kubun/pp"

	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/models"
)

func main() {

	// Instanciate DB
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

	// select first 250k entries, then fetch the image ids and re-process them for getting correct md5, bbox coordinates.
	// Scan
	type cnt struct {
		Count int
	}

	type res struct {
		Name   string
		Make   string
		Modl   string
		Year   string
		Images string
	}

	type entryProperty struct {
		ID          int
		Url         string
		VideoLink   string
		FileName    string
		Description string
	}

	type imgFile struct {
		Description  string `json:"Description"`
		FileName     string `json:"FileName"`
		SelectedType string `json:"SelectedType"`
		URL          string `json:"Url"`
		Video        string `json:"Video"`		
	}

	

	var results []models.VehicleImage
	// wrong, it is vehicle_images to select the first 250k
	DB.Raw("select * FROM vehicle_images WHERE id>0 and id<250000").Scan(&results)
	for _, result := range results {

		//var imgfile imgFile
		pp.Println("file: ", result.File.String())
		//if err := json.Unmarshal([]byte(result.File.String()), &imgfile); err != nil {
		//	log.Fatalln("unmarshal error, ", err)
		//}
		imageURL := result.File.String()

		pp.Println("Checksum:", result.Checksum, "BBox:", result.BBox, "Source:", result.Source, "imgfile", imageURL)

		if imageURL == "" {
			continue
		}

		imageURL = fmt.Sprintf("http://51.91.21.67:9008%s", imageURL)
                log.Println("imageURL:", imageURL)
		proxyURL := fmt.Sprintf("http://51.91.21.67:9007/labelme?url=%s", imageURL)
		log.Println("proxyURL:", proxyURL)
		if content, err := utils.GetJSON(proxyURL); err != nil {
			fmt.Printf("open file failure, got err %v", err)
		} else {

			if string(content) == "" {
				continue
			}

			var detection *models.Labelme
			pp.Println(string(content))
			if err := json.Unmarshal(content, &detection); err != nil {
				log.Warnln("unmarshal error, ", err)
				continue
			}

			_, checksum, err := utils.DecodeToFile(imageURL, detection.ImageData)
			if err != nil {
				log.Fatalln("decodeToFile error, ", err)
			}

			if len(detection.Shapes) != 1 {
				continue
			}

			// we expect online one focused image
			maxX := detection.Shapes[0].Points[0][0]
			maxY := detection.Shapes[0].Points[0][1]
			minX := detection.Shapes[0].Points[1][0]
			minY := detection.Shapes[0].Points[1][1]
			bbox := fmt.Sprintf("%d,%d,%d,%d", maxX, maxY, minX, minY)
			// image := models.VehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum, Source: carImage, BBox: bbox}

			// DB.First(&result)
			result.Checksum = checksum
			result.BBox = bbox
			// DB.Save(&result)
			pp.Println(result)

		}

	}

	os.Exit(0)

}

