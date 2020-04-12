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

func ImportFromURL(cfg *config.Config) error {

	file, err := os.Open(cfg.CatalogURL)
    if err != nil {
            return err
    }

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.LazyQuotes = true

	t := throttler.New(1, 10000000)

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

		go func(row []string) error {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)

			vehicle := models.Vehicle{}
			vehicle.Source = "stanford-cars"

			name := row[2]
			nameParts := strings.Split(name, " ")
			vehicle.Name = row[2]
			vehicle.Year = nameParts[len(nameParts)-1]
			vehicle.Manufacturer = nameParts[0]
			model := strings.Replace(name, nameParts[0], "", -1)
			model = strings.Replace(model, vehicle.Year, "", -1)
			vehicle.Modl = strings.TrimSpace(model)

			/*
			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("gid = ?", vehicle.Gid).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping gid=%s as already exists\n", vehicle.Gid)
					return nil
				}
			}
			*/

			imageSrcs, err := walkImages(row[1], cfg.ImageDirs)
			if err != nil {
				log.Fatal(err)
			}

			pp.Println(vehicle)
			pp.Println(imageSrcs)

			if len(imageSrcs) < 1 {
				fmt.Println("check")
				os.Exit(1)
			}

			// create temporary file

			tmpfilePath := filepath.Join(os.TempDir(), path.Base(imageSrcs[0].URL))
			file, err := os.Create(tmpfilePath)
			if err != nil {
				log.Fatal("Create tmpfilePath", err)
				return err
			}

			pp.Println("tmpfilePath", tmpfilePath)
			pp.Println("imageSrcs[0].URL", imageSrcs[0].URL)

			// make request to darknet service
			request, err := newfileUploadRequest("http://localhost:9003/crop", nil, "file", imageSrcs[0].URL)
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
					return err
				}

				if size == 0 {
					file.Close()
					log.Warnln("image to small")
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
			}

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
