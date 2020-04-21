package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/buyacar.co.uk/admin"
	"github.com/lucmichalski/cars-contrib/buyacar.co.uk/crawler"
	"github.com/lucmichalski/cars-contrib/buyacar.co.uk/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingBuyACar{},
}

var Resources = []interface{}{
	&models.SettingBuyACar{},
}

type buyACarPlugin string

func (o buyACarPlugin) Name() string      { return string(o) }
func (o buyACarPlugin) Section() string   { return `1001pneus.fr` }
func (o buyACarPlugin) Usage() string     { return `hello` }
func (o buyACarPlugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o buyACarPlugin) LongDesc() string  { return o.ShortDesc() }

func (o buyACarPlugin) Migrate() []interface{} {
	return Tables
}

func (o buyACarPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o buyACarPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o buyACarPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o buyACarPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.buyacar.co.uk", "buyacar.co.uk"},
		URLs: []string{
			"https://www.buyacar.co.uk/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		IsSitemapIndex:  false,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type buyACarCommands struct{}

func (t *buyACarCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
------------------------------------------------------------------------------------------------------------------------------
'########::'##::::'##:'##:::'##::::'###:::::'######:::::'###::::'########::::::::'######:::'#######:::::::'##::::'##:'##:::'##:
 ##.... ##: ##:::: ##:. ##:'##::::'## ##:::'##... ##:::'## ##::: ##.... ##::::::'##... ##:'##.... ##:::::: ##:::: ##: ##::'##::
 ##:::: ##: ##:::: ##::. ####::::'##:. ##:: ##:::..:::'##:. ##:: ##:::: ##:::::: ##:::..:: ##:::: ##:::::: ##:::: ##: ##:'##:::
 ########:: ##:::: ##:::. ##::::'##:::. ##: ##:::::::'##:::. ##: ########::::::: ##::::::: ##:::: ##:::::: ##:::: ##: #####::::
 ##.... ##: ##:::: ##:::: ##:::: #########: ##::::::: #########: ##.. ##:::::::: ##::::::: ##:::: ##:::::: ##:::: ##: ##. ##:::
 ##:::: ##: ##:::: ##:::: ##:::: ##.... ##: ##::: ##: ##.... ##: ##::. ##::'###: ##::: ##: ##:::: ##:'###: ##:::: ##: ##:. ##::
 ########::. #######::::: ##:::: ##:::: ##:. ######:: ##:::: ##: ##:::. ##: ###:. ######::. #######:: ###:. #######:: ##::. ##:
........::::.......::::::..:::::..:::::..:::......:::..:::::..::..:::::..::...:::......::::.......:::...:::.......:::..::::..::
`)

	return nil
}

func (t *buyACarCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"buyacar": buyACarPlugin("buyacar"), //OP
	}
}

var Plugins buyACarCommands
