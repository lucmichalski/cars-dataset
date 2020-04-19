package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/leboncoin.fr/admin"
	"github.com/lucmichalski/cars-contrib/leboncoin.fr/crawler"
	"github.com/lucmichalski/cars-contrib/leboncoin.fr/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingLeBonCoin{},
}

var Resources = []interface{}{
	&models.SettingLeBonCoin{},
}

type leBonCoinPlugin string

func (o leBonCoinPlugin) Name() string      { return string(o) }
func (o leBonCoinPlugin) Section() string   { return `leboncoin.fr` }
func (o leBonCoinPlugin) Usage() string     { return `hello` }
func (o leBonCoinPlugin) ShortDesc() string { return `leboncoin.fr crawler"` }
func (o leBonCoinPlugin) LongDesc() string  { return o.ShortDesc() }

func (o leBonCoinPlugin) Migrate() []interface{} {
	return Tables
}

func (o leBonCoinPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o leBonCoinPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o leBonCoinPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o leBonCoinPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.leboncoin.fr", "leboncoin.fr"},
		URLs: []string{
			"https://www.leboncoin.fr/sitemap_index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type leBonCoinCommands struct{}

func (t *leBonCoinCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
------------------------------------------------------------------------------------------------------------------
'##:::::::'########:'########:::'#######::'##::: ##::'######:::'#######::'####:'##::: ##::::::'########:'########::
 ##::::::: ##.....:: ##.... ##:'##.... ##: ###:: ##:'##... ##:'##.... ##:. ##:: ###:: ##:::::: ##.....:: ##.... ##:
 ##::::::: ##::::::: ##:::: ##: ##:::: ##: ####: ##: ##:::..:: ##:::: ##:: ##:: ####: ##:::::: ##::::::: ##:::: ##:
 ##::::::: ######::: ########:: ##:::: ##: ## ## ##: ##::::::: ##:::: ##:: ##:: ## ## ##:::::: ######::: ########::
 ##::::::: ##...:::: ##.... ##: ##:::: ##: ##. ####: ##::::::: ##:::: ##:: ##:: ##. ####:::::: ##...:::: ##.. ##:::
 ##::::::: ##::::::: ##:::: ##: ##:::: ##: ##:. ###: ##::: ##: ##:::: ##:: ##:: ##:. ###:'###: ##::::::: ##::. ##::
 ########: ########: ########::. #######:: ##::. ##:. ######::. #######::'####: ##::. ##: ###: ##::::::: ##:::. ##:
........::........::........::::.......:::..::::..:::......::::.......:::....::..::::..::...::..::::::::..:::::..::
`)

	return nil
}

func (t *leBonCoinCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"leBonCoin": leBonCoinPlugin("leBonCoin"), //OP
	}
}

var Plugins leBonCoinCommands
