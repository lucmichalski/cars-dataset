package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	"github.com/oschwald/geoip2-golang"
	"github.com/qor/admin"
	"github.com/qor/assetfs"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/validations"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	realip "github.com/thanhhh/gin-gonic-realip"
	ccsv "github.com/tsak/concurrent-csv-writer"

	padmin "github.com/lucmichalski/cars-dataset/pkg/admin"
	"github.com/lucmichalski/cars-dataset/pkg/middlewares"
	"github.com/lucmichalski/cars-dataset/pkg/models"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var (
	isHelp       bool
	isVerbose    bool
	isAdmin      bool
	isCrawl      bool
	isDataset    bool
	isTruncate   bool
	isClean      bool
	isCatalog    bool
	isDryMode    bool
	isNoCache    bool
	isNasha      bool
	isTor        bool
	isExtract    bool
	parallelJobs int
	torAddress   string
	analyzerURL  string
	geoIpFile    string
	pluginDir    string
	cacheDir     string
	usePlugins   []string
	geo          *geoip2.Reader
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
		p = strings.Replace(p, "cars-dataset-", "", -1)
		p = strings.Replace(p, "release/", "", -1)
		defaultPlugins = append(defaultPlugins, p)
	}

	pflag.BoolVarP(&isDryMode, "dry-mode", "", false, "do not insert data into database tables.")
	pflag.BoolVarP(&isCatalog, "catalog", "", false, "import datasets/catalogs.")
	pflag.StringVarP(&pluginDir, "plugin-dir", "", "./release", "plugins directory.")
	pflag.StringVarP(&geoIpFile, "geoip-db", "", "./shared/geoip2/GeoLite2-City.mmdb", "geoip filepath.")
	pflag.StringSliceVarP(&usePlugins, "plugins", "", defaultPlugins, "plugins to load.")
	pflag.IntVarP(&parallelJobs, "parallel-jobs", "j", 35, "parallel jobs.")
	pflag.BoolVarP(&isCrawl, "crawl", "c", false, "launch the crawler.")
	pflag.BoolVarP(&isNasha, "replicate", "r", false, "copy image to nasha")
	pflag.BoolVarP(&isDataset, "dataset", "d", false, "launch the crawler.")
	pflag.BoolVarP(&isClean, "clean", "", false, "auto-clean temporary files.")
	pflag.BoolVarP(&isAdmin, "admin", "", false, "launch the admin interface.")
	pflag.BoolVarP(&isTruncate, "truncate", "t", false, "truncate table content.")
	pflag.BoolVarP(&isExtract, "extract", "e", false, "extract data from urls.")
	pflag.BoolVarP(&isTor, "tor", "", false, "Proxy any GET requests with tor.")
	pflag.StringVarP(&torAddress, "tor-address", "", "sock5://localhost:5566", "Proxy addess with tor")
	pflag.StringVarP(&torAddress, "tor-privoxy", "", "http://localhost:8119", "Proxy address with tor-privoxy.")
	pflag.StringVarP(&analyzerURL, "analyzer-url", "", "http://localhost:9003", "Darknet Yolo model proxy web service address")
	pflag.StringVarP(&cacheDir, "cache-dir", "", "./shared/data", "cache directory.")
	pflag.BoolVarP(&isNoCache, "no-cache", "", false, "disable crawler cache.")
	pflag.BoolVarP(&isVerbose, "verbose", "v", false, "verbose mode.")
	pflag.BoolVarP(&isHelp, "help", "h", false, "help info.")
	pflag.Parse()
	if isHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// Instanciate geoip2 database
	geo = must(geoip2.Open(geoIpFile)).(*geoip2.Reader)

	dsl := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4,utf8&parseTime=True", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE"))
	fmt.Println("Dsl:", dsl)

	// Instanciate the mysql client
	DB, err := gorm.Open("mysql", dsl)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// ref. https://gorm.io/docs/generic_interface.html
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	DB.DB().SetMaxIdleConns(64)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	DB.DB().SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	DB.DB().SetConnMaxLifetime(time.Minute)

	// callback for images and validation
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)

	// truncate table
	if isTruncate {
		if err := DB.DropTableIfExists(&models.Vehicle{}).Error; err != nil {
			panic(err)
		}
		if err := DB.DropTableIfExists(&models.VehicleImage{}).Error; err != nil {
			panic(err)
		}
	}

	// migrate tables
	DB.AutoMigrate(&models.Vehicle{})
	DB.AutoMigrate(&models.VehicleImage{})
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
				fmt.Println("usePlugin", u, "currentPlugin", p)
				if strings.HasPrefix(p, "release/cars-dataset-"+u+".so") {
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
			c.IsClean = isClean
			c.AnalyzerURL = analyzerURL
			c.ConsumerThreads = 6
			pp.Println(c)
			c.DB = DB
			err := cmd.Crawl(c)
			if err != nil {
				log.Fatal(err)
			}
		}
		os.Exit(1)
	}

	// import catalog
	if isCatalog {
		for _, cmd := range ptPlugins.Commands {
			c := cmd.Config()
			c.DB = DB
			c.DryMode = isDryMode
			c.IsDebug = isVerbose
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

		padmin.SetupDashboard(DB, Admin)

		Admin.AddMenu(&admin.Menu{Name: "Crawl Management", Priority: 1})

		// Add media library
		Admin.AddResource(&media_library.MediaLibrary{}, &admin.Config{Menu: []string{"Crawl Management"}, Priority: -1})

		// Add VehicleImage as Media Librairy
		VehicleImagesResource := Admin.AddResource(&models.VehicleImage{}, &admin.Config{Menu: []string{"Crawl Management"}, Priority: -1})

		VehicleImagesResource.Filter(&admin.Filter{
			Name:       "SelectedType",
			Label:      "Media Type",
			Operations: []string{"contains"},
			Config:     &admin.SelectOneConfig{Collection: [][]string{{"video", "Video"}, {"image", "Image"}, {"file", "File"}, {"video_link", "Video Link"}}},
		})
		VehicleImagesResource.IndexAttrs("File", "Title", "BBox")

		VehicleImagesResource.UseTheme("grid")

		cars := Admin.AddResource(&models.Vehicle{}, &admin.Config{Menu: []string{"Crawl Management"}})
		cars.IndexAttrs("ID", "Name", "Modl", "Engine", "Year", "Source", "Manufacturer", "MainImage", "Images")

		cars.Meta(&admin.Meta{Name: "MainImage", Config: &media_library.MediaBoxConfig{
			RemoteDataResource: VehicleImagesResource,
			Max:                1,
			//Sizes: map[string]*media.Size{
			//	"main": {Width: 560, Height: 700},
			//},
		}})
		cars.Meta(&admin.Meta{Name: "MainImageURL", Valuer: func(record interface{}, context *qor.Context) interface{} {
			if p, ok := record.(*models.Vehicle); ok {
				result := bytes.NewBufferString("")
				tmpl, _ := template.New("").Parse("<img src='{{.image}}'></img>")
				tmpl.Execute(result, map[string]string{"image": p.MainImageURL()})
				return template.HTML(result.String())
			}
			return ""
		}})

		//cars.Filter(&admin.Filter{
		//	Name:   "Collections",
		//	Config: &admin.SelectOneConfig{RemoteDataResource: collection},
		//})

		cars.Filter(&admin.Filter{
			Name: "Manufacturer",
			Type: "string",
		})

		cars.Filter(&admin.Filter{
			Name: "Modl",
		})

		cars.Filter(&admin.Filter{
			Name: "Year",
			// Type: "number",
		})

		cars.Filter(&admin.Filter{
			Name: "CreatedAt",
		})

		// initalize an HTTP request multiplexer
		mux := http.NewServeMux()

		// Mount admin interface to mux
		Admin.MountTo("/admin", mux)

		router := gin.Default()

		// router.Use(realip.RealIP())
		// globally use middlewares
		router.Use(
			realip.RealIP(),
			middlewares.RecoveryWithWriter(os.Stderr),
			middlewares.Logger(geo),
			middlewares.CORS(),
			gin.ErrorLogger(),
		)

		// add basic auth
		admin := router.Group("/admin", gin.BasicAuth(gin.Accounts{"cars": "cars"}))
		{
			admin.Any("/*resources", gin.WrapH(mux))
		}

		router.Static("/system", "./public/system")
		router.Static("/public", "./public")

		fmt.Println("Listening on: 9008")
		s := &http.Server{
			Addr:           ":9008",
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()

	}

	if isNasha {

		// Scan
		type cnt struct {
			Count int
		}

		type res struct {
			Images string
		}

		type entryProperty struct {
			ID          int
			Url         string
			VideoLink   string
			FileName    string
			Description string
		}

		var count cnt
		DB.Raw("select count(images) as count FROM vehicles").Scan(&count)

		// instanciate throttler
		t := throttler.New(42, count.Count)

		counter := 0
		imgCounter := 0

		var results []res
		DB.Raw("select images FROM vehicles").Scan(&results)
		for _, result := range results {

			go func(r res) error {
				defer t.Done(nil)

				if r.Images == "" {
					return nil
				}

				var ep []entryProperty
				if err := json.Unmarshal([]byte(r.Images), &ep); err != nil {
					log.Fatalln("unmarshal error, ", err)
				}

				prefixPath := filepath.Join("/mnt/nasha/lucmichalski/cars-dataset/")
				os.MkdirAll(prefixPath, 0755)
				pp.Println("prefixPath:", prefixPath)

				for _, entry := range ep {
					sourceFile := filepath.Join("./", "public", entry.Url)
					input, err := ioutil.ReadFile(sourceFile)
					if err != nil {
						log.Warnln("reading file error, ", err)
						continue
					}
					destinationFile := filepath.Join(prefixPath, "public", entry.Url)
					err = ioutil.WriteFile(destinationFile, input, 0644)
					if err != nil {
						log.Fatalln("creating file error, ", err)
					}
					imgCounter++
				}

				percent := ((counter * 100) / count.Count)
				fmt.Printf("REF COUNTER=%v/%v (%v percent), IMG COUNTER=%v\n", counter, count.Count, percent, imgCounter)
				counter++
				return nil
			}(result)
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

		os.Exit(0)

	}

	if isDataset {

		// 322525 class
		// 2217792 images

		csvDataset, err := ccsv.NewCsvWriter("/mnt/nasha/lucmichalski/cars-dataset/datasets/cars_dataset.txt")
		if err != nil {
			panic("Could not open `dataset.txt` for writing")
		}

		// Flush pending writes and close file upon exit of Sitemap()
		defer csvDataset.Close()

		csvDataset.Write([]string{"name", "make", "model", "year", "image_path", "maxX", "maxY", "minX", "minY"})
		csvDataset.Flush()

		// Scan
		type cnt struct {
			Count int
		}

		type res struct {
			Name   string
			Make   string
			Modl   string
			Year   string
			Images string
		}

		type entryProperty struct {
			ID          int
			Url         string
			VideoLink   string
			FileName    string
			Description string
		}

		var count cnt
		DB.Raw("select count(id) as count FROM vehicles WHERE class='car' AND modl!='' AND year!='' AND manufacturer NOT IN ('Harley-Davidson', 'KTM', 'Triumph')").Scan(&count)

		// instanciate throttler
		t := throttler.New(42, count.Count)

		counter := 0
		imgCounter := 0

		var results []res
		DB.Raw("select name, manufacturer as make, modl, year, images FROM vehicles WHERE class='car' AND modl!='' AND year!='' AND manufacturer NOT IN ('Harley-Davidson', 'KTM', 'Triumph')").Scan(&results)
		for _, result := range results {

			go func(r res) error {
				defer t.Done(nil)

				if r.Images == "" {
					return nil
				}

				var ep []entryProperty
				// fmt.Println(result.Images)
				if err := json.Unmarshal([]byte(r.Images), &ep); err != nil {
					log.Fatalln("unmarshal error, ", err)
				}

				//if len(ep) < 2 {
				//	return nil
				//}

				manufacturer := strings.Replace(strings.ToUpper(r.Make), " ", "-", -1)
				switch manufacturer {
				case "CITROËN":
					manufacturer = "CITROEN"
				case "DS":
					manufacturer = "DS-AUTOMOBILES"
				case "LANDROVER":
					manufacturer = "LAND-ROVER"
				case "MERCEDES":
					manufacturer = "MERCEDES-BENZ"
				case "ROLLSROYCE":
					manufacturer = "ROLLS-ROYCE"
				}

				// prefixPath := filepath.Join("./", "datasets", "cars", result.Name)
				prefixPath := filepath.Join("/mnt/nasha/lucmichalski/cars-dataset/", "datasets", "cars", manufacturer, strings.ToUpper(r.Modl), r.Year)
				os.MkdirAll(prefixPath, 0755)
				// pp.Println("prefixPath:", prefixPath)

				for _, entry := range ep {

					if entry.ID < 250000 {
						continue
					}

					// get image Info (to test)
					var vi models.VehicleImage
					err := DB.First(&vi, entry.ID).Error
					if err != nil {
						log.Warnln("VehicleImage", err)
						continue
					}
					// fmt.Println("image checksum:", vi.Checksum, ", bounding_box:", vi.BBox)

					sourceFile := filepath.Join("./", "public", entry.Url)
					// pp.Println("sourceFile:", sourceFile)

					input, err := ioutil.ReadFile(sourceFile)
					if err != nil {
						log.Warnln("reading file error, ", err)
						continue
					}

					destinationFile := filepath.Join(prefixPath, vi.Checksum+filepath.Ext(entry.Url))
					// destinationFile := filepath.Join(prefixPath, strconv.Itoa(entry.ID)+"-"+filepath.Base(entry.Url))
					err = ioutil.WriteFile(destinationFile, input, 0644)
					if err != nil {
						// return err
						log.Fatalln("creating file error, ", err)
					}
					// pp.Println("destinationFile:", destinationFile)

					bbox := strings.Split(vi.BBox, ",")
					csvDataset.Write([]string{r.Name, strings.Replace(strings.ToUpper(r.Make), " ", "-", -1), strings.ToUpper(r.Modl), r.Year, destinationFile, bbox[0], bbox[1], bbox[2], bbox[3]})
					csvDataset.Flush()

					imgCounter++
				}

				percent := ((counter * 100) / count.Count)
				fmt.Printf("REF COUNTER=%v/%v (%v percent), IMG COUNTER=%v\n", counter, count.Count, percent, imgCounter)
				counter++

				return nil

			}(result)

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

		os.Exit(0)

	}

}

// fail fast on initialization
func must(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}
