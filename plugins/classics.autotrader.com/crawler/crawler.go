package crawler

import (
	"encoding/json"
	"fmt"
	"strings"
	"os"
	// "time"

	// "github.com/qor/oss/filesystem"
	// "github.com/qor/oss/s3"
	"github.com/k0kubun/pp"
	// "github.com/corpix/uarand"
	"github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/lucmichalski/cars-dataset/pkg/selenium"
	"github.com/lucmichalski/cars-dataset/pkg/selenium/chrome"
	slog "github.com/lucmichalski/cars-dataset/pkg/selenium/log"
	"github.com/nozzle/throttler"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
)

/*
	- cd plugins/classics.autotrader.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-classics.autotrader.com.so ; cd ../..
	- rsync -av —-progress -e "ssh -i ~/Downloads/ounsi.pem" /Volumes/HardDrive/go/src/github.com/lucmichalski/cars-dataset/public ubuntu@51.91.21.67:/home/ubuntu/cars-dataset/
	- https://github.com/jhaoheng/golang_selenium_training/blob/master/seleniumChromeDriver/chrome/chromeInit.go#L97
	- find . -name "cache*.html" -print0 | xargs -0 rm
	- for i in *.jpg; do rm "$i"; done
	- yolo_mark train2017_set common.txt common.names
	- find . -maxdepth 1 -type f | wc -l
	- ls -F | grep -v / | wc -l
*/

func Extract(cfg *config.Config) error {

	/*
  	// OSS's default storage is directory `public`, change it to S3
	oss.Storage = s3.New(&s3.Config{
		AccessID: "access_id", 
		AccessKey: "access_key", 
		Region: "region", 
		Bucket: "bucket", 
		Endpoint: "cdn.getqor.com", 
		ACL: awss3.BucketCannedACLPublicRead,
	})
	*/

	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless",
			"--no-sandbox",
			"--start-maximized",
			"--window-size=1024,768",
			"--disable-crash-reporter",
			"--hide-scrollbars",
			"--disable-gpu",
	        	"--disable-setuid-sandbox",
	        	"--disable-infobars",
	        	"--window-position=0,0",
	        	"--ignore-certifcate-errors",
	        	"--ignore-certifcate-errors-spki-list",
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
			"--proxy-server=http://tor-haproxy:8119",
    			// "--host-resolver-rules=\"MAP * 0.0.0.0 , EXCLUDE localhost\"",
		},
	}
	caps.AddChrome(chromeCaps)

	caps.SetLogLevel(slog.Server, slog.Off)
	caps.SetLogLevel(slog.Browser, slog.Off)
	caps.SetLogLevel(slog.Client, slog.Off)
	caps.SetLogLevel(slog.Driver, slog.Off)
	caps.SetLogLevel(slog.Performance, slog.Off)
	caps.SetLogLevel(slog.Profiler, slog.Off)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://selenium:%d/wd/hub", 4444))
	if err != nil {
		return err
	}
	defer wd.Quit()

	// wd.AddCookie()

	/*
	err = wd.SetImplicitWaitTimeout(time.Second * 2)
	if err != nil {
		return err
	}	

	err = wd.SetPageLoadTimeout(time.Second * 2)
	if err != nil {
		return err
	}
	*/

	var links []string
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
			if strings.HasSuffix(sitemap, ".gz") {
				log.Infoln("extract sitemap gz compressed...")
				locs, err := prefetch.ExtractSitemapGZ(sitemap)
				if err != nil {
					log.Fatal("ExtractSitemapGZ: ", err, "sitemap: ",sitemap)
					return err
				}
				utils.Shuffle(locs)
				for _, loc := range locs {
					if strings.HasPrefix(loc, "https://classics.autotrader.com/classic-cars") {
						links = append(links, loc)
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
					if strings.HasPrefix(loc, "https://classics.autotrader.com/classic-cars") {
						links = append(links, loc)
					}
				}				
			}
		}
	} else {
		links = append(links, cfg.URLs...)
	}	

	pp.Println("found:", len(links))

	t := throttler.New(1, len(links))

	utils.Shuffle(links)
	for _, link := range links {
		log.Println("processing link:", link)
		go func(link string) error {
			defer t.Done(nil)
			err := scrapeSelenium(link, cfg, wd)
			if err != nil {
				log.Warnln(err)
			}
			return err
		}(link)
		t.Throttle()
	}

	// throttler errors iteration
	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			log.Printf("error #%d: %s", i, err)
		}
		log.Fatal(t.Err())
	}
	
	// }

	return nil
}

