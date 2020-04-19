package main

import (
	"context"
	"fmt"
	"errors"

	adm "github.com/lucmichalski/cars-contrib/classicdriver.com/admin"
	"github.com/lucmichalski/cars-contrib/classicdriver.com/crawler"
	"github.com/lucmichalski/cars-contrib/classicdriver.com/models"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingClassicDriver{},
}

var Resources = []interface{}{
	&models.SettingClassicDriver{},
}

type classicDriverPlugin string

func (o classicDriverPlugin) Name() string      { return string(o) }
func (o classicDriverPlugin) Section() string   { return `classicdriver.com` }
func (o classicDriverPlugin) Usage() string     { return `hello` }
func (o classicDriverPlugin) ShortDesc() string { return `classicdriver.com crawler"` }
func (o classicDriverPlugin) LongDesc() string  { return o.ShortDesc() }

func (o classicDriverPlugin) Migrate() []interface{} {
	return Tables
}

func (o classicDriverPlugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o classicDriverPlugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o classicDriverPlugin) Catalog(cfg *config.Config) error {
	return errors.New("Not Implemented")
}

func (o classicDriverPlugin) Config() *config.Config {
	cfg := &config.Config{
		AllowedDomains: []string{"www.classicdriver.com", "classicdriver.com"},
		URLs: []string{
		    //"https://www.classicdriver.com/en/sitemap.xml?page=1",
		    "https://www.classicdriver.com/en/sitemap.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 1,
		//IsSitemapIndex: true,
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
	}
	return cfg
}

type classicDriverCommands struct{}

func (t *classicDriverCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-------------------------------------------------------------------------------------------------------------------------------------------------------------------
:'######::'##::::::::::'###:::::'######:::'######::'####::'######::'########::'########::'####:'##::::'##:'########:'########::::::::'######:::'#######::'##::::'##:
'##... ##: ##:::::::::'## ##:::'##... ##:'##... ##:. ##::'##... ##: ##.... ##: ##.... ##:. ##:: ##:::: ##: ##.....:: ##.... ##::::::'##... ##:'##.... ##: ###::'###:
 ##:::..:: ##::::::::'##:. ##:: ##:::..:: ##:::..::: ##:: ##:::..:: ##:::: ##: ##:::: ##:: ##:: ##:::: ##: ##::::::: ##:::: ##:::::: ##:::..:: ##:::: ##: ####'####:
 ##::::::: ##:::::::'##:::. ##:. ######::. ######::: ##:: ##::::::: ##:::: ##: ########::: ##:: ##:::: ##: ######::: ########::::::: ##::::::: ##:::: ##: ## ### ##:
 ##::::::: ##::::::: #########::..... ##::..... ##:: ##:: ##::::::: ##:::: ##: ##.. ##:::: ##::. ##:: ##:: ##...:::: ##.. ##:::::::: ##::::::: ##:::: ##: ##. #: ##:
 ##::: ##: ##::::::: ##.... ##:'##::: ##:'##::: ##:: ##:: ##::: ##: ##:::: ##: ##::. ##::: ##:::. ## ##::: ##::::::: ##::. ##::'###: ##::: ##: ##:::: ##: ##:.:: ##:
. ######:: ########: ##:::: ##:. ######::. ######::'####:. ######:: ########:: ##:::. ##:'####:::. ###:::: ########: ##:::. ##: ###:. ######::. #######:: ##:::: ##:
:......:::........::..:::::..:::......::::......:::....:::......:::........:::..:::::..::....:::::...:::::........::..:::::..::...:::......::::.......:::..:::::..::
`)

	return nil
}

func (t *classicDriverCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"classicDriver": classicDriverPlugin("classicDriver"), //OP
	}
}

var Plugins classicDriverCommands
