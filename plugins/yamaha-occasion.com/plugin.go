package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/yamaha-occasion.com/admin"
	"github.com/lucmichalski/cars-contrib/yamaha-occasion.com/crawler"
	"github.com/lucmichalski/cars-contrib/yamaha-occasion.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingYamahaOccasion{},
}

var Resources = []interface{}{
	&models.SettingYamahaOccasion{},
}

type yamahaOccasionPlugin string

func (o yamahaOccasionPlugin) Name() string      { return string(o) }
func (o yamahaOccasionPlugin) Section() string   { return `yamaha-occasion.com` }
func (o yamahaOccasionPlugin) Usage() string     { return `hello` }
func (o yamahaOccasionPlugin) ShortDesc() string { return `yamaha-occasion.com crawler"` }
func (o yamahaOccasionPlugin) LongDesc() string  { return o.ShortDesc() }

func (o yamahaOccasionPlugin) Migrate() []interface{} {
	return Tables
}

func (o yamahaOccasionPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o yamahaOccasionPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o yamahaOccasionPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o yamahaOccasionPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.yamaha-occasion.com", "yamaha-occasion.com"},
		URLs: []string{
			"https://yamaha-occasion.com/",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type yamahaOccasionCommands struct{}

func (t *yamahaOccasionCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
'##:::'##::::'###::::'##::::'##::::'###::::'##::::'##::::'###::::::::::::::'#######:::'######:::'######:::::'###:::::'######::'####::'#######::'##::: ##:::::::'######:::'#######::'##::::'##:
. ##:'##::::'## ##::: ###::'###:::'## ##::: ##:::: ##:::'## ##::::::::::::'##.... ##:'##... ##:'##... ##:::'## ##:::'##... ##:. ##::'##.... ##: ###:: ##::::::'##... ##:'##.... ##: ###::'###:
:. ####::::'##:. ##:: ####'####::'##:. ##:: ##:::: ##::'##:. ##::::::::::: ##:::: ##: ##:::..:: ##:::..:::'##:. ##:: ##:::..::: ##:: ##:::: ##: ####: ##:::::: ##:::..:: ##:::: ##: ####'####:
::. ##::::'##:::. ##: ## ### ##:'##:::. ##: #########:'##:::. ##:'#######: ##:::: ##: ##::::::: ##:::::::'##:::. ##:. ######::: ##:: ##:::: ##: ## ## ##:::::: ##::::::: ##:::: ##: ## ### ##:
::: ##:::: #########: ##. #: ##: #########: ##.... ##: #########:........: ##:::: ##: ##::::::: ##::::::: #########::..... ##:: ##:: ##:::: ##: ##. ####:::::: ##::::::: ##:::: ##: ##. #: ##:
::: ##:::: ##.... ##: ##:.:: ##: ##.... ##: ##:::: ##: ##.... ##:::::::::: ##:::: ##: ##::: ##: ##::: ##: ##.... ##:'##::: ##:: ##:: ##:::: ##: ##:. ###:'###: ##::: ##: ##:::: ##: ##:.:: ##:
::: ##:::: ##:::: ##: ##:::: ##: ##:::: ##: ##:::: ##: ##:::: ##::::::::::. #######::. ######::. ######:: ##:::: ##:. ######::'####:. #######:: ##::. ##: ###:. ######::. #######:: ##:::: ##:
:::..:::::..:::::..::..:::::..::..:::::..::..:::::..::..:::::..::::::::::::.......::::......::::......:::..:::::..:::......:::....:::.......:::..::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *yamahaOccasionCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"yamahaOccasion": yamahaOccasionPlugin("yamahaOccasion"), //OP
	}
}

var Plugins yamahaOccasionCommands
