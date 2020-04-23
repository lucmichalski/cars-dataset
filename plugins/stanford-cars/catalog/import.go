package catalog

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"os"
	"io"
	"io/ioutil"
	"mime/multipart"	
	"net/http"
	"bytes"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/h2non/filetype"
	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	"github.com/qor/media/media_library"
	"github.com/karrick/godirwalk"

	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/* 
	- snippets
	  - cd plugins/stanford-cars && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-stanford-cars.so ; cd ../..

	- CSV excerpt
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

type carInfo struct {
	name string
	model string
	year string
	make string
	imgs []string
}

func ImportFromURL(cfg *config.Config) error {

	file, err := os.Open(cfg.CatalogURL)
    if err != nil {
            return err
    }

	reader := csv.NewReader(file)
	reader.Comma = ','
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

		name := row[2]
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
				name: row[2],
				make: nameParts[0],
				model: strings.TrimSpace(model),
				year: nameParts[len(nameParts)-1],
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

			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("name = ? AND year = ? AND manufacturer = ? AND source = ?", vehicle.Name, vehicle.Year, vehicle.Manufacturer, vehicle.Source).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping name=%s,year=%s,manufacturer=%s,source=%s as already exists\n", vehicle.Name, vehicle.Year, vehicle.Manufacturer, vehicle.Source)
					return nil
				}
			}

			for _, imgSrc := range row.imgs {

				// create temporary file

				tmpfilePath := filepath.Join(os.TempDir(), path.Base(imgSrc))
				file, err := os.Create(tmpfilePath)
				if err != nil {
					log.Fatal("Create tmpfilePath", err)
					return err
				}

				pp.Println("tmpfilePath", tmpfilePath)
				pp.Println("imgSrc", imgSrc)

				if _, err := os.Stat(imgSrc); err != nil {
					continue
				}

				// make request to darknet service
				request, err := newfileUploadRequest("http://localhost:9005/crop", nil, "file", imgSrc)
				if err != nil {
					log.Fatalln("newfileUploadRequest", err)
				}
				client := &http.Client{}
				resp, err := client.Do(request)
				if err != nil {
					log.Fatalln("client.Do", err)
				} else {
					defer resp.Body.Close()

					_, err = io.Copy(file, resp.Body)
					if err != nil {
						log.Fatal("io.Copy", err)
						return err
					}

					buf, _ := ioutil.ReadFile(file.Name())
					kind, _ := filetype.Match(buf)
					pp.Println("kind: ", kind)

					fi, err := file.Stat()
					if err != nil {
						log.Fatal("file.Stat()", err)
						return err
					}

					size := fi.Size()

					checksum, err := utils.GetMD5File(tmpfilePath)
					if err != nil {
						log.Fatal("GetMD5File", err)
						continue
					}

					if size == 0 {
						file.Close()
						log.Warnln("image to small")
						continue
					}

					image := models.VehicleImage{Title: row.name, SelectedType: "image", Checksum: checksum, Source: imgSrc}

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

func walkImagesSlice(gid string, dirnames []string) (list []string, err error ){
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

func walkImages(gid string, dirnames []string) (list []*imageSrc, err error ){
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
