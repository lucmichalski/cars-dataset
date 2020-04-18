package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/cars.com/admin"
	"github.com/lucmichalski/cars-contrib/cars.com/crawler"
	"github.com/lucmichalski/cars-contrib/cars.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingCarsCom{},
}

var Resources = []interface{}{
	&models.SettingCarsCom{},
}

type carsPlugin string

func (o carsPlugin) Name() string      { return string(o) }
func (o carsPlugin) Section() string   { return `1001pneus.fr` }
func (o carsPlugin) Usage() string     { return `hello` }
func (o carsPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o carsPlugin) LongDesc() string  { return o.ShortDesc() }

func (o carsPlugin) Migrate() []interface{} {
	return Tables
}

func (o carsPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o carsPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o carsPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o carsPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.cars.com", "cars.com"},
		URLs: []string{
			"https://www.cars.com/secure/sitemap/s_sitemap_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type carsCommands struct{}

func (t *carsCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
------------------------------------------------------------------------------
:'######:::::'###::::'########:::'######::::::::'######:::'#######::'##::::'##:
'##... ##:::'## ##::: ##.... ##:'##... ##::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:::'##:. ##:: ##:::: ##: ##:::..::::::: ##:::..:: ##:::: ##: ####'####:
 ##:::::::'##:::. ##: ########::. ######::::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: #########: ##.. ##::::..... ##:::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##.... ##: ##::. ##::'##::: ##:'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ##:::: ##: ##:::. ##:. ######:: ###:. ######::. #######:: ##:::: ##:
:......:::..:::::..::..:::::..:::......:::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *carsCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"cars": carsPlugin("cars"), //OP
	}
}

var Plugins carsCommands
