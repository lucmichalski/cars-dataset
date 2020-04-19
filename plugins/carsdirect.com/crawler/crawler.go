package crawler

import (
	"encoding/json"
	"encoding/csv"
	"fmt"
	"strings"
	"os"
	"net/url"
	// "regexp"
	"sort"

	"github.com/k0kubun/pp"
	"github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/tsak/concurrent-csv-writer"
	"github.com/astaxie/flatmap"
	
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
)

/*
	Refs:
	- rsync -av -v --ignore-existing —-progress -e "ssh -i ~/Downloads/ounsi.pem" /Volumes/HardDrive/go/src/github.com/lucmichalski/cars-dataset/public ubuntu@35.179.44.166:/home/ubuntu/cars-dataset/
	- cd plugins/carsdirect.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-carsdirect.com.so ; cd ../..
	- good practices
		- https://intoli.com/blog/making-chrome-headless-undetectable/
		- https://github.com/ridershow/scraping_toolbox
		- https://github.com/Zenika/alpine-chrome/tree/master/with-webgl/swiftshader
		- https://github.com/microsoft/playwright
		- https://datadome.co/bot-detection/will-playwright-replace-puppeteer-for-bad-bot-play-acting/
		- https://datadome.co/pricing/
		- 
*/

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
		//colly.URLFilters(
		//	regexp.MustCompile("https://www\\.cars\\.com/vehicledetail/(.*)"),
		//),
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

	if _, err := os.Stat("shared/queue/carsdirect.com_sitemap.txt"); !os.IsNotExist(err) {
	    file, err := os.Open("shared/queue/carsdirect.com_sitemap.txt")
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
	csvSitemap, err := ccsv.NewCsvWriter("shared/queue/carsdirect.com_sitemap.txt")
	if err != nil {
		panic("Could not open `csvSitemap.csv` for writing")
	}

	// Flush pending writes and close file upon exit of Sitemap()
	defer csvSitemap.Close()

	// c.DisableCookies()

	// `https://www.carsdirect.com/(\d{4})/([^\/]+)/([^\/]+)/pictures`

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		if strings.Contains(e.Text, "/pictures") {
			q.AddURL(e.Text)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("error:", err, r.Request.URL, r.StatusCode)
		q.AddURL(r.Request.URL.String())
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {

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
		vehicle.Source = "carsdirect.com"
		vehicle.Class = "car"

		var make, model, year string
		var carInfo map[string]interface{}
		e.ForEach(`script[type="application/ld+json"]`, func(_ int, el *colly.HTMLElement) {
			jsonLdStr := strings.TrimSpace(el.Text)	
			if cfg.IsDebug {
				fmt.Println("jsonLdStr:", jsonLdStr)
			}
			if err := json.Unmarshal([]byte(jsonLdStr), &carInfo); err != nil {
				log.Fatalln("unmarshal error, ", err)
			}

			fm, err := flatmap.Flatten(carInfo)
			if err != nil {
				log.Fatal(err)
			}
			var ks []string
			for k :=range fm {
				ks = append(ks,k)		
			}
			sort.Strings(ks)

			if cfg.IsDebug {			
				for _, k :=range ks {
					fmt.Println(k,":",fm[k])
				}
			}

			if val, ok := fm["mainEntity.brand.name"]; ok {
				vehicle.Manufacturer = val
			}

			if val, ok := fm["mainEntity.vehicleModelDate"]; ok {
				vehicle.Year = val
			}

			if val, ok := fm["mainEntity.model"]; ok {
				vehicle.Modl = val
			}

			vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year

			if val, ok := fm["mainEntity.offers.highPrice"]; ok {
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "HighPrice", Value: strings.Replace(val, ".000000", "", -1)})
			}

			if val, ok := fm["mainEntity.offers.lowPrice"]; ok {
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "LowPrice", Value: strings.Replace(val, ".000000", "", -1)})
			}

			var fuelEfficiencyUnit string
			if val, ok := fm["mainEntity.fuelEfficiency.unitText"]; ok {
				fuelEfficiencyUnit = val
			}

			if val, ok := fm["mainEntity.fuelEfficiency.maxValue"]; ok {
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "FuelEfficiencyMax", Value: strings.Replace(val, ".000000", "", -1) + " " + fuelEfficiencyUnit})
			}

			if val, ok := fm["mainEntity.fuelEfficiency.minValue"]; ok {
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "FuelEfficiencyMin", Value: strings.Replace(val, ".000000", "", -1) + " " + fuelEfficiencyUnit})
			}

		})

		var carDataImage []string
		e.ForEach(`.photoCell img`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("src")
			if cfg.IsDebug {
				fmt.Println("carImage:", carImage)
			}
			// https://cdcssl.ibsrv.net/autodata/images/?IMG=USC50ALC061A01308.JPG&width=1144
			if strings.HasPrefix(carImage, "//") {
				carImage = "https:" + carImage
				carImage = strings.Replace(carImage, "&width=572", "", -1)
				carImage = strings.Replace(carImage, "?IMG=", "?width=1144&IMG=", -1)
			}
			carImage = url.QueryEscape(carImage)
			carDataImage = append(carDataImage, carImage)
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
			proxyURL := fmt.Sprintf("http://35.179.44.166:9004/crop?url=%s", carImage)
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

	// Start scraping on https://www.classicdriver.com
	if cfg.IsSitemapIndex {
		log.Infoln("extractSitemapIndex...")
		sitemaps, err := prefetch.ExtractSitemapIndex(cfg.URLs[0])
		if err != nil {
			log.Fatal("ExtractSitemapIndex:", err)
			return err
		}

		var links []string
		utils.Shuffle(sitemaps)
		for _, sitemap := range sitemaps {
			log.Infoln("processing ", sitemap)
			if strings.HasSuffix(sitemap, ".gz") {
				log.Infoln("extract sitemap gz compressed...")
				locs, err := prefetch.ExtractSitemapGZ(sitemap)
				if err != nil {
					log.Fatal("ExtractSitemapGZ: ", err, "sitemap: ",sitemap)
					return err
				}
				utils.Shuffle(locs)
				for _, loc := range locs {
					links = append(links, loc)
				}
			} else {
				locs, err := prefetch.ExtractSitemap(sitemap)
				if err != nil {
					log.Fatal("ExtractSitemap", err)
					return err
				}
				utils.Shuffle(locs)
				for _, loc := range locs {
					links = append(links, loc)
				}				
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
