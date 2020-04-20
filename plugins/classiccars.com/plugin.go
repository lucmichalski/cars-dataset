package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/classiccars.com/admin"
	"github.com/lucmichalski/cars-contrib/classiccars.com/crawler"
	"github.com/lucmichalski/cars-contrib/classiccars.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingClassicCars{},
}

var Resources = []interface{}{
	&models.SettingClassicCars{},
}

type classicCarsPlugin string

func (o classicCarsPlugin) Name() string      { return string(o) }
func (o classicCarsPlugin) Section() string   { return `classiccars.com` }
func (o classicCarsPlugin) Usage() string     { return `hello` }
func (o classicCarsPlugin) ShortDesc() string { return `classiccars.com crawler"` }
func (o classicCarsPlugin) LongDesc() string  { return o.ShortDesc() }

func (o classicCarsPlugin) Migrate() []interface{} {
	return Tables
}

func (o classicCarsPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o classicCarsPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o classicCarsPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o classicCarsPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.classiccars.com", "classiccars.com"},
		URLs: []string{
			"https://classiccars.com/sitemap_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type classicCarsCommands struct{}

func (t *classicCarsCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-------------------------------------------------------------------------------------------------------------------------------------------------
:'######::'##::::::::::'###:::::'######:::'######::'####::'######:::'######:::::'###::::'########:::'######::::::::'######:::'#######::'##::::'##:
'##... ##: ##:::::::::'## ##:::'##... ##:'##... ##:. ##::'##... ##:'##... ##:::'## ##::: ##.... ##:'##... ##::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:: ##::::::::'##:. ##:: ##:::..:: ##:::..::: ##:: ##:::..:: ##:::..:::'##:. ##:: ##:::: ##: ##:::..::::::: ##:::..:: ##:::: ##: ####'####:
 ##::::::: ##:::::::'##:::. ##:. ######::. ######::: ##:: ##::::::: ##:::::::'##:::. ##: ########::. ######::::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: ##::::::: #########::..... ##::..... ##:: ##:: ##::::::: ##::::::: #########: ##.. ##::::..... ##:::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##::::::: ##.... ##:'##::: ##:'##::: ##:: ##:: ##::: ##: ##::: ##: ##.... ##: ##::. ##::'##::: ##:'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ########: ##:::: ##:. ######::. ######::'####:. ######::. ######:: ##:::: ##: ##:::. ##:. ######:: ###:. ######::. #######:: ##:::: ##:
:......:::........::..:::::..:::......::::......:::....:::......::::......:::..:::::..::..:::::..:::......:::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *classicCarsCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"classicCars": classicCarsPlugin("classicCars"), //OP
	}
}

var Plugins classicCarsCommands
