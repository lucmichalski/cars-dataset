package main

import (
	"compress/gzip"
	"crypto/md5"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"plugin"
	"bytes"
	"html/template"

	"github.com/thanhhh/gin-gonic-realip"
	"github.com/gin-gonic/gin"
	"github.com/qor/qor/utils"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/assetfs"
	"github.com/h2non/filetype"
	"github.com/PuerkitoBio/goquery"
	"github.com/beevik/etree"
	"github.com/corpix/uarand"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/qor/validations"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	// "github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var (
	isHelp        bool
	isVerbose     bool
	isAdmin       bool
	isCrawl       bool
	isDataset     bool
	isTruncate    bool
	isClean       bool
	isCatalog     bool
	isDryMode     bool
	isNoCache     bool
	isExtract     bool
	parallelJobs  int
	pluginDir     string
	cacheDir      string
	usePlugins    []string
	queueMaxSize = 100000000
	cachePath    = "./data/cache"
)

func main() {

	listPlugins, err := filepath.Glob("./release/*.so")
	if err != nil {
		panic(err)
	}
	var defaultPlugins []string
	for _, p := range listPlugins {
		p = strings.Replace(p, ".so", "", -1)
		p = strings.Replace(p, "peaks-tires-", "", -1)
		p = strings.Replace(p, "release/", "", -1)
		defaultPlugins = append(defaultPlugins, p)
	}

	pflag.BoolVarP(&isDryMode, "dry-mode", "", false, "do not insert data into database tables.")
	pflag.BoolVarP(&isCatalog, "catalog", "", false, "import datasets/catalogs.")
	pflag.StringVarP(&pluginDir, "plugin-dir", "", "./release", "plugins directory.")
	pflag.StringSliceVarP(&usePlugins, "plugins", "", defaultPlugins, "plugins to load.")
	pflag.IntVarP(&parallelJobs, "parallel-jobs", "j", 35, "parallel jobs.")
	pflag.BoolVarP(&isCrawl, "crawl", "c", false, "launch the crawler.")
	pflag.BoolVarP(&isDataset, "dataset", "d", false, "launch the crawler.")
	pflag.BoolVarP(&isClean, "clean", "", false, "auto-clean temporary files.")
	pflag.BoolVarP(&isAdmin, "admin", "", false, "launch the admin interface.")
	pflag.BoolVarP(&isTruncate, "truncate", "t", false, "truncate table content.")
	pflag.BoolVarP(&isExtract, "extract", "e", false, "extract data from urls.")
	pflag.StringVarP(&cacheDir, "cache-dir", "", "./shared/data", "cache directory.")
	pflag.BoolVarP(&isNoCache, "no-cache", "", false, "disable crawler cache.")
	pflag.BoolVarP(&isVerbose, "verbose", "v", false, "verbose mode.")
	pflag.BoolVarP(&isHelp, "help", "h", false, "help info.")
	pflag.Parse()
	if isHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// Instanciate the sqlite3 client

	DB, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4,utf8&parseTime=True", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE")))
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// callback for images and validation
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)

	// truncate table
	if isTruncate {
		if err := DB.DropTableIfExists(&vehicle{}).Error; err != nil {
			panic(err)
		}
		if err := DB.DropTableIfExists(&vehicleImage{}).Error; err != nil {
			panic(err)
		}
	}

	// migrate tables
	DB.AutoMigrate(&vehicle{})
	DB.AutoMigrate(&vehicleImage{})
	DB.AutoMigrate(&media_library.MediaLibrary{})

	// load plugins
	ptPlugins := plugins.New()

	// The plugins (the *.so files) must be in a 'release' sub-directory
	allPlugins, err := filepath.Glob(pluginDir + "/*.so")
	if err != nil {
		panic(err)
	}

	var loadPlugins []string
	if len(usePlugins) > 0 {
		for _, p := range allPlugins {
			for _, u := range usePlugins {
				if strings.Contains(p, u) {
					loadPlugins = append(loadPlugins, p)
				}
			}
		}
	} else {
		loadPlugins = allPlugins
	}

	// register commands from plugins
	for _, filename := range loadPlugins {
		p, err := plugin.Open(filename)
		if err != nil {
			panic(err)
		}
		// lookup for symbols
		cmdSymbol, err := p.Lookup(plugins.CmdSymbolName)
		if err != nil {
			fmt.Printf("plugin %s does not export symbol \"%s\"\n",
				filename, plugins.CmdSymbolName)
			continue
		}
		// check if symbol is implemented in Plugins interface
		commands, ok := cmdSymbol.(plugins.Plugins)
		if !ok {
			fmt.Printf("Symbol %s (from %s) does not implement Plugins interface\n",
				plugins.CmdSymbolName, filename)
			continue
		}
		// initialize plugin
		if err := commands.Init(ptPlugins.Ctx); err != nil {
			fmt.Printf("%s initialization failed: %v\n", filename, err)
			continue
		}
		// register commands from plugin
		for name, cmd := range commands.Registry() {
			ptPlugins.Commands[name] = cmd
		}
	}

	// migrate table from plugins
	for _, cmd := range ptPlugins.Commands {
		for _, table := range cmd.Migrate() {
			DB.AutoMigrate(table)
		}
	}

	if isExtract {
		fmt.Print("extracting...\n")
		for _, cmd := range ptPlugins.Commands {
			fmt.Printf(" from %s", cmd.Name())
			c := cmd.Config()
			if !isNoCache {
				c.CacheDir = cacheDir
			}
			c.IsDebug = true
			c.ConsumerThreads = 4
			pp.Println(c)
			c.DB = DB
			cmd.Crawl(c)
		}
	}

	// import catalog
	if isCatalog {
		for _, cmd := range ptPlugins.Commands {
			c := cmd.Config()
			c.DB = DB
			c.DryMode = isDryMode			
			err := cmd.Catalog(c)
			if err != nil {
				panic(err)
			}
		}
	   os.Exit(0)
	}

	if isAdmin {

		// Initialize AssetFS
		AssetFS := assetfs.AssetFS().NameSpace("admin")

		// Register custom paths to manually saved views
		AssetFS.RegisterPath(filepath.Join(utils.AppRoot, "./templates/qor/admin/views"))
		AssetFS.RegisterPath(filepath.Join(utils.AppRoot, "./templates/qor/media/views"))

		// Initialize Admin
		Admin := admin.New(&admin.AdminConfig{
			SiteName: "Cars Dataset",
			DB:       DB,
			AssetFS:  AssetFS,
		})

		Admin.AddMenu(&admin.Menu{Name: "Crawl Management", Priority: 1})

		// Add VehicleImage as Media Libraray
		VehicleImagesResource := Admin.AddResource(&vehicleImage{}, &admin.Config{Menu: []string{"Crawl Management"}, Priority: -1})

		VehicleImagesResource.Filter(&admin.Filter{
			Name:       "SelectedType",
			Label:      "Media Type",
			Operations: []string{"contains"},
			Config:     &admin.SelectOneConfig{Collection: [][]string{{"video", "Video"}, {"image", "Image"}, {"file", "File"}, {"video_link", "Video Link"}}},
		})
		VehicleImagesResource.IndexAttrs("File", "Title")
 
		VehicleImagesResource.UseTheme("grid")

		cars := Admin.AddResource(&vehicle{}, &admin.Config{Menu: []string{"Crawl Management"}})
		cars.IndexAttrs("ID", "Name", "Modl", "Engine", "Year", "Source", "Manufacturer", "MainImage", "Images")

		cars.Meta(&admin.Meta{Name: "MainImage", Config: &media_library.MediaBoxConfig{
			RemoteDataResource: VehicleImagesResource,
			Max:                1,
			//Sizes: map[string]*media.Size{
			//	"main": {Width: 560, Height: 700},
			//},
		}})
		cars.Meta(&admin.Meta{Name: "MainImageURL", Valuer: func(record interface{}, context *qor.Context) interface{} {
			if p, ok := record.(*vehicle); ok {
				result := bytes.NewBufferString("")
				tmpl, _ := template.New("").Parse("<img src='{{.image}}'></img>")
				tmpl.Execute(result, map[string]string{"image": p.MainImageURL()})
				return template.HTML(result.String())
			}
			return ""
		}})

		// initalize an HTTP request multiplexer
		mux := http.NewServeMux()

		// Mount admin interface to mux
		Admin.MountTo("/admin", mux)

		router := gin.Default()

		router.Use(realip.RealIP())

		// add basic auth
		admin := router.Group("/admin", gin.BasicAuth(gin.Accounts{"cars": "cars"}))
		{
			admin.Any("/*resources", gin.WrapH(mux))
		}

		router.Static("/system", "./public/system")
		router.Static("/public", "./public")

		fmt.Println("Listening on: 9008")
		log.Fatal(router.Run(fmt.Sprintf("%s:%s", "", "9008")))
		// http.ListenAndServe(":9008", mux)

	}

	if isDataset {

		sName := "dataset.txt"
		sfile, err := os.Create(sName)
		if err != nil {
			log.Fatalf("Cannot create file %q: %s\n", sName, err)
		}
		defer sfile.Close()

		_, err = sfile.WriteString("name;image_path\n")
		if err != nil {
			log.Fatal(err)
		}

		// Scan
		type res struct {
			Name   string
			Images string
		}

		type entryProperty struct {
			ID          int
			Url         string
			VideoLink   string
			FileName    string
			Description string
		}

		var results []res
		DB.Raw("select name, images FROM vehicles").Scan(&results)
		for _, result := range results {
			if result.Images == "" {
				continue
			}

			var ep []entryProperty
			fmt.Println(result.Images)
			if err := json.Unmarshal([]byte(result.Images), &ep); err != nil {
				log.Fatalln("unmarshal error, ", err)
			}
			pp.Println(ep)

			if len(ep) < 2 {
				continue
			}

			prefixPath := filepath.Join("./", "datasets", "cars", result.Name)
			os.MkdirAll(prefixPath, 0755)
			pp.Println("prefixPath:", prefixPath)

			for _, entry := range ep {
				sourceFile := filepath.Join("./", "public", entry.Url)
				pp.Println("sourceFile:", sourceFile)

				input, err := ioutil.ReadFile(sourceFile)
				if err != nil {
					log.Fatalln("reading file error, ", err)
				}

				destinationFile := filepath.Join(prefixPath, strconv.Itoa(entry.ID)+"-"+filepath.Base(entry.Url))
				err = ioutil.WriteFile(destinationFile, input, 0644)
				if err != nil {
					log.Fatalln("creating file error, ", err)
				}
				pp.Println("destinationFile:", destinationFile)
				_, err = sfile.WriteString(fmt.Sprintf("%s;%s\n", result.Name, destinationFile))
				if err != nil {
					log.Fatal(err)
				}
				sfile.Sync()
			}
		}
		os.Exit(0)
	}

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.CacheDir(cachePath),
		/*
			colly.URLFilters(
				regexp.MustCompile("https://autosphere\\.fr/(|e.+)$"),
				regexp.MustCompile("https://www.autosphere\\.fr/h.+"),
			),
		*/
	)

	// create a request queue with 1 consumer thread
	q, _ := queue.New(
		parallelJobs, // Number of consumer threads set to 1 to avoid dead lock on database
		&queue.InMemoryQueueStorage{
			MaxSize: queueMaxSize,
		}, // Use default queue storage
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
		var vehicleExists vehicle
		if !DB.Where("url = ?", e.Request.Ctx.Get("url")).First(&vehicleExists).RecordNotFound() {
			fmt.Printf("skipping url=%s as already exists\n", e.Request.Ctx.Get("url"))
			return
		}

		vehicle := &vehicle{}
		vehicle.URL = e.Request.Ctx.Get("url")
		modele := e.ChildText("span[class=modele]")
		if isVerbose {
			fmt.Println("modele:", modele)
		}
		if modele == "" {
			return
		}
		version := e.ChildText("span[class=version]")
		if isVerbose {
			fmt.Println("version:", version)
		}

		var carInfo vehicleGtm
		e.ForEach(`div[id=gtm_goal]`, func(_ int, el *colly.HTMLElement) {
			info := el.Attr("data-gtm-goal")
			infoParts := strings.Split(info, "--**--")
			if len(infoParts) > 0 {
				if infoParts[0] != "" {
					if err := json.Unmarshal([]byte(infoParts[0]), &carInfo); err != nil {
						log.Fatalln("unmarshal error, ", err)
					}
				}
				if isVerbose {
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
		vehicle.Source = "autosphere.fr"

		vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Price", Value: carInfo.ProductPrice})
		vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Transmission", Value: carInfo.ProductTransmission})

		// Pictures
		var carImgLinks []string
		e.ForEach(`div[class=swiper-slide] > img`, func(_ int, el *colly.HTMLElement) {
			carPicSrc := el.Attr("src")
			carPicDataSrc := el.Attr("data-src")
			if isVerbose {
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

		carImgLinks = removeDuplicates(carImgLinks)
		if isVerbose {
			pp.Println(carImgLinks)
		}

		if len(carImgLinks) == 0 {
			return
		}

		for _, carImage := range carImgLinks {
			// download and scan image
			// crop car
			// resize image
			if carImage == "" {
				continue
			}

			proxyURL := fmt.Sprintf("http://darknet2:9004/crop?url=%s", carImage)
			log.Println("proxyURL:", proxyURL)
			if file, size, checksum, err := openFileByURL(proxyURL); err != nil {
				fmt.Printf("open file failure, got err %v", err)
			} else {
				defer file.Close()

				if size < 40000 {
					if isClean {
						// delete tmp file
						err := os.Remove(file.Name())
						if err != nil {
							log.Fatal(err)
						}
					}
					log.Infoln("----> Skipping file: ", file.Name(), "size: ", size)					
					continue
				}

				image := vehicleImage{Title: vehicle.Name, SelectedType: "image", Checksum: checksum}

				log.Println("----> Scanning file: ", file.Name(), "size: ", size)
				if err := image.File.Scan(file); err != nil {
					log.Fatalln("image.File.Scan, err:", err)
					continue
				}

				// transaction
				if err := DB.Create(&image).Error; err != nil {
					log.Fatalln("create image (%v) failure, got err %v\n", image, err)
					continue
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

				if isClean {
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
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Manufacturer", Value: manufacturer})
			case "Couleur":
				color = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Color", Value: color})
			case "Modèle":
				model = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Model", Value: model})
			case "Boîte de vitesse":
				gearbox = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "GearBox", Value: gearbox})
			case "Année":
				year = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Year", Value: year})
			case "Puissance Fiscale":
				power = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "HorsePower", Value: power})
			case "Type de véhicule":
				carType = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "CarType", Value: carType})
			case "Certificat CRIT'AIR":
				certCritAir = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "CRIT'AIR Certificat", Value: certCritAir})
			case "Co2":
				c02 = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Co2", Value: c02})
			case "Puissance Réelle":
				realPower = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "RealPower", Value: realPower})
			case "Carburant":
				gas = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "GasType", Value: gas})
			case "Portes":
				doors = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Doors", Value: doors})
			case "Places":
				places = strings.TrimSpace(texts[1])
				vehicle.VehicleProperties = append(vehicle.VehicleProperties, vehicleProperty{Name: "Places", Value: places})
			}

		})

		if isVerbose {
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

		if err := DB.Create(&vehicle).Error; err != nil {
			log.Fatalf("create vehicle (%v) failure, got err %v", vehicle, err)
			return
		}

		log.Infoln("Add manufacturer: ", carInfo.ProductBrand, ", Model:", carInfo.ProductModele, ", Year:", carInfo.ProductYear)

	})

	c.OnResponse(func(r *colly.Response) {
		if isVerbose {
			fmt.Println("OnResponse from", r.Ctx.Get("url"))
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		//if isVerbose {
		fmt.Println("Visiting", r.URL.String())
		//}
		r.Ctx.Put("url", r.URL.String())
	})

	// Start scraping on https://www.autosphere.fr
	log.Infoln("extractSitemapIndex...")
	sitemaps, err := extractSitemapIndex("https://www.autosphere.fr/sitemap.xml")
	if err != nil {
		log.Fatal("ExtractSitemapIndex:", err)
	}

	shuffle(sitemaps)
	for _, sitemap := range sitemaps {
		log.Infoln("processing ", sitemap)
		if strings.Contains(sitemap, ".gz") {
			log.Infoln("extract sitemap gz compressed...")
			locs, err := extractSitemapGZ(sitemap)
			if err != nil {
				log.Fatal("ExtractSitemapGZ", err)
			}
			shuffle(locs)
			for _, loc := range locs {
				q.AddURL(loc)
			}
		} else {
			q.AddURL(sitemap)
		}
	}

	// Consume URLs
	q.Run(c)

}

