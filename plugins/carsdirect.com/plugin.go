package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/carsdirect.com/admin"
	"github.com/lucmichalski/cars-contrib/carsdirect.com/crawler"
	"github.com/lucmichalski/cars-contrib/carsdirect.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingCarsDirect{},
}

var Resources = []interface{}{
	&models.SettingCarsDirect{},
}

type carsDirectPlugin string

func (o carsDirectPlugin) Name() string      { return string(o) }
func (o carsDirectPlugin) Section() string   { return `carsdirect.com` }
func (o carsDirectPlugin) Usage() string     { return `hello` }
func (o carsDirectPlugin) ShortDesc() string { return `carsdirect.com crawler"` }
func (o carsDirectPlugin) LongDesc() string  { return o.ShortDesc() }

func (o carsDirectPlugin) Migrate() []interface{} {
	return Tables
}

func (o carsDirectPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o carsDirectPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o carsDirectPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o carsDirectPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.carsdirect.com", "carsdirect.com"},
		URLs: []string{
			"https://www.carsdirect.com/sitemap.xml",
			"https://www.carsdirect.com/sitemaps/production.carsdirect.com.sitemap.1.xml",
			"https://www.carsdirect.com/sitemaps/production.carsdirect.com.sitemap.2.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type carsDirectCommands struct{}

func (t *carsDirectCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
----------------------------------------------------------------------------------------------------------------------------------------
:'######:::::'###::::'########:::'######::'########::'####:'########::'########::'######::'########:::::::'######:::'#######::'##::::'##:
'##... ##:::'## ##::: ##.... ##:'##... ##: ##.... ##:. ##:: ##.... ##: ##.....::'##... ##:... ##..:::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:::'##:. ##:: ##:::: ##: ##:::..:: ##:::: ##:: ##:: ##:::: ##: ##::::::: ##:::..::::: ##::::::::: ##:::..:: ##:::: ##: ####'####:
 ##:::::::'##:::. ##: ########::. ######:: ##:::: ##:: ##:: ########:: ######::: ##:::::::::: ##::::::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: #########: ##.. ##::::..... ##: ##:::: ##:: ##:: ##.. ##::: ##...:::: ##:::::::::: ##::::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##.... ##: ##::. ##::'##::: ##: ##:::: ##:: ##:: ##::. ##:: ##::::::: ##::: ##:::: ##::::'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ##:::: ##: ##:::. ##:. ######:: ########::'####: ##:::. ##: ########:. ######::::: ##:::: ###:. ######::. #######:: ##:::: ##:
:......:::..:::::..::..:::::..:::......:::........:::....::..:::::..::........:::......::::::..:::::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *carsDirectCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"carsdirect": carsDirectPlugin("carsdirect"), //OP
	}
}

var Plugins carsDirectCommands
