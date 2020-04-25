

package crawler

import (
	"encoding/json"
	"encoding/csv"
	"fmt"
	"strings"
	"os"
	"regexp"

	"github.com/k0kubun/pp"
	"github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/tsak/concurrent-csv-writer"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
	Refs:
	- cd plugins/cardealpage.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-cardealpage.com.so ; cd ../..
*/

func Extract(cfg *config.Config) error {


	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	if !cfg.DryMode {
		// Rotate two socks5 proxies
		rp, err := proxy.RoundRobinProxySwitcher("http://localhost:8118")
		if err != nil {
			log.Fatal(err)
		}
		c.SetProxyFunc(rp)
	}

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	// read cache sitemap
	utils.EnsureDir("./shared/queue/")

	if _, err := os.Stat("shared/queue/cardealpage.com_sitemap.txt"); !os.IsNotExist(err) {
	    file, err := os.Open("shared/queue/cardealpage.com_sitemap.txt")
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

		utils.Shuffle(data)
		for _, loc := range data {
			q.AddURL(loc[0])
		}

	}


	// regex rules on vehicles url
	vehicleURLRegexp, err := regexp.Compile(`https://www\.cardealpage\.com/([_0-9A-Za-z-]+)/([_%0-9A-Za-z-]+)/([0-9]+)/`)
	if err != nil {
		log.Warnln(err)
		return err
	}

	// save discovered links
	csvSitemap, err := ccsv.NewCsvWriter("shared/queue/cardealpage.com_sitemap.txt")
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
		if vehicleURLRegexp.MatchString(link) {
			// Print link
			fmt.Printf("Link found: %s\n", e.Request.AbsoluteURL(link))
			csvSitemap.Write([]string{e.Request.AbsoluteURL(link)})
			csvSitemap.Flush()
			q.AddURL(e.Request.AbsoluteURL(link))
		}
		// q.AddURL(e.Request.AbsoluteURL(link))
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

		if !vehicleURLRegexp.MatchString(e.Request.Ctx.Get("url")) {
			fmt.Println("vehicleURLRegexp failed for", e.Request.Ctx.Get("url"))
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
		vehicle.Source = "cardealpage.com"
		vehicle.Class = "car"

		carInfo := vehicleURLRegexp.FindAllStringSubmatch(e.Request.Ctx.Get("url"), -1)
		if len(carInfo[0]) < 2 {
			pp.Println(carInfo[0])
			fmt.Println("carInfo.Lenght:", len(carInfo[0]))
			return
		}

		vehicle.Manufacturer = carInfo[0][1]
		vehicle.Modl = strings.Replace(carInfo[0][2], "%20", " ", -1)

		e.ForEach(`div.btn_next`, func(_ int, el *colly.HTMLElement) {
			onclick := el.Attr("onclick")
			onclick = strings.Replace(onclick, "loadContent('', '", "", -1)
			onclick = strings.Replace(onclick, "', '');", "", -1)
			absLink := e.Request.AbsoluteURL(onclick)
			if cfg.IsDebug {
				fmt.Println("absLink:", absLink)
			}
			q.AddURL(absLink)
		})

		e.ForEach(`table[id=specifications] tr`, func(_ int, el *colly.HTMLElement) {
			var key, value string
			el.ForEach(`td.td1`, func(_ int, eli *colly.HTMLElement) {
				key = strings.TrimSpace(eli.Text)
			})

			el.ForEach(`td.td2`, func(_ int, eli *colly.HTMLElement) {
				value = strings.TrimSpace(eli.Text)
				if key == "Reg.Year / Month" {
					valueParts := strings.Split(value, "/")
					value = valueParts[0]
				}
				value = strings.TrimLeftFunc(value, func(c rune) bool {
					return c == '\r' || c == '\n' || c == '\t'
				})
				value = strings.TrimRightFunc(value, func(c rune) bool {
					return c == '\r' || c == '\n' || c == '\t'
				})				
			})
			switch key {
            case "Steering":
                    vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Steering", Value: value})
			case "Fuel":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "FuelType", Value: value})
			case "Transmission":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Transmission", Value: value})
			case "Drive":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Drive", Value: value})
			case "Doors":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Doors", Value: value})
			case "No. of Seats":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "No. of Seats", Value: value})
			case "Colour":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Color", Value: value})
            case "Reg.Year":
				fallthrough
			case "Reg.Year / Month":
				vehicle.Year = value
            case "Engine":
                    vehicle.Engine = value
			}

			if cfg.IsDebug {
				fmt.Println("key:", key, "value:", value)
			}

		})

		vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year

		var carDataImage []string
		// e.ForEach(`div.smallPhoto a:nth-child(-n+10)`, func(_ int, el *colly.HTMLElement) {
		e.ForEach(`div.smallPhoto a`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("href")
			if cfg.IsDebug {
				fmt.Println("carImage:", carImage)
			}
			carDataImage = append(carDataImage, carImage)
		})

		pp.Println(carDataImage)
		pp.Println(vehicle)
		///os.Exit(1)

		if vehicle.Manufacturer == "" && vehicle.Modl == "" && vehicle.Year == "" {
			fmt.Println("incomplete")
			return
		}

		// Pictures
		for _, carImage := range carDataImage {
			if carImage == "" {
				continue
			}

			// comment temprorarly as we develop on local
			proxyURL := fmt.Sprintf("http://localhost:9005/crop?url=%s", carImage)
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

		// log.Infoln("Add manufacturer: ", make, ", Model:", model, ", Year:", year)

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
