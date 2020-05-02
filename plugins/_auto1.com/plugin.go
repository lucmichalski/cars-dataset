package main

import (
	"context"
	"errors"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/auto1.com/admin"
	"github.com/lucmichalski/cars-contrib/auto1.com/crawler"
	"github.com/lucmichalski/cars-contrib/auto1.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAuto1{},
}

var Resources = []interface{}{
	&models.SettingAuto1{},
}

type auto1Plugin string

func (o auto1Plugin) Name() string      { return string(o) }
func (o auto1Plugin) Section() string   { return `auto1.com` }
func (o auto1Plugin) Usage() string     { return `hello` }
func (o auto1Plugin) ShortDesc() string { return `auto1.com crawler"` }
func (o auto1Plugin) LongDesc() string  { return o.ShortDesc() }

func (o auto1Plugin) Migrate() []interface{} {
	return Tables
}

func (o auto1Plugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o auto1Plugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o auto1Plugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o auto1Plugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.auto1.com", "auto1.com"},
		URLs: []string{
			"https://www.auto1.com",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  false,
		AnalyzerURL:     "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type auto1Commands struct{}

func (t *auto1Commands) Init(ctx context.Context) error {
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

func (t *auto1Commands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"auto1": auto1Plugin("auto1"), //OP
	}
}

var Plugins auto1Commands
