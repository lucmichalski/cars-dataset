package crawler

import (
	"encoding/json"
	"fmt"
	"strings"
	"os"

	"github.com/k0kubun/pp"
	"github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	// "github.com/iancoleman/strcase"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	// c.DisableCookies()
	c.OnHTML(`a[href]`, func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "fiches/index") {
			fmt.Printf("Link found: %s\n", e.Request.AbsoluteURL(link))
			q.AddURL(e.Request.AbsoluteURL(link))
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error:", err, r.Request.URL, r.StatusCode)
		q.AddURL(r.Request.URL.String())
	})

	c.OnHTML(`body`, func(e *colly.HTMLElement) {

		// check if we are processing a product page
		if !strings.Contains(e.Request.Ctx.Get("url"), "fiches/index") {
			return
		}

		// check in the databse if exists
		var vehicleExists models.Vehicle
		if !cfg.DB.Where("url = ?", e.Request.Ctx.Get("url")).First(&vehicleExists).RecordNotFound() {
			fmt.Printf("skipping url=%s as already exists\n", e.Request.Ctx.Get("url"))
			return
		}

		vehicle := &models.Vehicle{}
		vehicle.URL = e.Request.Ctx.Get("url")

		e.ForEach(`div[class=CLAnnonceCaracteristiques] table tr`, func(_ int, el *colly.HTMLElement) {
			var key, value string
			el.ForEach(`td:nth-child(1)`, func(_ int, eli *colly.HTMLElement) {
				key = eli.Text
			})

			el.ForEach(`td:nth-child(n+1)`, func(_ int, eli *colly.HTMLElement) {
				value = eli.Text
			})
			pp.Println("key:", key, "value=", value)
			// mapping key and values
			switch key {
			case "Marque":
				vehicle.Manufacturer = value
			case "Modèle":
				vehicle.Modl = value
			case "Année du modèle":
				vehicle.Year = value
			case "Catégorie":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Category", Value: strings.TrimSpace(value)})
			case "Cylindrée":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Displacement", Value: strings.TrimSpace(value)})
			case "Chassis":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Frame", Value: strings.TrimSpace(value)})
			case "Motorisation":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Motorization", Value: strings.TrimSpace(value)})
			case "Couleur":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Color", Value: strings.TrimSpace(value)})
			case "Référence":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Reference", Value: strings.TrimSpace(value)})
			case "Prix":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Price", Value: strings.TrimSpace(value)})
			}
		})

		vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year
		vehicle.Source = "yamaha-occasion.com"

		// Pictures
		var carImgLinks []string
		e.ForEach(`a.fancybox`, func(_ int, el *colly.HTMLElement) {
			carPicSrc := el.Attr("href")
			if cfg.IsDebug {
				if carPicSrc != "" {
					fmt.Println("carPicSrc:", carPicSrc)
				}
			}
			carImgLinks = append(carImgLinks, carPicSrc)
		})

		carImgLinks = utils.RemoveDuplicates(carImgLinks)
		if cfg.IsDebug {
			pp.Println(carImgLinks)
		}

		if vehicle.Manufacturer == "" && vehicle.Year == "" && vehicle.Modl == "" {
			return
		}

		if len(carImgLinks) == 0 {
			return
		}

		for _, carImage := range carImgLinks {
			if carImage == "" {
				continue
			}

			proxyURL := fmt.Sprintf("http://51.91.21.67:9003/crop?url=%s", carImage)
			log.Println("proxyURL:", proxyURL)
			if file, size, checksum, err := utils.OpenFileByURL(proxyURL); err != nil {
				fmt.Printf("open file failure, got err %v", err)
			} else {
				defer file.Close()

				if size < 40000 {
					if cfg.IsClean {
						// delete tmp file
						err := os.Remove(file.Name())
						if err != nil {
							log.Fatal(err)
						}
					}
					log.Infoln("----> Skipping file: ", file.Name(), "size: ", size)					
					continue
				}

				image := models.VehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum, Source: carImage}

				log.Println("----> Scanning file: ", file.Name(), "size: ", size)
				if err := image.File.Scan(file); err != nil {
					log.Fatalln("image.File.Scan, err:", err)
					continue
				}

				// transaction
				if !cfg.DryMode {
					if err := cfg.DB.Create(&image).Error; err != nil {
						log.Fatalln("create image (%v) failure, got err %v\n", image, err)
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

				if cfg.IsClean {
					// delete tmp file
					err := os.Remove(file.Name())
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		if len(vehicle.Images.Files) == 0 {
			return
		}

		pp.Println(vehicle)

		if !cfg.DryMode {
			if err := cfg.DB.Create(&vehicle).Error; err != nil {
				log.Fatalf("create vehicle (%v) failure, got err %v", vehicle, err)
				return
			}
		}

	})

	c.OnResponse(func(r *colly.Response) {
		if cfg.IsDebug {
			fmt.Println("OnResponse from", r.Ctx.Get("url"))
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		//if cfg.IsDebug {
		fmt.Println("Visiting", r.URL.String())
		//}
		r.Ctx.Put("url", r.URL.String())
	})

	// Start scraping on https://www.yamaha-occasion.com
	for _, u := range cfg.URLs {
		q.AddURL(u)
	}

	// Consume URLs
	q.Run(c)

	return nil
}
