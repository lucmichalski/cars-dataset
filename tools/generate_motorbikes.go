package main

import (
	"encoding/json"
	"fmt"
	"os"
	"image"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"io/ioutil"
	"strings"
	"strconv"

	"github.com/h2non/filetype"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qor/media"
	"github.com/qor/validations"
	// "github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"

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

	var count cnt
	DB.Raw("select count(id) as count FROM vehicles WHERE class='motorcycle'").Scan(&count)

	// instanciate throttler
	t := throttler.New(12, count.Count)

	counter := 0
	imgCounter := 0

	var results []res
	DB.Raw("select name, manufacturer as make, modl, year, images FROM vehicles WHERE class='motorcycle' LIMIT 10000").Scan(&results)
	for _, result := range results {

		go func(r res) error {
			defer t.Done(nil)

			if r.Images == "" {
				return nil
			}

			var ep []entryProperty
			if err := json.Unmarshal([]byte(r.Images), &ep); err != nil {
				log.Fatalln("unmarshal error, ", err)
			}

			// prefixPath := filepath.Join("./", "datasets", "cars", result.Name)
			prefixPath := filepath.Join("./", "datasets", "motorbikes")
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

				sourceFile := filepath.Join("./", "public", entry.Url)
				input, err := ioutil.ReadFile(sourceFile)
				if err != nil {
					log.Warnln("reading file error, ", err)
					continue
				}

				destinationFile := filepath.Join(prefixPath, vi.Checksum+filepath.Ext(entry.Url))
				err = ioutil.WriteFile(destinationFile, input, 0644)
				if err != nil {
					// return err
					log.Fatalln("creating file error, ", err)
				}

				kind, _ := filetype.Match(input)

				var src image.Image
				//if isVerbose {
				log.Println("kind.MIME.Value:", kind.MIME.Value)
				//}

				// var file *os.File 
			    file, err := os.Open(destinationFile)
			    check(err)

				switch kind.MIME.Value {
				case "image/jpeg":
					src, err = jpeg.Decode(file)
					if err != nil {
						// fmt.Println("jpeg.Decode failed")					
						// panic(err.Error())
						continue
					}
				case "image/png":
					src, err = png.Decode(file)
					if err != nil {
						// fmt.Println("png.Decode failed")
						// panic(err.Error())
						continue
					}
				default:
					log.Fatalln("unkown mime type")
				}

				b := src.Bounds()

				// create annotation file for labelme
				lc := &models.Labelme{}
				lc.FillColor = []int{255, 0, 0, 128}

				lc.ImagePath = vi.Checksum+filepath.Ext(entry.Url)
				//lc.ImageData = nil

				// get image size
				lc.ImageHeight = b.Max.Y
				lc.ImageWidth = b.Max.X

				lc.LineColor = []int{0, 255, 0, 128}
				lc.Version = "3.6.10"				

				sc := models.Shape{}
				sc.Label = "motorcycle"
				sc.ShapeType = "rectangle"

				bboxParts := strings.Split(vi.BBox, ",")
				var maxX, maxY, minX, minY int
				maxX, _ = strconv.Atoi(bboxParts[0])
				maxY, _ = strconv.Atoi(bboxParts[1])
				minX, _ = strconv.Atoi(bboxParts[2])
				minY, _ = strconv.Atoi(bboxParts[3])

				fmt.Println("minX=", minX, "maxX=", maxX, "minY=", minY, "maxY=", maxY)

				x := []int{int(maxX), int(maxY)}
				y := []int{int(minX), int(minY)}

				points := [][]int{x, y}
				sc.Points = append(sc.Points, points...)
				lc.Shapes = append(lc.Shapes, sc)

				jsonByte, err := json.Marshal(lc) 
				if err != nil {
					log.Fatalln("json marshall: ", err)
				}

				jsonString := string(jsonByte)
				jsonString = strings.Replace(jsonString, "imageData\":\"\"", "imageData\": null", -1)

				// write json file
				jsonFile := filepath.Join(prefixPath, vi.Checksum+".json")
				err = ioutil.WriteFile(jsonFile, []byte(jsonString), 0644)
				if err != nil {
					// return err
					log.Fatalln("creating file error, ", err)
				}

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

func check(e error) {
    if e != nil {
        panic(e)
    }
}
