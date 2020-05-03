package main

import (
	"context"
	"errors"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/thecarconnection.com/admin"
	"github.com/lucmichalski/cars-contrib/thecarconnection.com/crawler"
	"github.com/lucmichalski/cars-contrib/thecarconnection.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingTheCarConnection{},
}

var Resources = []interface{}{
	&models.SettingTheCarConnection{},
}

type theCarConnectionPlugin string

func (o theCarConnectionPlugin) Name() string      { return string(o) }
func (o theCarConnectionPlugin) Section() string   { return `thecarconnection.com` }
func (o theCarConnectionPlugin) Usage() string     { return `hello` }
func (o theCarConnectionPlugin) ShortDesc() string { return `thecarconnection.com crawler"` }
func (o theCarConnectionPlugin) LongDesc() string  { return o.ShortDesc() }

func (o theCarConnectionPlugin) Migrate() []interface{} {
	return Tables
}

func (o theCarConnectionPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o theCarConnectionPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o theCarConnectionPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o theCarConnectionPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.thecarconnection.com", "thecarconnection.com"},
		URLs: []string{
			"https://www.thecarconnection.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
		AnalyzerURL:     "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type theCarConnectionCommands struct{}

func (t *theCarConnectionCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
'########:'##::::'##:'########::'######:::::'###::::'########:::'######:::'#######::'##::: ##:'##::: ##:'########::'######::'########:'####::'#######::'##::: ##:::::::'######:::'#######::'##::::'##:
... ##..:: ##:::: ##: ##.....::'##... ##:::'## ##::: ##.... ##:'##... ##:'##.... ##: ###:: ##: ###:: ##: ##.....::'##... ##:... ##..::. ##::'##.... ##: ###:: ##::::::'##... ##:'##.... ##: ###::'###:
::: ##:::: ##:::: ##: ##::::::: ##:::..:::'##:. ##:: ##:::: ##: ##:::..:: ##:::: ##: ####: ##: ####: ##: ##::::::: ##:::..::::: ##::::: ##:: ##:::: ##: ####: ##:::::: ##:::..:: ##:::: ##: ####'####:
::: ##:::: #########: ######::: ##:::::::'##:::. ##: ########:: ##::::::: ##:::: ##: ## ## ##: ## ## ##: ######::: ##:::::::::: ##::::: ##:: ##:::: ##: ## ## ##:::::: ##::::::: ##:::: ##: ## ### ##:
::: ##:::: ##.... ##: ##...:::: ##::::::: #########: ##.. ##::: ##::::::: ##:::: ##: ##. ####: ##. ####: ##...:::: ##:::::::::: ##::::: ##:: ##:::: ##: ##. ####:::::: ##::::::: ##:::: ##: ##. #: ##:
::: ##:::: ##:::: ##: ##::::::: ##::: ##: ##.... ##: ##::. ##:: ##::: ##: ##:::: ##: ##:. ###: ##:. ###: ##::::::: ##::: ##:::: ##::::: ##:: ##:::: ##: ##:. ###:'###: ##::: ##: ##:::: ##: ##:.:: ##:
::: ##:::: ##:::: ##: ########:. ######:: ##:::: ##: ##:::. ##:. ######::. #######:: ##::. ##: ##::. ##: ########:. ######::::: ##::::'####:. #######:: ##::. ##: ###:. ######::. #######:: ##:::: ##:
:::..:::::..:::::..::........:::......:::..:::::..::..:::::..:::......::::.......:::..::::..::..::::..::........:::......::::::..:::::....:::.......:::..::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *theCarConnectionCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"theCarConnection": theCarConnectionPlugin("theCarConnection"), //OP
	}
}

var Plugins theCarConnectionCommands
