package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/qor/admin"

	adm "github.com/lucmichalski/cars-contrib/autoreflex.com/admin"
	"github.com/lucmichalski/cars-contrib/autoreflex.com/crawler"
	"github.com/lucmichalski/cars-contrib/autoreflex.com/models"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutoReflex{},
}

var Resources = []interface{}{
	&models.SettingAutoReflex{},
}

type autoReflexPlugin string

func (o autoReflexPlugin) Name() string      { return string(o) }
func (o autoReflexPlugin) Section() string   { return `1001pneus.fr` }
func (o autoReflexPlugin) Usage() string     { return `hello` }
func (o autoReflexPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o autoReflexPlugin) LongDesc() string  { return o.ShortDesc() }

func (o autoReflexPlugin) Migrate() []interface{} {
	return Tables
}

func (o autoReflexPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autoReflexPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autoReflexPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o autoReflexPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.autoreflex.com", "autoreflex.com"},
		URLs: []string{
			"http://www.autoreflex.com/sitemap/index.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  true,
		AnalyzerURL:     "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autoReflexCommands struct{}

func (t *autoReflexCommands) Init(ctx context.Context) error {
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

func (t *autoReflexCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"autoReflex": autoReflexPlugin("autoReflex"), //OP
	}
}

var Plugins autoReflexCommands
