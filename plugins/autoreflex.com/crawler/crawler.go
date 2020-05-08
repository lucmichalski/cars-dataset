package crawler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/corpix/uarand"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/gocolly/colly/v2/queue"
	"github.com/k0kubun/pp"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
)

/*
	Refs:
	- rsync -av --ignore-existing —-progress -e "ssh -i ~/Downloads/ounsi.pem" /Volumes/HardDrive/go/src/github.com/lucmichalski/cars-dataset/public ubuntu@51.91.21.67:/home/ubuntu/cars-dataset/
	- scp -i ~/Downloads/ounsi.pem /Volumes/HardDrive/go/src/github.com/lucmichalski/cars-dataset/public/* ubuntu@51.91.21.67:/home/ubuntu/cars-dataset/public/
	- cd plugins/autoreflex.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-autoreflex.com.so ; cd ../..
*/

func Extract(cfg *config.Config) error {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cfg.CacheDir),
	)

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("http://localhost:8119")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	reDate := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

	// create a request queue with 1 consumer thread until we solve the multi-threadin of the darknet model
	q, _ := queue.New(
		cfg.ConsumerThreads,
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		},
	)

	// read cache sitemap
	utils.EnsureDir("./shared/queue/")

	if _, err := os.Stat("shared/queue/autoreflex.com_sitemap.txt"); !os.IsNotExist(err) {
		file, err := os.Open("shared/queue/autoreflex.com_sitemap.txt")
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
		//if !strings.Contains(e.Request.Ctx.Get("url"), "listing") {
		//	return
		//}

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
		vehicle.Source = "autoreflex.com"
		vehicle.Class = "car"

		var make, model, year string
		e.ForEach(`div.header h1 a:first-child`, func(_ int, el *colly.HTMLElement) {
			fmt.Println("make", el.Text)
			make = el.Text
		})

		e.ForEach(`div.header h1 a:nth-of-type(n+1)`, func(_ int, el *colly.HTMLElement) {
			fmt.Println("model", el.Text)
			model = el.Text
		})

		e.ForEach(`div.specs`, func(_ int, el *colly.HTMLElement) {
			yearStr := reDate.FindString(el.Text)
			fmt.Println("yearStr", yearStr)
		})

		e.ForEach(`div.specs ul li:nth-of-type(n+2)`, func(_ int, el *colly.HTMLElement) {
			if year == "" {
				fmt.Println("year", el.Text)
				dateParts := strings.Split(el.Text, "/")
				if len(dateParts) > 2 {
					year = dateParts[2]
				}
			}
		})

		fmt.Println("make", make, "model", model, "year", year)
		vehicle.Manufacturer = make
		vehicle.Modl = model
		vehicle.Year = year

		vehicle.Name = vehicle.Manufacturer + " " + vehicle.Modl + " " + vehicle.Year

		var carDataImage []string
		e.ForEach(`.photo.photogrand img`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("src")
			if carImage != "" {
				carDataImage = append(carDataImage, carImage)
			}
		})

		e.ForEach(`img.fiche-vignette`, func(_ int, el *colly.HTMLElement) {
			carImage := el.Attr("data-grand")
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

			proxyURL := fmt.Sprintf("http://51.91.21.67:9005/labelme?url=%s", carImage)
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

				file, checksum, err := utils.DecodeToFile(carImage, detection.ImageData)
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
				image := models.VehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum, Source: carImage, BBox: bbox}

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
			fmt.Println(r.Request.URL, r.StatusCode)
			// fmt.Println(string(r.Body))
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

		// var links []string
		utils.Shuffle(sitemaps)
		for _, sitemap := range sitemaps {
			log.Infoln("processing ", sitemap)
			if strings.HasSuffix(sitemap, ".gz") {
				log.Infoln("extract sitemap gz compressed...")
				locs, err := prefetch.ExtractSitemapGZ(sitemap)
				if err != nil {
					log.Fatal("ExtractSitemapGZ: ", err, "sitemap: ", sitemap)
					return err
				}
				utils.Shuffle(locs)
				for _, loc := range locs {
					if strings.Contains(loc, "listing") {
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
					if strings.Contains(loc, "listing") {
						q.AddURL(loc)
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
