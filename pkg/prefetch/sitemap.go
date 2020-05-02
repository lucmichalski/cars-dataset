package prefetch

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
	// "io/ioutil"
	// "os"

	// "github.com/k0kubun/pp"
	"github.com/beevik/etree"
	"github.com/corpix/uarand"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	log "github.com/sirupsen/logrus"
	ccsv "github.com/tsak/concurrent-csv-writer"
	"golang.org/x/net/proxy"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

const (
	torProxyAddress   = "socks5://51.91.21.67:5566"
	torPrivoxyAddress = "socks5://51.91.21.67:8119"
)

func Sitemap(cfg *config.Config) error {

	if cfg == nil {
		return errors.New("Please specify a config for the crawler.")
	}

	if len(cfg.URLs) == 0 {
		return errors.New("Please specify the root url for the crawler.")
	}

	if cfg.CacheDir == "" {
		cfg.CacheDir = "./shared/data"
	}

	if cfg.ConsumerThreads == 0 {
		cfg.ConsumerThreads = runtime.NumCPU()
	}

	// init concurrent csv writer
	// Create `sample.csv` in current directory
	csv, err := ccsv.NewCsvWriter("sitemap.txt")
	if err != nil {
		panic("Could not open `sample.csv` for writing")
	}

	// Flush pending writes and close file upon exit of Sitemap()
	defer csv.Close()

	// Create a Collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.AllowedDomains(cfg.AllowedDomains...),
		colly.CacheDir(cfg.CacheDir),
	)

	// create a request queue with `x` consumer threads
	q, _ := queue.New(
		cfg.ConsumerThreads, // Number of consumer threads
		&queue.InMemoryQueueStorage{
			MaxSize: cfg.QueueMaxSize,
		}, // Use default queue storage
	)

	c.OnError(func(r *colly.Response, err error) {
		log.Errorln("error:", err, r.Request.URL, string(r.Body))
		csv.Write([]string{"error:" + r.Request.URL.String()})
		csv.Flush()
		q.AddURL(r.Request.URL.String())
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
		csv.Write([]string{e.Text})
		csv.Flush()
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
		csv.Write([]string{e.Text})
		csv.Flush()
	})

	if cfg.IsSitemapIndex {

		log.Infoln("extractSitemapIndex...")
		sitemaps, err := ExtractSitemapIndex(cfg.URLs[0])
		if err != nil {
			log.Fatal("ExtractSitemapIndex:", err)
			return err
		}

		for _, sitemap := range sitemaps {
			log.Infoln("processing ", sitemap)
			if strings.Contains(sitemap, ".gz") {
				log.Infoln("extract sitemap gz compressed...")
				locs, err := ExtractSitemapGZ(sitemap)
				if err != nil {
					log.Warnln("ExtractSitemapGZ", err)
					return err
				}
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
			csv.Write([]string{u})
		}
	}

	// Consume URLs
	q.Run(c)

	csv.Close()

	return nil
}

func ExtractSitemapIndex(rawUrl string) ([]string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	tbProxyURL, err := url.Parse(torProxyAddress)
	if err != nil {
		return nil, err
	}

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}
	tbTransport := &http.Transport{
		Dial: tbDialer.Dial,
	}
	client.Transport = tbTransport

	request, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	/*
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// log.Println(string(body))	
	fmt.Println("Body:", string(body))
	os.Exit(1)
	*/

	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(response.Body); err != nil {
		return nil, err
	}
	var urls []string
	index := doc.SelectElement("sitemapindex")
	sitemaps := index.SelectElements("sitemap")
	for _, sitemap := range sitemaps {
		loc := sitemap.SelectElement("loc")
		l := loc.Text()
		l = strings.TrimLeftFunc(l, func(c rune) bool {
			return c == '\r' || c == '\n' || c == '\t'
		})
		l = strings.TrimSpace(l)
		log.Infoln("loc:", l)
		urls = append(urls, l)
	}
	return urls, nil
}

func ExtractSitemapGZ(rawUrl string) ([]string, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	tbProxyURL, err := url.Parse(torProxyAddress)
	if err != nil {
		return nil, err
	}

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}
	tbTransport := &http.Transport{
		Dial: tbDialer.Dial,
	}
	client.Transport = tbTransport
	request, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	var reader io.ReadCloser
	reader, err = gzip.NewReader(response.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer reader.Close()

	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(reader); err != nil {
		return nil, err
	}
	var urls []string
	urlset := doc.SelectElement("urlset")
	entries := urlset.SelectElements("url")
	for _, entry := range entries {
		loc := entry.SelectElement("loc")
		l := loc.Text()
		l = strings.TrimLeftFunc(l, func(c rune) bool {
			return c == '\r' || c == '\n' || c == '\t'
		})
		l = strings.TrimSpace(l)
		log.Infoln("loc:", l)
		urls = append(urls, l)
	}
	return urls, err
}

func ExtractSitemap(rawUrl string) ([]string, error) {

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	tbProxyURL, err := url.Parse(torProxyAddress)
	if err != nil {
		return nil, err
	}
	// pp.Println(tbProxyURL)

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}
	tbTransport := &http.Transport{
		Dial: tbDialer.Dial,
	}
	client.Transport = tbTransport

	// client := new(http.Client)
	request, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(response.Body); err != nil {
		panic(err)
	}
	var urls []string
	urlset := doc.SelectElement("urlset")
	if urlset != nil {
		entries := urlset.SelectElements("url")
		for _, entry := range entries {
			loc := entry.SelectElement("loc")
			l := loc.Text()
			l = strings.TrimLeftFunc(l, func(c rune) bool {
				return c == '\r' || c == '\n' || c == '\t'
			})
			l = strings.TrimSpace(l)
			log.Infoln("loc:", l)
			urls = append(urls, l)
		}
	}
	return urls, err
}
