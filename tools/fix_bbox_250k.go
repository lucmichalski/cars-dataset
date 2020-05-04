package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nozzle/throttler"
	"github.com/qor/media"
	"github.com/qor/validations"

	"github.com/lucmichalski/cars-dataset/pkg/models"
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

	// instanciate throttler
	t := throttler.New(48, count.Count)

	counter := 0
	imgCounter := 0

	var results []res
	DB.Raw("select name, manufacturer as make, modl, year, images FROM vehicles WHERE class='car'").Scan(&results)
	for _, result := range results {

		go func(r res) error {
			defer t.Done(nil)

			if r.Images == "" {
				return nil
			}

			var ep []entryProperty
			// fmt.Println(result.Images)
			if err := json.Unmarshal([]byte(r.Images), &ep); err != nil {
				log.Fatalln("unmarshal error, ", err)
			}

			//if len(ep) < 2 {
			//      return nil
			//}

			// prefixPath := filepath.Join("./", "datasets", "cars", result.Name)
			prefixPath := filepath.Join("./", "datasets", "cars", strings.Replace(strings.ToUpper(r.Make), " ", "-", -1), strings.ToUpper(r.Modl), r.Year)
			os.MkdirAll(prefixPath, 0755)
			// pp.Println("prefixPath:", prefixPath)

			for _, entry := range ep {

				// get image Info (to test)
				var vi models.VehicleImage
				err := DB.First(&vi, entry.ID).Error
				if err != nil {
					log.Warnln("VehicleImage", err)
					continue
				}
				// fmt.Println("image checksum", vi.Checksum)

				sourceFile := filepath.Join("./", "public", entry.Url)
				// pp.Println("sourceFile:", sourceFile)

				input, err := ioutil.ReadFile(sourceFile)
				if err != nil {
					log.Warnln("reading file error, ", err)
					continue
				}

				destinationFile := filepath.Join(prefixPath, vi.Checksum+filepath.Ext(entry.Url))
				// destinationFile := filepath.Join(prefixPath, strconv.Itoa(entry.ID)+"-"+filepath.Base(entry.Url))
				err = ioutil.WriteFile(destinationFile, input, 0644)
				if err != nil {
					// return err
					log.Fatalln("creating file error, ", err)
				}
				// pp.Println("destinationFile:", destinationFile)

				csvDataset.Write([]string{r.Name, strings.Replace(strings.ToUpper(r.Make), " ", "-", -1), strings.ToUpper(r.Modl), r.Year, destinationFile})
				csvDataset.Flush()

				imgCounter++
			}

			percent := (counter * 100) / count.Count
			fmt.Printf("REF COUNTER=%d/%d (%.2f%), IMG COUNTER=%d\n", counter, count.Count, percent, imgCounter)
			counter++

			return nil

		}(result)

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

	os.Exit(0)

}
