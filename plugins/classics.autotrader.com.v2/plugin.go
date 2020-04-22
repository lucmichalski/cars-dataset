package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/classics.autotrader.com.v2/admin"
	"github.com/lucmichalski/cars-contrib/classics.autotrader.com.v2/crawler"
	"github.com/lucmichalski/cars-contrib/classics.autotrader.com.v2/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingClassicAutoTraderV2{},
}

var Resources = []interface{}{
	&models.SettingClassicAutoTraderV2{},
}

type classicAutoTraderV2Plugin string

func (o classicAutoTraderV2Plugin) Name() string      { return string(o) }
func (o classicAutoTraderV2Plugin) Section() string   { return `classics.autotrader.com.v2` }
func (o classicAutoTraderV2Plugin) Usage() string     { return `hello` }
func (o classicAutoTraderV2Plugin) ShortDesc() string { return `classics.autotrader.com.v2 crawler"` }
func (o classicAutoTraderV2Plugin) LongDesc() string  { return o.ShortDesc() }

func (o classicAutoTraderV2Plugin) Migrate() []interface{} {
	return Tables
}

func (o classicAutoTraderV2Plugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o classicAutoTraderV2Plugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o classicAutoTraderV2Plugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o classicAutoTraderV2Plugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"classics.autotrader.com"},
		URLs: []string{
			"https://classics.autotrader.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type classicAutoTraderV2Commands struct{}

func (t *classicAutoTraderV2Commands) Init(ctx context.Context) error {
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

func (t *classicAutoTraderV2Commands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"classicAutoTraderV2": classicAutoTraderV2Plugin("classicAutoTraderV2"), //OP
	}
}

var Plugins classicAutoTraderV2Commands
