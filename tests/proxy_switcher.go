package main

import (
	//"bytes"
	"log"
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(colly.AllowURLRevisit())

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("socks5://localhost:1080")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Print the response
	c.OnResponse(func(r *colly.Response) {
		log.Printf("Proxy Address: %s\n", r.Request.ProxyURL)
		log.Println("Body", string(r.Body))
		//log.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
	})

	// Fetch httpbin.org/ip five times
	for i := 0; i < 5; i++ {
		fmt.Println("Visiting")
		c.Visit("https://httpbin.org/ip")
	}

        c.Visit("https://www.carvana.com/robots.txt")

	c.Visit("https://www.carvana.com/sitemap.xml")

}
