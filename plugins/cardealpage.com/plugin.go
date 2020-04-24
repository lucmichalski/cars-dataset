package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/cardealpage.com/admin"
	"github.com/lucmichalski/cars-contrib/cardealpage.com/crawler"
	"github.com/lucmichalski/cars-contrib/cardealpage.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingCarDealPage{},
}

var Resources = []interface{}{
	&models.SettingCarDealPage{},
}

type carDealPagelugin string

func (o carDealPagelugin) Name() string      { return string(o) }
func (o carDealPagelugin) Section() string   { return `1001pneus.fr` }
func (o carDealPagelugin) Usage() string     { return `hello` }
func (o carDealPagelugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o carDealPagelugin) LongDesc() string  { return o.ShortDesc() }

func (o carDealPagelugin) Migrate() []interface{} {
	return Tables
}

func (o carDealPagelugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o carDealPagelugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o carDealPagelugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o carDealPagelugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.cardealpage.com", "cardealpage.com"},
		URLs: []string{
			"https://www.cardealpage.com",
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
		"cars": carDealPagelugin("cars"), //OP
	}
}

var Plugins carsCommands
