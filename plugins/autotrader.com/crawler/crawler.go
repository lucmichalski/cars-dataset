package crawler

import (
	// "encoding/json"
	"fmt"
	"strings"
	// "time"

	"github.com/k0kubun/pp"
	// "github.com/corpix/uarand"
	// "github.com/qor/media/media_library"
	log "github.com/sirupsen/logrus"
	"github.com/x0rzkov/selenium"
	"github.com/x0rzkov/selenium/chrome"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	// "github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/utils"
	"github.com/lucmichalski/cars-dataset/pkg/prefetch"
)

/*
	- cd plugins/autotrader.com && GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ../../release/cars-dataset-autotrader.com.so ; cd ../..
*/


func Extract(cfg *config.Config) error {

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
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
		},
	}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		return err
	}
	defer wd.Quit()

	urls := []string{
		"https://motorcycles.autotrader.com/motorcycles/2019/bmw/c400x/200865678",
		"https://motorcycles.autotrader.com/motorcycles/2012/bmw/k1300s/200853800",
		"https://motorcycles.autotrader.com/motorcycles/2016/harley_davidson/trike/200618472",
		"https://motorcycles.autotrader.com/motorcycles/2000/american_motorcycle/other_american_motorcycle_models/200887118",
	}

	for _, url := range urls {
		scrapeSelenium(url, cfg, wd)
	}

	// if cfg.IsSitemapIndex {
	log.Infoln("extractSitemapIndex...")
	sitemaps, err := prefetch.ExtractSitemapIndex("https://motorcycles.autotrader.com/sitemap.xml")
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
					scrapeSelenium(loc, cfg, wd)
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
				if strings.HasPrefix(loc, "https://motorcycles.autotrader.com/motorcycles") {
					scrapeSelenium(loc, cfg, wd)
				}
			}				
		}
	}	
	// }

	return nil
}

func scrapeSelenium(url string, cfg *config.Config, wd selenium.WebDriver) (error) {

	wd.Get(url)

	// write email
	makeCnt, err := wd.FindElement(selenium.ByCSSSelector, "ol.breadcrumbs li:first-child")
	if err != nil {
		log.Warnln(err)
	}

	make, err := makeCnt.Text()
	if err != nil {
		log.Warnln(err)
	}
	pp.Println("make:", make)

	modelCnt, err := wd.FindElement(selenium.ByCSSSelector, "ol.breadcrumbs li:nth-of-type(n+2)")
	if err != nil {
		log.Warnln(err)
	}

	model, err := modelCnt.Text()
	if err != nil {
		log.Warnln(err)
	}
	pp.Println("model:", model)

	yearCnt, err := wd.FindElement(selenium.ByCSSSelector, "ol.breadcrumbs li:nth-of-type(n+3)")
	if err != nil {
		log.Warnln(err)
	}

	year, err := yearCnt.Text()
	if err != nil {
		log.Warnln(err)
	}
	pp.Println("year:", year)	

	imagesCnt, err := wd.FindElements(selenium.ByCSSSelector, ".vdp-gallery-secondary-slides img")
	if err != nil {
		log.Warnln(err)
	}

	for _ , imageCnt := range imagesCnt {
		image, err := imageCnt.GetAttribute("src")
		if err != nil {
			log.Warnln(err)
		}
		// https://0.cdn.autotraderspecialty.com/2019-BMW-C400X-motorcycle--Motorcycle-200865678-93f9b5e8be16c827321aff99dde0f97f.jpg?r=pad&w=735&h=551&c=%23f5f5f5
		// https://0.cdn.autotraderspecialty.com/2019-BMW-C400X-motorcycle--Motorcycle-200865678-26a69225bdba04f38cfd5a234c03b6f4.jpg?r=pad&w=143&h=107&c=%23f5f5f5
		image = strings.Replace(image, "w=143", "w=735", -1)
		image = strings.Replace(image, "h=107", "h=551", -1)
		pp.Println("image:", image)
	}	

	return nil
}
