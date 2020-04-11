package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/carvana.com/admin"
	"github.com/lucmichalski/cars-contrib/carvana.com/catalog"
	"github.com/lucmichalski/cars-contrib/carvana.com/crawler"
	"github.com/lucmichalski/cars-contrib/carvana.com/indexer"
	"github.com/lucmichalski/cars-contrib/carvana.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)


var Tables = []interface{}{
	&models.SettingCarvana{}, /*&models.ImageCarvana{}*/
}

var Resources = []interface{}{
	&models.SettingCarvana{},
}

type carvanaPlugin string

func (o carvanaPlugin) Name() string      { return string(o) }
func (o carvanaPlugin) Section() string   { return `carvana.com` }
func (o carvanaPlugin) Usage() string     { return `hello` }
func (o carvanaPlugin) ShortDesc() string { return `carvana.com crawler"` }
func (o carvanaPlugin) LongDesc() string  { return o.ShortDesc() }

func (o carvanaPlugin) Migrate() []interface{} {
	return Tables
}

func (o carvanaPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o carvanaPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o carvanaPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.carvana.com", "carvana.com"},
		URLs: []string{
			"https://www.carvana.com/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
	}
	return cfg
}

type carvanaCommands struct{}

func (t *carvanaCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
---------------------------------------------------------------------------------------------------------------
:'######:::::'###::::'########::'##::::'##::::'###::::'##::: ##::::'###::::::::::'######:::'#######::'##::::'##:
'##... ##:::'## ##::: ##.... ##: ##:::: ##:::'## ##::: ###:: ##:::'## ##::::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:::'##:. ##:: ##:::: ##: ##:::: ##::'##:. ##:: ####: ##::'##:. ##::::::: ##:::..:: ##:::: ##: ####'####:
 ##:::::::'##:::. ##: ########:: ##:::: ##:'##:::. ##: ## ## ##:'##:::. ##:::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: #########: ##.. ##:::. ##:: ##:: #########: ##. ####: #########:::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##.... ##: ##::. ##:::. ## ##::: ##.... ##: ##:. ###: ##.... ##:'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ##:::: ##: ##:::. ##:::. ###:::: ##:::: ##: ##::. ##: ##:::: ##: ###:. ######::. #######:: ##:::: ##:
:......:::..:::::..::..:::::..:::::...:::::..:::::..::..::::..::..:::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *carvanaCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"carvana": carvanaPlugin("carvana"), //OP
	}
}

var Plugins carvanaCommands
