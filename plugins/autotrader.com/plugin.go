package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/autotrader.com/admin"
	"github.com/lucmichalski/cars-contrib/autotrader.com/crawler"
	"github.com/lucmichalski/cars-contrib/autotrader.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutoTraderCom{},
}

var Resources = []interface{}{
	&models.SettingAutoTraderCom{},
}

type autoTraderComPlugin string

func (o autoTraderComPlugin) Name() string      { return string(o) }
func (o autoTraderComPlugin) Section() string   { return `1001pneus.fr` }
func (o autoTraderComPlugin) Usage() string     { return `hello` }
func (o autoTraderComPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o autoTraderComPlugin) LongDesc() string  { return o.ShortDesc() }

func (o autoTraderComPlugin) Migrate() []interface{} {
	return Tables
}

func (o autoTraderComPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autoTraderComPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autoTraderComPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.autotrader.com", "autotrader.com"},
		URLs: []string{
			"https://www.autotrader.com/sitemap.xml",
			"https://www.autotrader.com/sitemap_certified_geo.xml.gz",
			"https://www.autotrader.com/sitemap_new_geo.xml.gz",
			"https://www.autotrader.com/sitemap_used_geo.xml.gz",
			"https://www.autotrader.com/sitemap_dlr.xml.gz",
			"https://www.autotrader.com/sitemap_mm.xml.gz",
			"https://www.autotrader.com/sitemap_mmt.xml.gz",
			"https://www.autotrader.com/sitemap_mmty.xml.gz",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autoTraderComCommands struct{}

func (t *autoTraderComCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------------------------------------------------------------
:::'###::::'##::::'##:'########::'#######::'########:'########:::::'###::::'########::'########:'########::::::::'######:::'#######::'##::::'##:
::'## ##::: ##:::: ##:... ##..::'##.... ##:... ##..:: ##.... ##:::'## ##::: ##.... ##: ##.....:: ##.... ##::::::'##... ##:'##.... ##: ###::'###:
:'##:. ##:: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##::'##:. ##:: ##:::: ##: ##::::::: ##:::: ##:::::: ##:::..:: ##:::: ##: ####'####:
'##:::. ##: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ########::'##:::. ##: ##:::: ##: ######::: ########::::::: ##::::::: ##:::: ##: ## ### ##:
 #########: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##.. ##::: #########: ##:::: ##: ##...:::: ##.. ##:::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##.... ##: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##::. ##:: ##.... ##: ##:::: ##: ##::::::: ##::. ##::'###: ##::: ##: ##:::: ##: ##:.:: ##:
 ##:::: ##:. #######::::: ##::::. #######::::: ##:::: ##:::. ##: ##:::: ##: ########:: ########: ##:::. ##: ###:. ######::. #######:: ##:::: ##:
..:::::..:::.......::::::..::::::.......::::::..:::::..:::::..::..:::::..::........:::........::..:::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *autoTraderComCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"1001pneus": autoTraderComPlugin("1001pneus"), //OP
	}
}

var Plugins autoTraderComCommands
