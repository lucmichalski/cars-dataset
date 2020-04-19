package crawler

import (
	"encoding/json"
	// "encoding/xml"
	"fmt"
	"os"
	"regexp"
	// "net/url"
	"strings"

	sitemapz "github.com/oxffaa/gopher-parse-sitemap"
	"github.com/yterajima/go-sitemap"
	"github.com/k0kubun/pp"
	"github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	// "github.com/PuerkitoBio/goquery"
	
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
)

/*
	Refs:
	- cd plugins/classicdriver.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-classicdriver.com.so ; cd ../..
*/

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
		colly.URLFilters(
			regexp.MustCompile("https://www\\.classicdriver\\.com/en/car/(.*)"),
		),
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

		var name, make, model, year string
		e.ForEach(`div.panel-panel.panel-title-and-subtitle h1`, func(_ int, el *colly.HTMLElement) {
			name = strings.TrimSpace(el.Text)		
			if cfg.IsDebug {
				fmt.Println("name:", name)
			}
			nameParts := strings.Split(name, " ")			
			year = nameParts[0]
			make = nameParts[1]
			model = nameParts[2]
		})
		if cfg.IsDebug {
			fmt.Println("year:", year, "make:", make, "model:", model)
		}

		var carDataImage []string
		e.ForEach(`ul.slides img`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("src")
			if cfg.IsDebug {
				fmt.Println("carImage:", carImage)
			}
			carDataImage = append(carDataImage, carImage)
		})

		e.ForEach(`div.field-name-field-car-type div.field-item`, func(_ int, el *colly.HTMLElement) {
			carType := strings.TrimSpace(el.Text)
			vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Car Type", Value: carType})
		})		

		if make == "" && model == "" && year == "" {
			return
		}

		vehicle.Manufacturer = make
		vehicle.Year = year
		vehicle.Modl = model
		vehicle.Name = make + " " + model + " " + year
		vehicle.Source = "classicdriver.com"

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

				image := models.VehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum}

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

	log.Infoln("sitemapURL: ", cfg.URLs[0])

	smap, err := sitemap.Get(cfg.URLs[0], nil)
	if err != nil {
		fmt.Println(err)
	}

	// Print URL in sitemap.xml
	for _, URL := range smap.URL {
		fmt.Println(URL.Loc)
	}

	result := make([]string, 0, 0)
	err = sitemapz.ParseFromSite(cfg.URLs[0], func(e sitemapz.Entry) error {
	    result = append(result, e.GetLocation())
	    return nil
	})

	pp.Println(result)

	// Start scraping on https://www.classicdriver.com
	if cfg.IsSitemapIndex {
		log.Infoln("extractSitemapIndex...")
		pp.Println(cfg.URLs)
		sitemaps, err := prefetch.ExtractSitemapIndex(cfg.URLs[0])
		pp.Println(sitemaps)
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



