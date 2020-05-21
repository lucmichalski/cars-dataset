package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"github.com/k0kubun/pp"
	// "github.com/gocolly/colly/v2/queue"
	// "github.com/gocolly/twocaptcha"
	"github.com/lucmichalski/cars-dataset/pkg/selenium"
	"github.com/lucmichalski/cars-dataset/pkg/selenium/chrome"
	slog "github.com/lucmichalski/cars-dataset/pkg/selenium/log"
	// "github.com/lucmichalski/cars-dataset/pkg/utils"
)

var sitemapURLs = []string{
	"https://www.similarweb.com/sitemaps/website/website-index.xml.gz",
	"https://www.similarweb.com/sitemaps/website/top-website-index.xml.gz",
	"https://www.similarweb.com/sitemaps/app/top-app-index.xml.gz",
	"https://www.similarweb.com/sitemaps/app/app-index.xml.gz",
}

func main() {

	// Instantiate default collector
	c := colly.NewCollector(
		colly.Async(true),
		colly.CacheDir("../shared/data"),
		colly.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"),
	)

	// Rotate 25 socks5 proxies behing haproxy
	rp, err := proxy.RoundRobinProxySwitcher("sock5://localhost:5566") // 1080 // 8119 // 5566
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*similarweb.*",
		Parallelism: 1,
		RandomDelay: 15 * time.Second,
	})

	/*
		q, _ := queue.New(
			6,
			&queue.InMemoryQueueStorage{
				MaxSize: 10000000,
			},
		)
	*/

	c.DisableCookies()

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
			// "--proxy-server=http://localhost:1080",
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

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://127.0.0.1:%d/wd/hub", 4444))
	if err != nil {
		log.Fatal(err)
	}
	defer wd.Quit()

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "\nError:", err)
	})

	// Print the response
	c.OnResponse(func(r *colly.Response) {
		log.Printf("Proxy Address: %s\n", r.Request.ProxyURL)
		log.Println("Body", string(r.Body))
	})

	/*
		// Create a callback on the XPath query searching for the URLs
		c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
			fmt.Println(e.Text)
			q.AddURL(e.Text)
		})

		// Create a callback on the XPath query searching for the URLs
		c.OnXML("//sitemapindex/sitemap/loc", func(e *colly.XMLElement) {
			fmt.Println(e.Text)
			q.AddURL(e.Text)
		})
	*/

	if _, err := os.Stat("top-1m.csv"); !os.IsNotExist(err) {
		file, err := os.Open("top-1m.csv")
		if err != nil {
			log.Fatal(err)
		}

		reader := csv.NewReader(file)
		reader.Comma = ','
		reader.LazyQuotes = true
		data, err := reader.ReadAll()
		if err != nil {
			log.Fatal(err)
		}

		// utils.Shuffle(data)
		for _, loc := range data {
			scrapeSelenium("https://www.similarweb.com/website/"+loc[1], wd)
			time.Sleep(2 * time.Second)
		}
	}

	// sitemaps are not working
	// for _, sitemap := range sitemapURLs {
	//	q.AddURL(sitemap)
	// }

	// q.Run(c)

	// c.Visit("https://www.carvana.com/sitemap.xml")
	// c.Wait()

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func scrapeSelenium(url string, wd selenium.WebDriver) error {

	err := wd.Get(url)
	if err != nil {
		return err
	}

	fmt.Println("url", url)

	f, err := os.Create("test.html")
	checkErr(err)
	defer f.Close()
	str, err := wd.PageSource()
	checkErr(err)
	_, err = f.WriteString(str)
	checkErr(err)
	// fmt.Println(wd.PageSource())
	f.Sync()

	headline, err := wd.FindElement(selenium.ByCSSSelector, "span[itemprop=headline]")
	checkErr(err)
	headline_txt, err := headline.Text()
	checkErr(err)
	pp.Println("headline_txt", headline_txt)

	os.Exit(1)

	overview_date, err := wd.FindElement(selenium.ByCSSSelector, "span[class=\"websiteHeader-dateFull\"]")
	checkErr(err)
	overview_date_txt, err := overview_date.Text()
	checkErr(err)
	pp.Println("overview_date_txt", overview_date_txt)

	global_rank, err := wd.FindElement(selenium.ByXPATH, "//li[contains(@class,\"globalRank\")]/div[contains(@class,\"valueContainer\")]")
	checkErr(err)
	global_rank_txt, err := global_rank.Text()
	checkErr(err)
	pp.Println("global_rank_txt", global_rank_txt)

	country_rank, err := wd.FindElement(selenium.ByXPATH, "//li[contains(@class,\"countryRank\")]/div[contains(@class,\"valueContainer\")]")
	checkErr(err)
	country_rank_txt, err := country_rank.Text()
	checkErr(err)
	pp.Println("country_rank_txt", country_rank_txt)

	category_rank, err := wd.FindElement(selenium.ByXPATH, "//li[contains(@class,\"categoryRank\")]/div[contains(@class,\"valueContainer\")]")
	checkErr(err)
	category_rank_txt, err := category_rank.Text()
	checkErr(err)
	pp.Println("category_rank_txt", category_rank_txt)

	total_visits, err := wd.FindElement(selenium.ByXPATH, "//div[@data-type=\"visits\"]//span[contains(@class,\"countValue\")]")
	checkErr(err)
	total_visits_txt, err := total_visits.Text()
	checkErr(err)
	pp.Println("total_visits_txt", total_visits_txt)

	avg_visit_duration, err := wd.FindElement(selenium.ByXPATH, "//span[@data-type=\"time\"]/span")
	checkErr(err)
	avg_visit_duration_txt, err := avg_visit_duration.Text()
	checkErr(err)
	pp.Println("avg_visit_duration_txt", avg_visit_duration_txt)

	pages_per_visit, err := wd.FindElement(selenium.ByXPATH, "//span[@data-type=\"ppv\"]/span")
	checkErr(err)
	pages_per_visit_txt, err := pages_per_visit.Text()
	checkErr(err)
	pp.Println("pages_per_visit_txt", pages_per_visit_txt)

	bounce_rate, err := wd.FindElement(selenium.ByXPATH, "//span[@data-type=\"bounce\"]/span")
	checkErr(err)
	bounce_rate_txt, err := bounce_rate.Text()
	checkErr(err)
	pp.Println("bounce_rate_txt", bounce_rate_txt)
	//traffic_by_countries_names = [i for i in driver.find_elements_by_xpath('//span[contains(@class,"country-container")]/a')]
	//traffic_by_countries_values = [i for i in driver.find_elements_by_xpath('//span[contains(@class,"traffic-share-value")]/span')]
	//traffic_by_countries = list(zip(traffic_by_countries_names, traffic_by_countries_values))

	//traffic_sources_texts = [i for i in driver.find_elements_by_xpath('//span[@class="trafficSourcesChart-text"] | //a[@class="trafficSourcesChart-reference js-goToSection"]')]
	//traffic_sources_values = [i for i in driver.find_elements_by_xpath('//div[@class="trafficSourcesChart-value"]')]
	//traffic_sources = list(zip(traffic_sources_texts, traffic_sources_values))
	lastamt, err := wd.FindElement(selenium.ByXPATH, "//g[@class=\"highcharts-tooltip\"]/text/tspan[3]")
	checkErr(err)
	lastamt_txt, err := lastamt.Text()
	checkErr(err)
	pp.Println("lastamt_txt", lastamt_txt)

	referrals_percent, err := wd.FindElement(selenium.ByXPATH, "//span[@class=\"subheading-value referrals\"]")
	checkErr(err)
	referrals_percent_txt, err := referrals_percent.Text()
	checkErr(err)
	pp.Println("referrals_percent_txt", referrals_percent_txt)
	//top_referring_sites_names = [i for i in driver.find_elements_by_xpath('//div[@class="referralsSites referring"]//ul[@class="websitePage-list"]//div[@class="websitePage-listItemTitle"]/a')]
	//top_referring_sites_values = [i for i in driver.find_elements_by_xpath('//div[@class="referralsSites referring"]//ul[@class="websitePage-list"]//span[@class="websitePage-trafficShare"]')]
	//top_referring_sites = list(zip(top_referring_sites_names, top_referring_sites_values))
	//top_destination_sites_names = [i for i in driver.find_elements_by_xpath('//div[@class="referralsSites destination"]//ul[@class="websitePage-list"]//div[@class="websitePage-listItemTitle"]/a')]
	//top_destination_sites_values = [i for i in driver.find_elements_by_xpath('//div[@class="referralsSites destination"]//ul[@class="websitePage-list"]//span[@class="websitePage-trafficShare"]')]
	//top_destination_sites = list(zip(top_destination_sites_names, top_destination_sites_values))

	search_percent, err := wd.FindElement(selenium.ByXPATH, "//span[@class=\"subheading-value searchText\"]")
	checkErr(err)
	search_percent_txt, err := search_percent.Text()
	checkErr(err)
	pp.Println("search_percent_txt", search_percent_txt)

	organic_keywords_percent, err := wd.FindElement(selenium.ByXPATH, "//div[@class=\"searchPie-text searchPie-text--left  \"]/span[@class=\"searchPie-number\"]")
	checkErr(err)
	organic_keywords_percent_txt, err := organic_keywords_percent.Text()
	checkErr(err)
	pp.Println("organic_keywords_percent_txt", organic_keywords_percent_txt)

	paid_keywords_percent, err := wd.FindElement(selenium.ByXPATH, "//div[@class=\"searchPie-text searchPie-text--right  \"]/span[@class=\"searchPie-number\"]")
	checkErr(err)
	paid_keywords_percent_txt, err := paid_keywords_percent.Text()
	checkErr(err)
	pp.Println("paid_keywords_percent", paid_keywords_percent_txt)
	// top_5_organic_keywords_words = [i for i in driver.find_elements_by_xpath('//div[contains(@class,"searchKeywords-text searchKeywords-text--left")]//span[@class="searchKeywords-words"]')]
	// top_5_organic_keywords_values = [i for i in driver.find_elements_by_xpath('//div[contains(@class,"searchKeywords-text searchKeywords-text--left")]//span[@class="searchKeywords-trafficShare"]')]
	// top_5_organic_keywords = list(zip(top_5_organic_keywords_words, top_5_organic_keywords_values))
	// top_5_paid_keywords_words = [i for i in driver.find_elements_by_xpath('//div[contains(@class,"searchKeywords-text searchKeywords-text--right")]//span[@class="searchKeywords-words"]')]
	// top_5_paid_keywords_values = [i for i in driver.find_elements_by_xpath('//div[contains(@class,"searchKeywords-text searchKeywords-text--right")]//span[@class="searchKeywords-trafficShare"]')]
	// top_5_paid_keywords = list(zip(top_5_paid_keywords_words, top_5_paid_keywords_values))

	social_percent, err := wd.FindElement(selenium.ByXPATH, "//span[@class=\"subheading-value social\"]")
	checkErr(err)
	social_percent_txt, err := social_percent.Text()
	pp.Println("social_percent", social_percent_txt)
	// social_items_names = [i for i in driver.find_elements_by_xpath('//ul[@class="socialList"]//a[@class="socialItem-title name link"]')]
	// social_items_values = [i for i in driver.find_elements_by_xpath('//ul[@class="socialList"]//div[@class="socialItem-value"]')]
	// social_items = list(zip(social_items_names, social_items_values))
	//try:
	//	display_advertising_percent = driver.find_element_by_xpath('//span[@class="subheading-value display"]')
	//except:
	//	display_advertising_percent = None
	// top_publishers = [i for i in driver.find_elements_by_xpath('//div[@class="websitePage-engagementInfo"]//a[@class="js-tooltipTarget websitePage-listItemLink"]')]

	// website_contents_subdomains_texts = [i for i in driver.find_elements_by_xpath('//div[@class="websiteContent-tableLine"]//span[@class="websiteContent-itemText"]')]
	// website_contents_subdomains_values = [i for i in driver.find_elements_by_xpath('//div[@class="websiteContent-tableLine"]//span[@class="websiteContent-itemPercentage js-value"]')]
	// website_contents_subdomains = list(zip(website_contents_subdomains_texts, website_contents_subdomains_values))

	// also_visited_websites = [i for i in driver.find_elements_by_xpath('//section[contains(@class,"alsoVisitedSection")]//a[@class="js-tooltipTarget websitePage-listItemLink"]')]
	// similarity_sites = [i.get_attribute('data-site') for i in driver.find_elements_by_xpath('//li[@class="similarSitesList-item"]')]

	return nil
}
