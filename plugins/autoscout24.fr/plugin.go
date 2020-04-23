package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/autoscout24.fr/admin"
	"github.com/lucmichalski/cars-contrib/autoscout24.fr/crawler"
	"github.com/lucmichalski/cars-contrib/autoscout24.fr/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutoScout24Fr{},
}

var Resources = []interface{}{
	&models.SettingAutoScout24Fr{},
}

type autoScout24FrPlugin string

func (o autoScout24FrPlugin) Name() string      { return string(o) }
func (o autoScout24FrPlugin) Section() string   { return `autoscout24.fr` }
func (o autoScout24FrPlugin) Usage() string     { return `hello` }
func (o autoScout24FrPlugin) ShortDesc() string { return `autoscout24.fr crawler"` }
func (o autoScout24FrPlugin) LongDesc() string  { return o.ShortDesc() }

func (o autoScout24FrPlugin) Migrate() []interface{} {
	return Tables
}

func (o autoScout24FrPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autoScout24FrPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autoScout24FrPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o autoScout24FrPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.autoscout24.fr", "autoscout24.fr"},
		URLs: []string{
			"https://www.autoscout24.fr/lst?sort=price&desc=0&ustate=N%2CU&atype=C",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 6,
		IsSitemapIndex: false,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autoScout24FrCommands struct{}

func (t *autoScout24FrCommands) Init(ctx context.Context) error {
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

func (t *autoScout24FrCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"autoscout24": autoScout24FrPlugin("autoscout24"), //OP
	}
}

var Plugins autoScout24FrCommands
