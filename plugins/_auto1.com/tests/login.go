package main

import (
	"log"

	"github.com/gocolly/colly/v2"
)

func main() {
	// create a new collector
	c := colly.NewCollector()

	// authenticate
	err := c.Post("https://www.auto1.com/en/en/merchant/signup", map[string]string{"user_signup[email]": "michalskiluc79@gmail.com", "user_signup[password]": "aado33ve79T!"})
	if err != nil {
		log.Fatal(err)
	}

	// attach callbacks after login
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})

	// start scraping
	c.Visit("https://www.auto1.com/")
}
