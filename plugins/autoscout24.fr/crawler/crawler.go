package crawler

import (
	"encoding/json"
	"encoding/csv"
	"fmt"
	"strings"
	"os"

	"github.com/k0kubun/pp"
	"github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/tsak/concurrent-csv-writer"

	"github.com/lucmichalski/cars-dataset/pkg/pluck"		
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
	Refs:
	- cd plugins/autoscout24.fr && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-autoscout24.fr.so ; cd ../..
*/

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains(cfg.AllowedDomains...),
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

	// read cache sitemap
	utils.EnsureDir("./shared/queue/")

	if _, err := os.Stat("shared/queue/autoscout24.fr_sitemap.txt"); !os.IsNotExist(err) {
	    file, err := os.Open("shared/queue/autoscout24.fr_sitemap.txt")
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

		for _, loc := range data {
			q.AddURL(loc[0])
		}
	}

	// save discovered links
	csvSitemap, err := ccsv.NewCsvWriter("shared/queue/autoscout24.fr_sitemap.txt")
	if err != nil {
		panic("Could not open `csvSitemap.csv` for writing")
	}

	// Flush pending writes and close file upon exit of Sitemap()
	defer csvSitemap.Close()

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error:", err, r.Request.URL, r.StatusCode)
		q.AddURL(r.Request.URL.String())
	})

	c.OnHTML(`a[href]`, func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "offres/") {
			fmt.Printf("Link found: %s\n", e.Request.AbsoluteURL(link))
			csvSitemap.Write([]string{e.Request.AbsoluteURL(link)})
			csvSitemap.Flush()
		}
		q.AddURL(e.Request.AbsoluteURL(link))
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

		// filter offers
		if !strings.Contains(e.Request.Ctx.Get("url"), "offres/") {
			return
		}

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
		vehicle.Source = "autoscout24.fr"
		// vehicle.Class = "car"

		e.ForEach(`as24-tracking`, func(_ int, el *colly.HTMLElement) {
			vehicleType := el.Attr("category")
			if vehicleType == "moto" {
				vehicle.Class = "motorcycle"
			} else {
				vehicle.Class = "car"
			}
		})

		p, err := pluck.New()
		if err != nil {
			return
		}
		p.Add(pluck.Config{
			Activators:  []string{"Marque</dt>\n<dd>"}, // must be found in order, before capturing commences
			Permanent:   1,      // number of activators that stay permanently (counted from left to right)
			Deactivator: "</dd>",   // restarts capturing
			Limit:       1,      // specifies the number of times capturing can occur
			Name: "make",   // the key in the returned map, after completion
		})
		p.Add(pluck.Config{
			Activators:  []string{"Modèle</dt>\n<dd>"}, // must be found in order, before capturing commences
			Permanent:   1,      // number of activators that stay permanently (counted from left to right)
			Deactivator: "</dd>",   // restarts capturing
			Limit:       1,      // specifies the number of times capturing can occur
			Name: "model",   // the key in the returned map, after completion
		})
		p.Add(pluck.Config{
			Activators:  []string{"Année</dt>\n<dd>"}, // must be found in order, before capturing commences
			Permanent:   1,      // number of activators that stay permanently (counted from left to right)
			Deactivator: "</dd>",   // restarts capturing
			Limit:       1,      // specifies the number of times capturing can occur
			Name: "year",   // the key in the returned map, after completion
		})

		p.Add(pluck.Config{
			Activators:  []string{"Couleur extérieure</dt>\n<dd>"}, // must be found in order, before capturing commences
			Permanent:   1,      // number of activators that stay permanently (counted from left to right)
			Deactivator: "</dd>",   // restarts capturing
			Limit:       1,      // specifies the number of times capturing can occur
			Name: "color",   // the key in the returned map, after completion
		})
		p.Add(pluck.Config{
			Activators:  []string{"Portes</dt>\n<dd>"}, // must be found in order, before capturing commences
			Permanent:   1,      // number of activators that stay permanently (counted from left to right)
			Deactivator: "</dd>",   // restarts capturing
			Limit:       1,      // specifies the number of times capturing can occur
			Name: "doors",   // the key in the returned map, after completion
		})
		p.Add(pluck.Config{
			Activators:  []string{"Sièges</dt>\n<dd>"}, // must be found in order, before capturing commences
			Permanent:   1,      // number of activators that stay permanently (counted from left to right)
			Deactivator: "</dd>",   // restarts capturing
			Limit:       1,      // specifies the number of times capturing can occur
			Name: "seats",   // the key in the returned map, after completion
		})

		p.PluckURL(e.Request.Ctx.Get("url"))
		rev := p.Result()

		switch v := rev["make"].(type) {
		case string:
			vehicle.Manufacturer = utils.RemoveAllTags(v)
		}

		switch v := rev["model"].(type) {
		case string:
			vehicle.Modl = utils.RemoveAllTags(v)
		}

		switch v := rev["year"].(type) {
		case string:
			vehicle.Year = utils.RemoveAllTags(v)
		}

                if vehicle.Manufacturer == "" && vehicle.Modl == "" && vehicle.Year == "" {
                        return
                }

		switch v := rev["color"].(type) {
		case string:
			vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Color", Value: utils.RemoveAllTags(v)})
		}

		switch v := rev["doors"].(type) {
		case string:
			vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Doors", Value: v})
		}

		switch v := rev["seats"].(type) {
		case string:
			vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Seats", Value: v})
		}

		vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year

		var carDataImage []string
		e.ForEach(`.gallery-picture__image`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("src")
			if carImage == "" {
				carImage = el.Attr("data-src")
			}
			if cfg.IsDebug {
				fmt.Println("carImage:", carImage)
			}
			if carImage != "" {
				carDataImage = append(carDataImage, carImage)
			}
		})

		pp.Println(carDataImage)
		pp.Println(vehicle)

		if vehicle.Manufacturer == "" && vehicle.Modl == "" && vehicle.Year == "" {
			return
		}

		// Pictures
		for _, carImage := range carDataImage {
			if carImage == "" {
				continue
			}

			// comment temprorarly as we develop on local
			proxyURL := fmt.Sprintf("http://localhost:9004/crop?url=%s", carImage)
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
						log.Warnln("create image (%v) failure, got err %v\n", image, err)
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




	for _, u := range cfg.URLs {
		q.AddURL(u)
	}

	// Consume URLs
	q.Run(c)

	return nil
}
