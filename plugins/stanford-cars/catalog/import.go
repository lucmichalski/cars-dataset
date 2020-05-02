package catalog

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	//"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	// "path"
	"path/filepath"
	"strings"

	// "github.com/h2non/filetype"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	"github.com/nozzle/throttler"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
	- snippets
	  - cd plugins/stanford-cars && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-stanford-cars.so ; cd ../..

	- converter
	  -

	- CSV excerpt
		"image_type","image_relpath","name"
		TRAIN,./cars_train/02443.jpg,HUMMER H3T Crew Cab 2010
		TRAIN,./cars_train/02444.jpg,Ford F-150 Regular Cab 2012
		TRAIN,./cars_train/02445.jpg,Buick Rainier SUV 2007
	- CSV excerpt #2
		TRAIN;/opt/cars_train/02443.jpg;(640, 480);74;62;617;411;HUMMER H3T Crew Cab 2010;(0.5398437500000001, 0.4927083333333333, 0.8484375000000001, 0.7270833333333333)
		TRAIN;/opt/cars_train/02444.jpg;(800, 576);70;60;737;541;Ford F-150 Regular Cab 2012;(0.504375, 0.5217013888888888, 0.83375, 0.8350694444444444)
		TRAIN;/opt/cars_train/02445.jpg;(786, 491);30;99;743;427;Buick Rainier SUV 2007;(0.4917302798982189, 0.5356415478615071, 0.9071246819338423, 0.6680244399185336)
*/

type imageSrc struct {
	URL      string
	Size     int64
	File     string
	Checksum string
	Type     string
}

type carInfo struct {
	name  string
	model string
	year  string
	make  string
	imgs  []string
}

func ImportFromURL(cfg *config.Config) error {

	file, err := os.Open(cfg.CatalogURL)
	if err != nil {
		return err
	}

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	t := throttler.New(1, 10000000)

	cars := make(map[string]*carInfo, 0)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err == csv.ErrFieldCount {
				continue
			}
			return err
		}

		//pp.Println(row)
		//os.Exit(1)

		name := row[7]
		nameParts := strings.Split(name, " ")
		year := nameParts[len(nameParts)-1]
		model := strings.Replace(name, nameParts[0], "", -1)
		model = strings.Replace(model, year, "", -1)

		var imageSrcs []string
		// imageSrcs = append(imageSrcs, "./shared/datasets/stanford-cars/cars_test/"+row[1])
		imageSrcs = append(imageSrcs, "./shared/datasets/stanford-cars/cars_train/"+row[1])

		if _, ok := cars[name]; ok {
			cars[name].imgs = append(cars[name].imgs, imageSrcs...)
		} else {
			car := &carInfo{
				name:  row[7],
				make:  nameParts[0],
				model: strings.TrimSpace(model),
				year:  nameParts[len(nameParts)-1],
			}
			car.imgs = append(car.imgs, imageSrcs...)
			cars[name] = car
		}
	}

	for _, row := range cars {

		go func(row *carInfo) error {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)

			vehicle := models.Vehicle{}
			vehicle.Source = "stanford-cars"

			vehicle.Modl = row.model
			vehicle.Name = row.name
			vehicle.Year = row.year
			vehicle.Manufacturer = row.make
			vehicle.Class = "car"

			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("name = ? AND year = ? AND manufacturer = ? AND source = ?", vehicle.Name, vehicle.Year, vehicle.Manufacturer, vehicle.Source).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping name=%s,year=%s,manufacturer=%s,source=%s as already exists\n", vehicle.Name, vehicle.Year, vehicle.Manufacturer, vehicle.Source)
					return nil
				}
			}

			for _, imgSrc := range row.imgs {

				carImage := fmt.Sprintf("http://51.91.21.67:8881/%s", imgSrc)
				carImage = strings.Replace(carImage, "/opt/", "", -1)
                                carImage = strings.Replace(carImage, "./shared/datasets/stanford-cars/cars_train/", "", -1)

				proxyURL := fmt.Sprintf("http://51.91.21.67:9003/labelme?url=%s", carImage)
				log.Println("proxyURL:", proxyURL)
				if content, err := utils.GetJSON(proxyURL); err != nil {
					fmt.Printf("open file failure, got err %v", err)
				} else {

					if string(content) == "" {
						continue					
					}

					var detection *models.Labelme
					if err := json.Unmarshal(content, &detection); err != nil {
						log.Warnln("unmarshal error, ", err)
						continue
					}

					file, checksum, err := utils.DecodeToFile(carImage, detection.ImageData)
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
					image := models.VehicleImage{Title: vehicle.Manufacturer + " " + vehicle.Modl, SelectedType: "image", Checksum: checksum, Source: carImage, BBox: bbox}

					log.Println("----> Scanning file: ", file.Name())
					image.File.Scan(file)

					if !cfg.DryMode {
						if err := cfg.DB.Create(&image).Error; err != nil {
							log.Printf("create variation_image (%v) failure, got err %v\n", image, err)
							continue
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
				}
			}

			if !cfg.DryMode {
				pp.Println(vehicle)
				if err := cfg.DB.Create(&vehicle).Error; err != nil {
					log.Fatalf("create product (%v) failure, got err %v", vehicle, err)
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

func walkImagesSlice(gid string, dirnames []string) (list []string, err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, gid) {
						// fmt.Println("found ", osPathname, "gid", gid)
						list = append(list, osPathname)
						// list = append(list, "file://"+osPathname)
					}
				}
				return nil
			},
			Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
		})
	}
	return
}

func walkImages(gid string, dirnames []string) (list []*imageSrc, err error) {
	for _, dirname := range dirnames {
		err = godirwalk.Walk(dirname, &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				if !de.IsDir() {
					if strings.Contains(osPathname, gid) {
						fmt.Println("found ", osPathname, "gid", gid)
						list = append(list, &imageSrc{URL: osPathname, Type: "image_link"})
						// list = append(list, "file://"+osPathname)
					}
				}
				return nil
			},
			Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
		})
	}
	return
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