type vehicle struct {
	gorm.Model
	URL               string `gorm:"index:url"`
	Name              string `gorm:"index:name"`
	Modl              string `gorm:"index:modl"`
	Engine            string `gorm:"index:engine"`
	Year              string `gorm:"index:year"`
	Source            string `gorm:"index:source"`
	Gid               string `gorm:"index:gid"`
	Manufacturer      string `gorm:"index:manufacturer"`
	MainImage         media_library.MediaBox
	Images            media_library.MediaBox
	VehicleProperties vehicleProperties `sql:"type:text"`
}

func (v vehicle) MainImageURL(styles ...string) string {
	style := "original"
	if len(styles) > 0 {
		style = styles[0]
	}

	if len(v.MainImage.Files) > 0 {
		return v.MainImage.URL(style)
	}
	return "/images/no_image.png"
}

type vehicleGtm struct {
	ProductBrand        string `json:"ProductBrand"`
	ProductDistance     int    `json:"ProductDistance"`
	ProductFuel         string `json:"ProductFuel"`
	ProductKilometrage  string `json:"ProductKilometrage"`
	ProductModele       string `json:"ProductModele"`
	ProductPrice        string `json:"ProductPrice"`
	ProductTransmission string `json:"ProductTransmission"`
	ProductYear         string `json:"ProductYear"`
	Event               string `json:"event"`
	ID                  string `json:"id"`
}

