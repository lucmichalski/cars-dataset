package main

import (
	//"bytes"
	"log"
	"fmt"
	// "time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
        "github.com/gocolly/colly/v2/queue"
	//"github.com/gocolly/colly/v2/debug"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// colly.AllowURLRevisit()
		// Attach a debugger to the collector
		// colly.Debugger(&debug.LogDebugger{}),
		colly.CacheDir("../shared/data"),
		//colly.Async(true),
	)

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("socks5://localhost:1080")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	//c.Limit(&colly.LimitRule{
		// DomainGlob:  "*httpbin.*",
	//	Parallelism: 1,
	//	RandomDelay: 15 * time.Second,
	//})

	q, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{
			MaxSize: 10000000,
		},
	)

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Print the response
	c.OnResponse(func(r *colly.Response) {
		// log.Printf("Proxy Address: %s\n", r.Request.ProxyURL)
		log.Println("Body", string(r.Body))
		//log.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
	})

        // Create a callback on the XPath query searching for the URLs
        c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
                fmt.Println(e.Text)
		q.AddURL(e.Text)
                // c.Visit(e.Text)
        })

        // Create a callback on the XPath query searching for the URLs
        c.OnXML("//sitemapindex/sitemap/loc", func(e *colly.XMLElement) {
                fmt.Println(e.Text)
		// c.Visit(e.Text)
		q.AddURL(e.Text)
        })

	q.AddURL("https://www.carvana.com/sitemap.xml")
	// c.Visit("https://www.carvana.com/sitemap.xml")
	// c.Wait()
	q.Run(c)
}