func scrapeSelenium(url string, cfg *config.Config, wd selenium.WebDriver) (error) {

	err := wd.Get(url)
	if err != nil {
		return err
	}

	/*
	src, err := wd.PageSource()
	if err != nil {
		return err
	}
	fmt.Println("source", src)
	*/

	// check in the databse if exists
	var vehicleExists models.Vehicle
	if !cfg.DB.Where("url = ?", url).First(&vehicleExists).RecordNotFound() {
		fmt.Printf("skipping url=%s as already exists\n", url)
		return nil
	}

	// create vehicle 
	vehicle := &models.Vehicle{}
	vehicle.URL = url
	vehicle.Source = "classics.autotrader.com"
	vehicle.Class = "car"

	// write email
	makeCnt, err := wd.FindElement(selenium.ByCSSSelector, "ol.breadcrumbs li:first-child")
	if err != nil {
		return err
	}

	make, err := makeCnt.Text()
	if err != nil {
		return err
	}
	pp.Println("make:", make)

	modelCnt, err := wd.FindElement(selenium.ByCSSSelector, "ol.breadcrumbs li:nth-of-type(n+2)")
	if err != nil {
		return err
	}

	model, err := modelCnt.Text()
	if err != nil {
		return err
	}
	pp.Println("model:", model)

	yearCnt, err := wd.FindElement(selenium.ByCSSSelector, "ol.breadcrumbs li:nth-of-type(n+3)")
	if err != nil {
		return err
	}

	year, err := yearCnt.Text()
	if err != nil {
		return err
	}
	pp.Println("year:", year)	

	vehicle.Manufacturer = make
	vehicle.Year = year
	vehicle.Modl = model
	vehicle.Name = make + " " + model + " " + year

	imagesCnt, err := wd.FindElements(selenium.ByCSSSelector, ".vdp-gallery-secondary-slides img")
	if err != nil {
		return err
	}

	for _ , imageCnt := range imagesCnt {
		image, err := imageCnt.GetAttribute("src")
		if err != nil {
			continue
		}
		// https://0.cdn.autotraderspecialty.com/2019-BMW-C400X-motorcycle--Motorcycle-200865678-93f9b5e8be16c827321aff99dde0f97f.jpg?r=pad&w=735&h=551&c=%23f5f5f5
		// https://0.cdn.autotraderspecialty.com/2019-BMW-C400X-motorcycle--Motorcycle-200865678-26a69225bdba04f38cfd5a234c03b6f4.jpg?r=pad&w=143&h=107&c=%23f5f5f5
		image = strings.Replace(image, "w=143", "w=735", -1)
		image = strings.Replace(image, "h=107", "h=551", -1)
		pp.Println("image:", image)

		if image == "" {
			continue
		}

		proxyURL := fmt.Sprintf("http://51.91.21.67:9005/crop?url=%s", image)
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

			image := models.VehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum, Source: image}

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
		return nil
	}

	pp.Println(vehicle)

	// save vehicle
	if !cfg.DryMode {
		if err := cfg.DB.Create(&vehicle).Error; err != nil {
			log.Fatalf("create vehicle (%v) failure, got err %v", vehicle, err)
			return err
		}
	}	

	return nil
}
