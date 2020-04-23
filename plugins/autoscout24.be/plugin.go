package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/autoscout24.be/admin"
	"github.com/lucmichalski/cars-contrib/autoscout24.be/crawler"
	"github.com/lucmichalski/cars-contrib/autoscout24.be/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingAutoScout24Be{},
}

var Resources = []interface{}{
	&models.SettingAutoScout24Be{},
}

type autoScout24BePlugin string

func (o autoScout24BePlugin) Name() string      { return string(o) }
func (o autoScout24BePlugin) Section() string   { return `autoscout24.be` }
func (o autoScout24BePlugin) Usage() string     { return `hello` }
func (o autoScout24BePlugin) ShortDesc() string { return `autoscout24.be crawler"` }
func (o autoScout24BePlugin) LongDesc() string  { return o.ShortDesc() }

func (o autoScout24BePlugin) Migrate() []interface{} {
	return Tables
}

func (o autoScout24BePlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o autoScout24BePlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o autoScout24BePlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o autoScout24BePlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.autoscout24.be", "autoscout24.be"},
		URLs: []string{
			"https://autoscout24.be/fr",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 6,
		IsSitemapIndex: false,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type autoScout24BeCommands struct{}

func (t *autoScout24BeCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
----------------------------------------------------------------------------------------------------------------------------------------------
:::'###::::'##::::'##:'########::'#######:::'######:::'######:::'#######::'##::::'##:'########::'#######::'##:::::::::::::'########::'########:
::'## ##::: ##:::: ##:... ##..::'##.... ##:'##... ##:'##... ##:'##.... ##: ##:::: ##:... ##..::'##.... ##: ##:::'##::::::: ##.... ##: ##.....::
:'##:. ##:: ##:::: ##:::: ##:::: ##:::: ##: ##:::..:: ##:::..:: ##:::: ##: ##:::: ##:::: ##::::..::::: ##: ##::: ##::::::: ##:::: ##: ##:::::::
'##:::. ##: ##:::: ##:::: ##:::: ##:::: ##:. ######:: ##::::::: ##:::: ##: ##:::: ##:::: ##:::::'#######:: ##::: ##::::::: ########:: ######:::
 #########: ##:::: ##:::: ##:::: ##:::: ##::..... ##: ##::::::: ##:::: ##: ##:::: ##:::: ##::::'##:::::::: #########:::::: ##.... ##: ##...::::
 ##.... ##: ##:::: ##:::: ##:::: ##:::: ##:'##::: ##: ##::: ##: ##:::: ##: ##:::: ##:::: ##:::: ##::::::::...... ##::'###: ##:::: ##: ##:::::::
 ##:::: ##:. #######::::: ##::::. #######::. ######::. ######::. #######::. #######::::: ##:::: #########::::::: ##:: ###: ########:: ########:
..:::::..:::.......::::::..::::::.......::::......::::......::::.......::::.......::::::..:::::.........::::::::..:::...::........:::........::
`)

	return nil
}

func (t *autoScout24BeCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"autoscout24be": autoScout24BePlugin("autoscout24be"), //OP
	}
}

var Plugins autoScout24BeCommands
