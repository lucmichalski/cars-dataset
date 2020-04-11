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

	pmodels "github.com/lucmichalski/cars-contrib/autotrader.com/models"
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

	c.DisableCookies()

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
		modele := e.ChildText("span[class=modele]")
		if cfg.IsDebug {
			fmt.Println("modele:", modele)
		}
		if modele == "" {
			return
		}
		version := e.ChildText("span[class=version]")
		if cfg.IsDebug {
			fmt.Println("version:", version)
		}

		var carInfo pmodels.VehicleGtm
		e.ForEach(`div[id=gtm_goal]`, func(_ int, el *colly.HTMLElement) {
			info := el.Attr("data-gtm-goal")
			infoParts := strings.Split(info, "--**--")
			if len(infoParts) > 0 {
				if infoParts[0] != "" {
					if err := json.Unmarshal([]byte(infoParts[0]), &carInfo); err != nil {
						log.Fatalln("unmarshal error, ", err)
					}
				}
				if cfg.IsDebug {
					pp.Println(carInfo)
				}
			}
		})

		if carInfo.ProductModele == "" {
			return
		}

		vehicle.Manufacturer = carInfo.ProductBrand
		vehicle.Engine = version
		vehicle.Year = carInfo.ProductYear
		vehicle.Modl = carInfo.ProductModele
		vehicle.Name = carInfo.ProductBrand + " " + carInfo.ProductModele + " " + carInfo.ProductYear
		vehicle.Source = "autotrader.com"

		vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Price", Value: carInfo.ProductPrice})
		vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Transmission", Value: carInfo.ProductTransmission})

		// Pictures
		var carImgLinks []string
		e.ForEach(`div[class=swiper-slide] > img`, func(_ int, el *colly.HTMLElement) {
			carPicSrc := el.Attr("src")
			carPicDataSrc := el.Attr("data-src")
			if cfg.IsDebug {
				if carPicSrc != "" {
					fmt.Println("carPicSrc:", carPicSrc)
				}
				if carPicDataSrc != "" {
					fmt.Println("carPicDataSrc:", carPicDataSrc)
				}
			}
			carPicDataSrc = strings.Replace(carPicDataSrc, "mini/", "", -1)
			carImgLinks = append(carImgLinks, carPicDataSrc)
			carPicSrc = strings.Replace(carPicSrc, "mini/", "", -1)
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

			proxyURL := fmt.Sprintf("http://darknet:9003/crop?url=%s", carImage)
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

		var manufacturer, color, model, gearbox, year, power, carType, certCritAir, c02, realPower, gas, doors, places string
		// e.ForEach(`div[class=swiper-slide] > img`, func(_ int, el *colly.HTMLElement) {
		e.DOM.Find("div.row-fluid.description_vehicule").Children().Each(func(idx int, sel *goquery.Selection) {
			texts := strings.Split(sel.Text(), ":")

			texts[0] = strings.TrimSpace(texts[0])
			texts[0] = strings.TrimLeftFunc(texts[0], func(c rune) bool {
				return c == '\r' || c == '\n' || c == '\t'
			})

			if len(texts) > 1 {
				texts[1] = strings.TrimLeftFunc(texts[1], func(c rune) bool {
					return c == '\r' || c == '\n' || c == '\t'
				})
			}

			// pp.Println("left info", texts)
			switch texts[0] {
			case "Marque":
				manufacturer = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Manufacturer", Value: manufacturer})
			case "Couleur":
				color = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Color", Value: color})
			case "Modèle":
				model = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Model", Value: model})
			case "Boîte de vitesse":
				gearbox = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "GearBox", Value: gearbox})
			case "Année":
				year = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Year", Value: year})
			case "Puissance Fiscale":
				power = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "HorsePower", Value: power})
			case "Type de véhicule":
				carType = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "CarType", Value: carType})
			case "Certificat CRIT'AIR":
				certCritAir = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "CRIT'AIR Certificat", Value: certCritAir})
			case "Co2":
				c02 = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Co2", Value: c02})
			case "Puissance Réelle":
				realPower = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "RealPower", Value: realPower})
			case "Carburant":
				gas = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "GasType", Value: gas})
			case "Portes":
				doors = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Doors", Value: doors})
			case "Places":
				places = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, models.VehicleProperty{Name: "Places", Value: places})
			}

		})

		if cfg.IsDebug {
			fmt.Println("manufacturer:", manufacturer)
			fmt.Println("color:", color)
			fmt.Println("model:", model)
			fmt.Println("gearbox:", gearbox)
			fmt.Println("year:", year)
			fmt.Println("power:", power)
			fmt.Println("carType:", carType)
			fmt.Println("certCritAir:", certCritAir)
			fmt.Println("c02:", c02)
			fmt.Println("realPower:", realPower)
			fmt.Println("gas:", gas)
			fmt.Println("doors:", doors)
			fmt.Println("places:", places)
		}

		if !cfg.DryMode {
			if err := cfg.DB.Create(&vehicle).Error; err != nil {
				log.Fatalf("create vehicle (%v) failure, got err %v", vehicle, err)
				return
			}
		}

		log.Infoln("Add manufacturer: ", carInfo.ProductBrand, ", Model:", carInfo.ProductModele, ", Year:", carInfo.ProductYear)

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