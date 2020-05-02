package main

import (
	"context"
	"errors"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/classics.autotrader.com/admin"
	"github.com/lucmichalski/cars-contrib/classics.autotrader.com/crawler"
	"github.com/lucmichalski/cars-contrib/classics.autotrader.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingClassicAutoTrader{},
}

var Resources = []interface{}{
	&models.SettingClassicAutoTrader{},
}

type classicClassicAutoTraderPlugin string

func (o classicClassicAutoTraderPlugin) Name() string      { return string(o) }
func (o classicClassicAutoTraderPlugin) Section() string   { return `1001pneus.fr` }
func (o classicClassicAutoTraderPlugin) Usage() string     { return `hello` }
func (o classicClassicAutoTraderPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o classicClassicAutoTraderPlugin) LongDesc() string  { return o.ShortDesc() }

func (o classicClassicAutoTraderPlugin) Migrate() []interface{} {
	return Tables
}

func (o classicClassicAutoTraderPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o classicClassicAutoTraderPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o classicClassicAutoTraderPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o classicClassicAutoTraderPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.classics.autotrader.com", "classics.autotrader.com", "classics.classics.autotrader.com"},
		URLs: []string{
			"https://classics.autotrader.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
		AnalyzerURL:     "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type classicClassicAutoTraderCommands struct{}

func (t *classicClassicAutoTraderCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------------------------------------------------------------
:'######::'##::::::::::'###:::::'######:::'######::'####::'######:::'######:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
'##... ##: ##:::::::::'## ##:::'##... ##:'##... ##:. ##::'##... ##:'##... ##::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
 ##:::..:: ##::::::::'##:. ##:: ##:::..:: ##:::..::: ##:: ##:::..:: ##:::..:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
 ##::::::: ##:::::::'##:::. ##:. ######::. ######::: ##:: ##:::::::. ######:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
 ##::::::: ##::::::: #########::..... ##::..... ##:: ##:: ##::::::::..... ##::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
 ##::: ##: ##::::::: ##.... ##:'##::: ##:'##::: ##:: ##:: ##::: ##:'##::: ##:'###:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
. ######:: ########: ##:::: ##:. ######::. ######::'####:. ######::. ######:: ###:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
:......:::........::..:::::..:::......::::......:::....:::......::::......:::...::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
------------------------------------------------------------------------------------------------------------------------------------------------
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

func (t *classicClassicAutoTraderCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"classicClassicAutoTrader": classicClassicAutoTraderPlugin("classicClassicAutoTrader"), //OP
	}
}

var Plugins classicClassicAutoTraderCommands
