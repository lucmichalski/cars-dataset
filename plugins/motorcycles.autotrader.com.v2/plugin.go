package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/admin"
	"github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/crawler"
	"github.com/lucmichalski/cars-contrib/motorcycles.autotrader.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutoTraderMotorcyclesV2{},
}

var Resources = []interface{}{
	&models.SettingAutoTraderMotorcyclesV2{},
}

type autoTraderMotorcyclesV2Plugin string

func (o autoTraderMotorcyclesV2Plugin) Name() string      { return string(o) }
func (o autoTraderMotorcyclesV2Plugin) Section() string   { return `1001pneus.fr` }
func (o autoTraderMotorcyclesV2Plugin) Usage() string     { return `hello` }
func (o autoTraderMotorcyclesV2Plugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o autoTraderMotorcyclesV2Plugin) LongDesc() string  { return o.ShortDesc() }

func (o autoTraderMotorcyclesV2Plugin) Migrate() []interface{} {
	return Tables
}

func (o autoTraderMotorcyclesV2Plugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autoTraderMotorcyclesV2Plugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autoTraderMotorcyclesV2Plugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o autoTraderMotorcyclesV2Plugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"autotrader.com", "motorcycles.autotrader.com"},
		URLs: []string{
			"https://motorcycles.autotrader.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autoTraderMotorcyclesV2Commands struct{}

func (t *autoTraderMotorcyclesV2Commands) Init(ctx context.Context) error {
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

func (t *autoTraderMotorcyclesV2Commands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"autoTraderMotorcyclesV2": autoTraderMotorcyclesV2Plugin("autoTraderMotorcyclesV2"), //OP
	}
}

var Plugins autoTraderMotorcyclesV2Commands
