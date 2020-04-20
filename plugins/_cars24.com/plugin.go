package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/cars24.com/admin"
	"github.com/lucmichalski/cars-contrib/cars24.com/crawler"
	"github.com/lucmichalski/cars-contrib/cars24.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingCars24{},
}

var Resources = []interface{}{
	&models.SettingCars24{},
}

type cars24Plugin string

func (o cars24Plugin) Name() string      { return string(o) }
func (o cars24Plugin) Section() string   { return `cars24.com` }
func (o cars24Plugin) Usage() string     { return `hello` }
func (o cars24Plugin) ShortDesc() string { return `cars24.com crawler"` }
func (o cars24Plugin) LongDesc() string  { return o.ShortDesc() }

func (o cars24Plugin) Migrate() []interface{} {
	return Tables
}

func (o cars24Plugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o cars24Plugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o cars24Plugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o cars24Plugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.cars24.com", "cars24.com"},
		URLs: []string{
			"https://www.cars24.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type cars24Commands struct{}

func (t *cars24Commands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
----------------------------------------------------------------------------------------------------
:'######:::::'###::::'########:::'######:::'#######::'##::::::::::::::'######:::'#######::'##::::'##:
'##... ##:::'## ##::: ##.... ##:'##... ##:'##.... ##: ##:::'##:::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:::'##:. ##:: ##:::: ##: ##:::..::..::::: ##: ##::: ##::::::: ##:::..:: ##:::: ##: ####'####:
 ##:::::::'##:::. ##: ########::. ######:::'#######:: ##::: ##::::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: #########: ##.. ##::::..... ##:'##:::::::: #########:::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##.... ##: ##::. ##::'##::: ##: ##::::::::...... ##::'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ##:::: ##: ##:::. ##:. ######:: #########::::::: ##:: ###:. ######::. #######:: ##:::: ##:
:......:::..:::::..::..:::::..:::......:::.........::::::::..:::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *cars24Commands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"cars24": cars24Plugin("cars24"), //OP
	}
}

var Plugins cars24Commands
