package main

import (
	"context"
	"fmt"

	adm "github.com/lucmichalski/cars-contrib/autosphere.fr/admin"
	"github.com/lucmichalski/cars-contrib/autosphere.fr/crawler"
	"github.com/lucmichalski/cars-contrib/autosphere.fr/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutosphere{},
}

var Resources = []interface{}{
	&models.SettingAutosphere{},
}

type autospherePlugin string

func (o autospherePlugin) Name() string      { return string(o) }
func (o autospherePlugin) Section() string   { return `autosphere.fr` }
func (o autospherePlugin) Usage() string     { return `hello` }
func (o autospherePlugin) ShortDesc() string { return `autosphere.fr crawler"` }
func (o autospherePlugin) LongDesc() string  { return o.ShortDesc() }

func (o autospherePlugin) Migrate() []interface{} {
	return Tables
}

func (o autospherePlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autospherePlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autospherePlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.autosphere.fr", "autosphere.fr"},
		URLs: []string{
			"https://www.autosphere.fr/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autosphereCommands struct{}

func (t *autosphereCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------------------------------------------------------------------------
:::'###::::'##::::'##:'########::'#######:::'######::'########::'##::::'##:'########:'########::'########::::::'########:'########::
::'## ##::: ##:::: ##:... ##..::'##.... ##:'##... ##: ##.... ##: ##:::: ##: ##.....:: ##.... ##: ##.....::::::: ##.....:: ##.... ##:
:'##:. ##:: ##:::: ##:::: ##:::: ##:::: ##: ##:::..:: ##:::: ##: ##:::: ##: ##::::::: ##:::: ##: ##:::::::::::: ##::::::: ##:::: ##:
'##:::. ##: ##:::: ##:::: ##:::: ##:::: ##:. ######:: ########:: #########: ######::: ########:: ######:::::::: ######::: ########::
 #########: ##:::: ##:::: ##:::: ##:::: ##::..... ##: ##.....::: ##.... ##: ##...:::: ##.. ##::: ##...::::::::: ##...:::: ##.. ##:::
 ##.... ##: ##:::: ##:::: ##:::: ##:::: ##:'##::: ##: ##:::::::: ##:::: ##: ##::::::: ##::. ##:: ##:::::::'###: ##::::::: ##::. ##::
 ##:::: ##:. #######::::: ##::::. #######::. ######:: ##:::::::: ##:::: ##: ########: ##:::. ##: ########: ###: ##::::::: ##:::. ##:
..:::::..:::.......::::::..::::::.......::::......:::..:::::::::..:::::..::........::..:::::..::........::...::..::::::::..:::::..::
`)

	return nil
}

func (t *autosphereCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"autosphere": autospherePlugin("autosphere"), //OP
	}
}

var Plugins autosphereCommands
