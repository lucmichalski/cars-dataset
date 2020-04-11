package catalog

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	"github.com/qor/media/media_library"
	"github.com/karrick/godirwalk"

	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
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

	csvMap := make(map[int]string, 0)

	t := throttler.New(32, len(data))

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

			// var imageURLs []string
			// imageSrcs := make([]*imageSrc, 0)

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

			if !cfg.DryMode {
				var vehicleExists models.Vehicle
				if !cfg.DB.Where("gid = ?", vehicle.Gid).First(&vehicleExists).RecordNotFound() {
					fmt.Printf("skipping gid=%s as already exists\n", vehicle.Gid)
					return nil
				}
			}

			imageSrcs, err := walkImages(vehicle.Gid, cfg.ImageDir)
			if err != nil {
				log.Fatal(err)
			}

			for i, imgSrc := range imageSrcs {
				if file, size, checksum, err := utils.OpenFileByURL(imgSrc.URL); err != nil {
					fmt.Printf("open file failure, got err %v", err)
					file.Close()
					continue
				} else {
					imageSrcs[i].Size = size
					imageSrcs[i].Checksum = checksum
					imageSrcs[i].File = file.Name()
				}
			}

			sort.Slice(imageSrcs[:], func(i, j int) bool {
				return imageSrcs[i].Size > imageSrcs[j].Size
			})

			pp.Println("imageSrcs:", imageSrcs)

			for _, imgSrc := range imageSrcs {
				if file, _, checksum, err := utils.OpenFileByURL(imgSrc.URL); err != nil {
					fmt.Printf("open file failure, got err %v", err)
					file.Close()
					continue
				} else {

					if imgSrc.Size == 0 {
						file.Close()
						continue
					}

					image := models.VehicleImage{Title: vehicle.Manufacturer + " " + vehicle.Modl, SelectedType: "image"}

					var imageExists models.VehicleImage
					if !cfg.DB.Where("checksum = ?", checksum).First(&imageExists).RecordNotFound() {
						fmt.Printf("skipping checksum=%s as already exists\n", checksum)
						image.ID = imageExists.ID
						image.CreatedAt = imageExists.CreatedAt
						continue
					} else {

						log.Println("----> Scanning file: ", file.Name())
						image.File.Scan(file)

						if !cfg.DryMode {
							if err := cfg.DB.Create(&image).Error; err != nil {
								log.Printf("create variation_image (%v) failure, got err %v\n", image, err)
								// return err
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

func walkImages(gid, dirname string) (list []*imageSrc, err error ){
	err = godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				if strings.Contains(osPathname, gid) {
					fmt.Println("found ", osPathname)
					list = append(list, &imageSrc{URL: "file://"+osPathname, Type: "image_link"})
					// list = append(list, "file://"+osPathname)
				}
			}
			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})
	return 
}
