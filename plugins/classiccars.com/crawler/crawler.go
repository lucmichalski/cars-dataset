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
	"github.com/PuerkitoBio/goquery"
	
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
)

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
		/*
			colly.URLFilters(
				regexp.MustCompile("https://autosphere\\.fr/(|e.+)$"),
				regexp.MustCompile("https://www.autosphere\\.fr/h.+"),
			),
		*/
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
		if !cfg.DB.Where("url = ?", e.Request.Ctx.Get("url")).First(&vehicleExists).RecordNotFound() {
			fmt.Printf("skipping url=%s as already exists\n", e.Request.Ctx.Get("url"))
			return
		}

		vehicle := &models.Vehicle{}
		vehicle.URL = e.Request.Ctx.Get("url")

		/*
		<div id="listing-content" class="fx-item fx-va-top fi-3pan-2nd-col"
			 data-favorite="false"
			 data-listing="1310951"
			 data-listing-url="/listings/view/1310951/1985-land-rover-defender-for-sale-in-oceanside-california-92057"
			 data-listing-thumbnail=""
			 data-listing-year="1985"
			 data-listing-make="Land Rover"
			 data-listing-model="Defender"
			 data-listing-formatted-price="$25,000">
		*/

		var gid, year, make, model, formattedPrice string
		e.ForEach(`div[id=listing-content]`, func(_ int, el *colly.HTMLElement) {
			gid  = el.Attr("data-listing")
			year = el.Attr("data-listing-year")
			make = el.Attr("data-listing-make")
			model = el.Attr("data-listing-model")
			formattedPrice = el.Attr("data-listing-formatted-price")
		})
		if cfg.IsDebug {
			fmt.Println("gid:", gid, "year:", year, "make:", make, "model:", model, "formattedPrice:", formattedPrice)
		}

		vehicle.Manufacturer = make
		vehicle.Gid = gid
		vehicle.Year = year
		vehicle.Modl = model
		vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year
		vehicle.Source = "classiccars.com"
		vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Price", Value: formattedPrice})

		// modele := e.ChildText("span[class=modele]")
		if make == "" && year == "" && model == "" {
			return
		}

		e.DOM.Find("ul.panel-mod.pm-details-list li.border-btm:nth-child(n+3)").Each(func(idx int, sel *goquery.Selection) {
			texts := strings.Split(sel.Text(), ":")

			texts[0] = strings.TrimSpace(texts[0])
			texts[0] = strings.TrimLeftFunc(texts[0], func(c rune) bool {
				return c == '\r' || c == '\n' || c == '\t'
			})

			if len(texts) > 1 {
				texts[1] = strings.TrimLeftFunc(texts[1], func(c rune) bool {
					return c == '\r' || c == '\n' || c == '\t'
				})
				texts[1] = strings.TrimRightFunc(texts[1], func(c rune) bool {
					return c == '\r' || c == '\n' || c == '\t'
				})
			}

			pp.Println("texts", texts)
			switch texts[0] {
			case "Exterior Color":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "ExteriorColor", Value: strings.TrimSpace(texts[1])})
			case "Interior Color":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "InteriorColor", Value: strings.TrimSpace(texts[1])})
			case "Transmission":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Transmission", Value: strings.TrimSpace(texts[1])})
			case "Engine Size":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "EngineSize", Value: strings.TrimSpace(texts[1])})
				vehicle.Engine = strings.TrimSpace(texts[1])
			case "Sunroof":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Sunroof", Value: strings.TrimSpace(texts[1])})
			case "Seat Material":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "SeatMaterial", Value: strings.TrimSpace(texts[1])})
			case "Air Conditioning":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "AirConditioning", Value: strings.TrimSpace(texts[1])})
			case "Tinted Windows":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "TintedWindows", Value: strings.TrimSpace(texts[1])})
			case "Bucket Seats":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "BucketSeats", Value: strings.TrimSpace(texts[1])})
			case "Power Brakes":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "PowerBrakes", Value: strings.TrimSpace(texts[1])})
			case "Drive Train":
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "DriveTrain", Value: strings.TrimSpace(texts[1])})
			}
		})

		// Pictures
		var carImgLinks []string
		e.ForEach(`div.swiper-slide`, func(_ int, el *colly.HTMLElement) {
			carPicSrc := el.Attr("data-jumbo")
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

		if len(carImgLinks) == 0 {
			return
		}

		for _, carImage := range carImgLinks {
			if carImage == "" {
				continue
			}

			proxyURL := fmt.Sprintf("http://localhost:9006/crop?url=%s", carImage)
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
						if strings.Contains(loc, "listings/view") {
							q.AddURL(loc)
						}
					}
				} else {
					locs, err := prefetch.ExtractSitemap(sitemap)
					if err != nil {
						log.Fatal("ExtractSitemap", err)
						return err
					}
					utils.Shuffle(locs)
					for _, loc := range locs {
						if strings.Contains(loc, "listings/view") {
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
