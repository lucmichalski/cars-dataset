package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/turo.com/admin"
	"github.com/lucmichalski/cars-contrib/turo.com/crawler"
	"github.com/lucmichalski/cars-contrib/turo.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingTuro{},
}

var Resources = []interface{}{
	&models.SettingTuro{},
}

type turoPlugin string

func (o turoPlugin) Name() string      { return string(o) }
func (o turoPlugin) Section() string   { return `turo.com` }
func (o turoPlugin) Usage() string     { return `hello` }
func (o turoPlugin) ShortDesc() string { return `turo.com crawler"` }
func (o turoPlugin) LongDesc() string  { return o.ShortDesc() }

func (o turoPlugin) Migrate() []interface{} {
	return Tables
}

func (o turoPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o turoPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o turoPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o turoPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.turo.com", "turo.com"},
		URLs: []string{
			"https://www.turo.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: false,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type turoCommands struct{}

func (t *turoCommands) Init(ctx context.Context) error {
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

func (t *turoCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"turo": turoPlugin("turo"), //OP
	}
}

var Plugins turoCommands