type vehicleProperties []vehicleProperty

type vehicleProperty struct {
	Name  string
	Value string
}

func (vehicleProperties *vehicleProperties) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, vehicleProperties)
	case string:
		if v != "" {
			return vehicleProperties.Scan([]byte(v))
		}
	default:
		return errors.New("not supported")
	}
	return nil
}

func (vehicleProperties vehicleProperties) Value() (driver.Value, error) {
	if len(vehicleProperties) == 0 {
		return nil, nil
	}
	return json.Marshal(vehicleProperties)
}

type vehicleImage struct {
	gorm.Model
	Title        string
	Checksum     string
	SelectedType string
	File         media_library.MediaLibraryStorage `sql:"size:4294967295;" media_library:"url:/system/{{class}}/{{primary_key}}/{{column}}.{{extension}}"`
}

func (vehicleImage vehicleImage) Validate(db *gorm.DB) {
	if strings.TrimSpace(vehicleImage.Title) == "" {
		db.AddError(validations.NewError(vehicleImage, "Title", "Title can not be empty"))
	}
}

func (vehicleImage *vehicleImage) SetSelectedType(typ string) {
	vehicleImage.SelectedType = typ
}

func (vehicleImage *vehicleImage) GetSelectedType() string {
	return vehicleImage.SelectedType
}

