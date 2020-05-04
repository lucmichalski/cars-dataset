package catalog

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	  - cd plugins/vmmrdb && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-vmmrdb.so ; cd ../..
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

		dirParts := strings.Split(row[0], "/")
		if len(dirParts) == 0 {
			continue
		}
                pp.Println(row[0])

		nameParts := strings.Split(dirParts[0], "_")

		if len(nameParts) != 3 {
			continue
		}

		make := nameParts[0]
		model := nameParts[1]
		year := nameParts[2]
		name := make + " " + model + " " + year

		var imageSrcs []string
		imageSrcs = append(imageSrcs, strings.Replace(row[0], " ", "%20", -1))

		if _, ok := cars[name]; ok {
			cars[name].imgs = append(cars[name].imgs, imageSrcs...)
		} else {
			car := &carInfo{
				name:  name,
				make:  make,
				model: strings.TrimSpace(model),
				year:  year,
			}
			car.imgs = append(car.imgs, imageSrcs...)
			cars[name] = car
		}
		// pp.Println(cars[name])

	}

	for _, row := range cars {

		go func(row *carInfo) error {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)

			vehicle := models.Vehicle{}
			vehicle.Source = "vmmrdb"

			vehicle.Modl = row.model
			vehicle.Name = row.name
			vehicle.Year = row.year
			vehicle.Manufacturer = row.make
			vehicle.Class = "car"

			pp.Println(vehicle)

			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("name = ? AND year = ? AND manufacturer = ? AND source = ?", vehicle.Name, vehicle.Year, vehicle.Manufacturer, vehicle.Source).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping name=%s,year=%s,manufacturer=%s,source=%s as already exists\n", vehicle.Name, vehicle.Year, vehicle.Manufacturer, vehicle.Source)
					return nil
				}
			}

			pp.Println("row.imgs", row.imgs)

			for _, imgSrc := range row.imgs {

				carImage := fmt.Sprintf("http://51.91.21.67:8882/%s", imgSrc)

				proxyURL := fmt.Sprintf("http://51.91.21.67:9007/labelme?url=%s", carImage)
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
