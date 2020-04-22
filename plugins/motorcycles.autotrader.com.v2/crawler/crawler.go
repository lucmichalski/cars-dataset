package crawler

import (
	"encoding/json"
	"fmt"
	"strings"
	"os"

	"github.com/k0kubun/pp"
	// "github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		1,
		//cfg.ConsumerThreads,
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
		if !cfg.DB.Where("url = ?", e.Request.Ctx.Get("url")).First(&vehicleExists).RecordNotFound() {
			fmt.Printf("skipping url=%s as already exists\n", e.Request.Ctx.Get("url"))
			return
		}

		vehicle := &models.Vehicle{}
		vehicle.URL = e.Request.Ctx.Get("url")

		vehicle.Source = "motorcycles.autotrader.com"
		vehicle.Class = "motorcycle"

		make := e.ChildText(`ol.breadcrumbs li:first-child`)
		model := e.ChildText(`ol.breadcrumbs li:nth-of-type(n+2)`)
		year := e.ChildText(`ol.breadcrumbs ol.breadcrumbs li:nth-of-type(n+3)`)

		vehicle.Manufacturer = make
		vehicle.Year = year
		vehicle.Modl = model
		vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year

		// modele := e.ChildText("span[class=modele]")
		if make == "" && year == "" && model == "" {
			return
		}

		// Pictures
		var carImgLinks []string
		e.ForEach(`.vdp-gallery-secondary-slides img`, func(_ int, el *colly.HTMLElement) {
			carPicSrc := el.Attr("src")
			if cfg.IsDebug {
				if carPicSrc != "" {
					fmt.Println("carPicSrc:", carPicSrc)
				}
			}
			carPicSrc = strings.Replace(carPicSrc, "w=143", "w=735", -1)
			carPicSrc = strings.Replace(carPicSrc, "h=107", "h=551", -1)
			carImgLinks = append(carImgLinks, carPicSrc)
		})

		carImgLinks = utils.RemoveDuplicates(carImgLinks)
		if cfg.IsDebug {
			pp.Println(carImgLinks)
		}

		if len(carImgLinks) == 0 {
			return
		}

		for _, carImage := range carImgLinks {
			if carImage == "" {
				continue
			}

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

		// log.Infoln("Add manufacturer: ", carInfo.ProductBrand, ", Model:", carInfo.ProductModele, ", Year:", carInfo.ProductYear)

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

	// Start scraping on https://www.classiccars.com
	if cfg.IsSitemapIndex {
		log.Infoln("extractSitemapIndex...")
		for _, rootUrl := range cfg.URLs {
			sitemaps, err := prefetch.ExtractSitemapIndex(rootUrl)
			if err != nil {
				log.Fatal("ExtractSitemapIndex:", err)
				return err
			}

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
						if strings.HasPrefix(loc, "https://motorcycles.autotrader.com/motorcycles") {
							q.AddURL(loc)
						}
					}
				} else {
					if !strings.Contains(sitemap, "sitemap_vehicles") {
						continue
					}
					locs, err := prefetch.ExtractSitemap(sitemap)
					if err != nil {
						log.Fatal("ExtractSitemap", err)
						return err
					}
					utils.Shuffle(locs)
					for _, loc := range locs {
						if strings.HasPrefix(loc, "https://motorcycles.autotrader.com/motorcycles") {
							q.AddURL(loc)
						}
					}				
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