func (vehicleImage *vehicleImage) ScanMediaOptions(mediaOption media_library.MediaOption) error {
	if bytes, err := json.Marshal(mediaOption); err == nil {
		return vehicleImage.File.Scan(bytes)
	} else {
		return err
	}
}

func (vehicleImage *vehicleImage) GetMediaOption() (mediaOption media_library.MediaOption) {
	mediaOption.Video = vehicleImage.File.Video
	mediaOption.FileName = vehicleImage.File.FileName
	mediaOption.URL = vehicleImage.File.URL()
	mediaOption.OriginalURL = vehicleImage.File.URL("original")
	mediaOption.CropOptions = vehicleImage.File.CropOptions
	mediaOption.Sizes = vehicleImage.File.GetSizes()
	mediaOption.Description = vehicleImage.File.Description
	return
}

func openFileByURL(rawURL string) (*os.File, int64, string, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, 0, "", err
	} else {
		q := fileURL.Query()
		var segments []string
		if q.Get("url") != "" {
			segments = strings.Split(q.Get("url"), "/")
		} else {
			path := fileURL.Path
			segments = strings.Split(path, "/")
		} 

		fileName := getMD5Hash(rawURL) + "-" + segments[len(segments)-1]
		filePath := filepath.Join(os.TempDir(), fileName)

		/*
		if fi, err := os.Stat(filePath); err == nil {
			file, err := os.Open(filePath)
			checksum, err := getMD5File(filePath)
			if err != nil {
				return file, 0, "", err
			}
			return file, fi.Size(), checksum, err
		}
		*/

		file, err := os.Create(filePath)
		if err != nil {
			return file, 0, "", err
		}

		check := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := check.Get(rawURL) // add a filter to check redirect
		if err != nil {
			return file, 0, "", err
		}
		defer resp.Body.Close()
		fmt.Printf("----> Downloaded %v\n", rawURL)

		fmt.Println("Content-Length:", resp.Header.Get("Content-Length"))

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return file, 0, "", err
		}

		buf, _ := ioutil.ReadFile(file.Name())
		kind, _ := filetype.Match(buf)
		pp.Println("kind: ", kind)

		fi, err := file.Stat()
		if err != nil {
			return file, 0, "", err
		}

		checksum, err := getMD5File(filePath)
		if err != nil {
			return file, 0, "", err
		}

		return file, fi.Size(), checksum, nil
	}
}

func shuffle(slice interface{}) {
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	length := rv.Len()
	for i := length - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
}

func getMD5File(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil))
	return
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func extractSitemapIndex(url string) ([]string, error) {
	client := new(http.Client)
	request, err := http.NewRequest("GET", url, nil)
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
		return nil, err
	}
	var urls []string
	index := doc.SelectElement("sitemapindex")
	sitemaps := index.SelectElements("sitemap")
	for _, sitemap := range sitemaps {
		loc := sitemap.SelectElement("loc")
		log.Infoln("loc:", loc.Text())
		urls = append(urls, loc.Text())
	}
	return urls, nil
}

func extractSitemapGZ(url string) ([]string, error) {
	client := new(http.Client)
	request, err := http.NewRequest("GET", url, nil)
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
		panic(err)
	}
	var urls []string
	urlset := doc.SelectElement("urlset")
	entries := urlset.SelectElements("url")
	for _, entry := range entries {
		loc := entry.SelectElement("loc")
		log.Infoln("loc:", loc.Text())
		urls = append(urls, loc.Text())
	}
	return urls, err
}

func removeDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

