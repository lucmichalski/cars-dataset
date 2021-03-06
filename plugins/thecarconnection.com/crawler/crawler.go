package crawler

import (
	"encoding/json"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	// "github.com/k0kubun/pp"
	// "github.com/corpix/uarand"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
        // "github.com/gocolly/colly/v2/proxy"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"

	// "github.com/PuerkitoBio/goquery"
	pmodels "github.com/lucmichalski/cars-contrib/thecarconnection.com/models"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
	Refs:
	- cd plugins/thecarconnection.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-thecarconnection.com.so ; cd ../..
	- https://medium.com/syncedreview/vroom-vroom-new-dataset-rolls-out-64-000-pictures-of-cars-b99ac99843ea
	- https://github.com/nicolas-gervais/predicting-car-price-from-scraped-data/tree/master/picture-scraper
		- cols :'Make', 'Model', 'Year', 'MSRP', 'Front Wheel Size (in)', 'SAE Net Horsepower @ RPM', 'Displacement', 'Engine Type', 'Width, Max w/o mirrors (in)', 'Height, Overall (in)', 'Length, Overall (in)', 'Gas Mileage', 'Drivetrain', 'Passenger Capacity', 'Passenger Doors', 'Body Style'
		- https://github.com/nicolas-gervais/predicting-car-price-from-scraped-data/blob/master/picture-scraper/scrape.py#L33
*/

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
		colly.URLFilters(
			regexp.MustCompile("https://www\\.thecarconnection\\.com/overview/(.*)"),
			regexp.MustCompile("https://www\\.thecarconnection\\.com/photos/(.*)"),
		),
	)
	/*
        rp, err := proxy.RoundRobinProxySwitcher("http://localhost:8119")
        if err != nil {
                log.Fatal(err)
        }
        c.SetProxyFunc(rp)
	*/
	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	if _, err := os.Stat("shared/queue/thecarconnection.com_sitemap.txt"); !os.IsNotExist(err) {
		file, err := os.Open("shared/queue/thecarconnection.com_sitemap.txt")
		if err != nil {
			return err
		}

		reader := csv.NewReader(file)
		reader.Comma = ';'
		reader.LazyQuotes = true
		data, err := reader.ReadAll()
		if err != nil {
			return err
		}

		utils.Shuffle(data)
		for _, loc := range data {
			q.AddURL(loc[0])
		}
	}

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

		make := strings.TrimSpace(e.ChildText("div[id=breadcrumbs] a[id=a_bc_1]"))
		if cfg.IsDebug {
			fmt.Println("model:", make)
		}

		model := strings.TrimSpace(e.ChildText("div[id=breadcrumbs] a[id=a_bc_2]"))
		if cfg.IsDebug {
			fmt.Println("model:", model)
		}

		year := strings.TrimSpace(e.ChildText("div[id=breadcrumbs] a[id=a_bc_3]"))
		if cfg.IsDebug {
			fmt.Println("year:", year)
		}

		price := e.ChildText("li.style-select.style.selected span[class=price]")
		if cfg.IsDebug {
			fmt.Println("price:", price)
		}

		engine := e.ChildText("li.style-select.style.selected span[class=name]")
		if cfg.IsDebug {
			fmt.Println("engine:", engine)
		}

		vehicle.Engine = engine
		vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Price", Value: price})

		// "https://www.thecarconnection.com/specifications/jaguar_f-type_2021",
		/*
			e.ForEach(`div.category-details div.specs-set-item`, func(_ int, el *colly.HTMLElement) {
				var key, value string
				el.ForEach(`span.key`, func(_ int, eli *colly.HTMLElement) {
					key = strings.TrimSpace(eli.Text)
				})
				el.ForEach(`span.value`, func(_ int, eli *colly.HTMLElement) {
					value = strings.TrimSpace(eli.Text)
				})
				if cfg.IsDebug {
					fmt.Println("key", key, " <========> value", value)
				}
				if key == "Engine" {
					vehicle.Engine = value
				}
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: key, Value: value})
			})
		*/
		// os.Exit(1)

		var carDataImage []*pmodels.Car
		e.ForEach(`div.gallery`, func(_ int, el *colly.HTMLElement) {
			carDataImageRaw := el.Attr("data-model")
			if cfg.IsDebug {
				// fmt.Println("carDataImageRaw:", carDataImageRaw)
			}
			if err := json.Unmarshal([]byte(carDataImageRaw), &carDataImage); err != nil {
				log.Fatalln("unmarshal error, ", err)
			}
		})

		if make == "" && model == "" && year == "" {
			return
		}

		// autosphere.fr legacy
		vehicle.Manufacturer = make
		// vehicle.Engine = version
		vehicle.Year = year
		vehicle.Modl = model
		vehicle.Name = make + " " + model + " " + year
		vehicle.Source = "thecarconnection.com"

		// get additional data

		// vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Price", Value: carInfo.ProductPrice})
		// vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Transmission", Value: carInfo.ProductTransmission})

		// Pictures

		for _, carImage := range carDataImage {
			if carImage.Images.Large.URL == "" {
				continue
			}

			proxyURL := fmt.Sprintf("http://51.91.21.67:9004/labelme?url=%s", strings.Replace( carImage.Images.Large.URL, " ", "%20", -1))
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

				file, checksum, err := utils.DecodeToFile(carImage.Images.Large.URL, detection.ImageData)
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
				image := models.VehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum, Source: carImage.Images.Large.URL, BBox: bbox}

				log.Println("----> Scanning file: ", file.Name())
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

	// Start scraping on https://www.thecarconnection.com
	if cfg.IsSitemapIndex {
		log.Infoln("extractSitemapIndex...")
		sitemaps, err := prefetch.ExtractSitemapIndex(cfg.URLs[0])
		if err != nil {
			log.Warnln("ExtractSitemapIndex:", err)
			// return err
		}

		utils.Shuffle(sitemaps)
		for _, sitemap := range sitemaps {
			log.Infoln("processing ", sitemap)
			if strings.Contains(sitemap, ".gz") {
				log.Infoln("extract sitemap gz compressed...")
				locs, err := prefetch.ExtractSitemapGZ(sitemap)
				if err != nil {
					log.Warnln("ExtractSitemapGZ", err)
					continue
					// return err
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
