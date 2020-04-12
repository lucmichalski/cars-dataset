package catalog

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"os"
	"bytes"
	"io"
	"path/filepath"
	"mime/multipart"
	"net/http"
	"path"
	"io/ioutil"

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
		- cd plugins/carvana-kaggle && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-carvana-kaggle.so ; cd ../..

	- CSV exceprt	
		"id","year","make","model","trim1","trim2"
		"0004d4463b50","2014","Acura","TL","TL","w/SE"
		"00087a6bd4dc","2014","Acura","RLX","RLX","w/Tech"
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
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}

	csvMap := make(map[int]string, 0)

	t := throttler.New(1, len(data))

	for idx, row := range data {
		// skip header
		if idx == 0 {
			for i, header := range row {
				csvMap[i] = header
			}
			continue
		}

		go func(row []string) error {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)

			vehicle := models.Vehicle{}
			vehicle.Source = "carvana-kaggle"
			for id, header := range csvMap {
				switch header {
				case "id":
					vehicle.Gid = row[id]
				case "year":
					vehicle.Year = row[id]
				case "make":
					vehicle.Manufacturer = row[id]
				case "model":
					vehicle.Modl = row[id]
				case "trim2":
					vehicle.Engine = row[id]
				}
			}

			vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year

			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("gid = ?", vehicle.Gid).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping gid=%s as already exists\n", vehicle.Gid)
					return nil
				}
			}

			pp.Println("vehicle.Gid: ", vehicle.Gid)
			pp.Println("cfg.ImageDirs: ", cfg.ImageDirs)
			imageSrcs, err := walkImages(vehicle.Gid, cfg.ImageDirs)
			if err != nil {
				log.Fatal(err)
			}

			pp.Println(vehicle)
			pp.Println(imageSrcs)

			for i, imgSrc := range imageSrcs {

                file, err := os.Open(imgSrc.URL)
                if err != nil {
                        return err
                }

                fi, err := file.Stat()
                if err != nil {
                        return err
                }

                size := fi.Size()
                checksum, err := utils.GetMD5File(file.Name())
                if err != nil {
                        return err
                }

				imageSrcs[i].Size = size
				imageSrcs[i].Checksum = checksum
				imageSrcs[i].File = file.Name()

				file.Close()

			}

			sort.Slice(imageSrcs[:], func(i, j int) bool {
				return imageSrcs[i].Size > imageSrcs[j].Size
			})

			pp.Println("imageSrcs:", imageSrcs)

			for _, imgSrc := range imageSrcs {

				tmpfilePath := filepath.Join(os.TempDir(), path.Base(imgSrc.URL))
				file, err := os.Create(tmpfilePath)
				if err != nil {
					log.Fatal("Create tmpfilePath", err)
					continue
				}				

				// make request to darknet service
				request, err := newfileUploadRequest("http://localhost:9004/crop", nil, "file", imgSrc.URL)
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

					image := models.VehicleImage{Title: vehicle.Manufacturer + " " + vehicle.Modl, SelectedType: "image", Checksum: checksum}

					log.Println("----> Scanning file: ", file.Name())
					image.File.Scan(file)

					if !cfg.DryMode {
						if err := cfg.DB.Create(&image).Error; err != nil {
							log.Printf("create variation_image (%v) failure, got err %v\n", image, err)
							file.Close()
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
