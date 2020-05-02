package main

import (
	"context"
	"errors"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/admin"
	"github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/crawler"
	"github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutoTraderMotorcycles{},
}

var Resources = []interface{}{
	&models.SettingAutoTraderMotorcycles{},
}

type autoTraderMotorcyclesPlugin string

func (o autoTraderMotorcyclesPlugin) Name() string      { return string(o) }
func (o autoTraderMotorcyclesPlugin) Section() string   { return `1001pneus.fr` }
func (o autoTraderMotorcyclesPlugin) Usage() string     { return `hello` }
func (o autoTraderMotorcyclesPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o autoTraderMotorcyclesPlugin) LongDesc() string  { return o.ShortDesc() }

func (o autoTraderMotorcyclesPlugin) Migrate() []interface{} {
	return Tables
}

func (o autoTraderMotorcyclesPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autoTraderMotorcyclesPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autoTraderMotorcyclesPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o autoTraderMotorcyclesPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"autotrader.com", "motorcycles.autotrader.com"},
		URLs: []string{
			"https://motorcycles.autotrader.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
		AnalyzerURL:     "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autoTraderMotorcyclesCommands struct{}

func (t *autoTraderMotorcyclesCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------------------------------------------------------------
'##::::'##::'#######::'########::'#######::'########:::'######::'##:::'##::'######::'##:::::::'########::'######::::::::::::::::::::::::::::::::
 ###::'###:'##.... ##:... ##..::'##.... ##: ##.... ##:'##... ##:. ##:'##::'##... ##: ##::::::: ##.....::'##... ##:::::::::::::::::::::::::::::::
 ####'####: ##:::: ##:::: ##:::: ##:::: ##: ##:::: ##: ##:::..:::. ####::: ##:::..:: ##::::::: ##::::::: ##:::..::::::::::::::::::::::::::::::::
 ## ### ##: ##:::: ##:::: ##:::: ##:::: ##: ########:: ##:::::::::. ##:::: ##::::::: ##::::::: ######:::. ######::::::::::::::::::::::::::::::::
 ##. #: ##: ##:::: ##:::: ##:::: ##:::: ##: ##.. ##::: ##:::::::::: ##:::: ##::::::: ##::::::: ##...:::::..... ##:::::::::::::::::::::::::::::::
 ##:.:: ##: ##:::: ##:::: ##:::: ##:::: ##: ##::. ##:: ##::: ##:::: ##:::: ##::: ##: ##::::::: ##:::::::'##::: ##:'###::::::::::::::::::::::::::
 ##:::: ##:. #######::::: ##::::. #######:: ##:::. ##:. ######::::: ##::::. ######:: ########: ########:. ######:: ###::::::::::::::::::::::::::
..:::::..:::.......::::::..::::::.......:::..:::::..:::......::::::..::::::......:::........::........:::......:::...:::::::::::::::::::::::::::
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

func (t *autoTraderMotorcyclesCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"autoTraderMotorcycles": autoTraderMotorcyclesPlugin("autoTraderMotorcycles"), //OP
	}
}

var Plugins autoTraderMotorcyclesCommands
