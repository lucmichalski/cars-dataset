package crawler

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	//"regexp"
	"github.com/k0kubun/pp"
	// "github.com/corpix/uarand"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"

	// "github.com/PuerkitoBio/goquery"
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
	Refs:
	- cd plugins/cars24.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-cars24.com.so ; cd ../..
*/

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
		//colly.URLFilters(
		//	regexp.MustCompile("https://www\\.cars24\\.com/buy-used(.*)"),
		//),
	)

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	// c.DisableCookies()

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error:", err, r.Request.URL, r.StatusCode)
		q.AddURL(r.Request.URL.String())
	})

	c.OnHTML(`body`, func(e *colly.HTMLElement) {

		// check in the databse if exists
		var vehicleExists models.Vehicle
		if !cfg.DryMode {
			if !cfg.DB.Where("url = ?", e.Request.Ctx.Get("url")).First(&vehicleExists).RecordNotFound() {
				fmt.Printf("skipping url=%s as already exists\n", e.Request.Ctx.Get("url"))
				return
			}
		}

		vehicle := &models.Vehicle{}
		vehicle.URL = e.Request.Ctx.Get("url")

		var make, model, version string
		e.ForEach(`ol.breadcrumb li:last-child`, func(_ int, el *colly.HTMLElement) {
			element := el.Text
			elements := strings.Split(element, " ")
			make = elements[0]
			model = strings.TrimSpace(strings.Replace(element, make, "", -1))
			if cfg.IsDebug {
				fmt.Println("make:", make)
				fmt.Println("model:", model)
			}
		})

		var name string
		e.ForEach(`h1.d-inline`, func(_ int, el *colly.HTMLElement) {
			name = strings.TrimSpace(el.Text)
			if cfg.IsDebug {
				fmt.Println("name:", name)
			}
		})

		version = strings.Replace(name, make, "", -1)
		version = strings.TrimSpace(strings.Replace(version, model, "", -1))

		var carDataImage []string
		e.ForEach(`div.slick-slide img`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("src")
			if cfg.IsDebug {
				fmt.Println("carImage:", carImage)
			}
			carDataImage = append(carDataImage, carImage)
		})

		var year string
		e.ForEach(`div[id=overview-container] li`, func(_ int, el *colly.HTMLElement) {
			var label, value string
			el.ForEach(`label`, func(_ int, eli *colly.HTMLElement) {
				label = strings.TrimSpace(eli.Text)
				if cfg.IsDebug {
					fmt.Println("label:", label)
				}
			})
			el.ForEach(`p`, func(_ int, eli *colly.HTMLElement) {
				value = strings.TrimSpace(eli.Text)
				if cfg.IsDebug {
					fmt.Println("value:", value)
				}
			})
			if label == "Year" {
				year = value
			} else {
				if label == "Fuel Type" || label == "Transmission" || label == "RTO" {
					vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: label, Value: value})
				}
			}
		})

		if cfg.IsDebug {
			fmt.Println("year:", year)
		}

		if make == "" && model == "" && year == "" {
			return
		}

		// overview-container

		// autosphere.fr legacy
		vehicle.Manufacturer = make
		vehicle.Engine = version
		vehicle.Year = year
		vehicle.Modl = model
		vehicle.Name = make + " " + model + " " + year
		vehicle.Source = "cars24.com"

		// Pictures
		for _, carImage := range carDataImage {
			if carImage == "" {
				continue
			}

			// comment temprorarly as we develop on local
			proxyURL := fmt.Sprintf("http://localhost:9003/crop?url=%s", carImage)
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

		log.Infoln("Add manufacturer: ", make, ", Model:", model, ", Year:", year)

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

	// Start scraping on https://www.cars24.com
	if cfg.IsSitemapIndex {
		log.Infoln("extractSitemapIndex...")
		sitemaps, err := prefetch.ExtractSitemapIndex(cfg.URLs[0])
		if err != nil {
			log.Fatal("ExtractSitemapIndex:", err)
			return err
		}

		utils.Shuffle(sitemaps)
		for _, sitemap := range sitemaps {
			log.Infoln("processing ", sitemap)
			if strings.Contains(sitemap, ".gz") {
				log.Infoln("extract sitemap gz compressed...")
				locs, err := prefetch.ExtractSitemapGZ(sitemap)
				if err != nil {
					log.Fatal("ExtractSitemapGZ", err)
					return err
				}
				utils.Shuffle(locs)
				for _, loc := range locs {
					q.AddURL(loc)
				}
			} else {
				q.AddURL(sitemap)
			}
		}
	} else {
		for _, u := range cfg.URLs {
			q.AddURL(u)
		}
	}

	// Consume URLs
	q.Run(c)

	return nil
}
