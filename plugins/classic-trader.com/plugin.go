package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/qor/admin"

	adm "github.com/lucmichalski/cars-contrib/classic-trader.com/admin"
	"github.com/lucmichalski/cars-contrib/classic-trader.com/crawler"
	"github.com/lucmichalski/cars-contrib/classic-trader.com/models"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingClassicTraderCom{},
}

var Resources = []interface{}{
	&models.SettingClassicTraderCom{},
}

type classicTraderPlugin string

func (o classicTraderPlugin) Name() string      { return string(o) }
func (o classicTraderPlugin) Section() string   { return `1001pneus.fr` }
func (o classicTraderPlugin) Usage() string     { return `hello` }
func (o classicTraderPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o classicTraderPlugin) LongDesc() string  { return o.ShortDesc() }

func (o classicTraderPlugin) Migrate() []interface{} {
	return Tables
}

func (o classicTraderPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o classicTraderPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o classicTraderPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o classicTraderPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.classic-trader.com", "classic-trader.com"},
		URLs: []string{
			"https://cdn.classic-trader.com/I/sitemap/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
		AnalyzerURL:     "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type classicTraderCommands struct{}

func (t *classicTraderCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
:'######::'##::::::::::'###:::::'######:::'######::'####::'######:::::::::::'########:'########:::::'###::::'########::'########:'########::::::::'######:::'#######::'##::::'##:
'##... ##: ##:::::::::'## ##:::'##... ##:'##... ##:. ##::'##... ##::::::::::... ##..:: ##.... ##:::'## ##::: ##.... ##: ##.....:: ##.... ##::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:: ##::::::::'##:. ##:: ##:::..:: ##:::..::: ##:: ##:::..:::::::::::::: ##:::: ##:::: ##::'##:. ##:: ##:::: ##: ##::::::: ##:::: ##:::::: ##:::..:: ##:::: ##: ####'####:
 ##::::::: ##:::::::'##:::. ##:. ######::. ######::: ##:: ##:::::::'#######:::: ##:::: ########::'##:::. ##: ##:::: ##: ######::: ########::::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: ##::::::: #########::..... ##::..... ##:: ##:: ##:::::::........:::: ##:::: ##.. ##::: #########: ##:::: ##: ##...:::: ##.. ##:::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##::::::: ##.... ##:'##::: ##:'##::: ##:: ##:: ##::: ##::::::::::::: ##:::: ##::. ##:: ##.... ##: ##:::: ##: ##::::::: ##::. ##::'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ########: ##:::: ##:. ######::. ######::'####:. ######:::::::::::::: ##:::: ##:::. ##: ##:::: ##: ########:: ########: ##:::. ##: ###:. ######::. #######:: ##:::: ##:
:......:::........::..:::::..:::......::::......:::....:::......:::::::::::::::..:::::..:::::..::..:::::..::........:::........::..:::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *classicTraderCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"classicTrader": classicTraderPlugin("classicTrader"), //OP
	}
}

var Plugins classicTraderCommands
