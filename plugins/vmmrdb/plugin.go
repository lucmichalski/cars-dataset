package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucmichalski/cars-contrib/vmmrdb/catalog"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{}

var Resources = []interface{}{}

type vmmrdbPlugin string

func (o vmmrdbPlugin) Name() string      { return string(o) }
func (o vmmrdbPlugin) Section() string   { return `Stanford Cars` }
func (o vmmrdbPlugin) Usage() string     { return `hello` }
func (o vmmrdbPlugin) ShortDesc() string { return `Stanford Cars data importer"` }
func (o vmmrdbPlugin) LongDesc() string  { return o.ShortDesc() }

func (o vmmrdbPlugin) Migrate() []interface{} {
	return Tables
}

func (o vmmrdbPlugin) Resources(Admin *admin.Admin) {}

func (o vmmrdbPlugin) Crawl(cfg *config.Config) error {
	return errors.New("Not implemented")
}

func (o vmmrdbPlugin) Catalog(cfg *config.Config) error {
	return catalog.ImportFromURL(cfg)
}

func (o vmmrdbPlugin) Config() *config.Config {
	cfg := &config.Config{
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
		CatalogURL:  "./shared/datasets/vmmrdb/listing.csv",
		ImageDirs:   []string{"./shared/datasets/vmmrdb/"},
	}
	return cfg
}

type vmmrdbCommands struct{}

func (t *vmmrdbCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
-----------------------------------------------------------------
'##::::'##:'##::::'##:'##::::'##:'########::'########::'########::
 ##:::: ##: ###::'###: ###::'###: ##.... ##: ##.... ##: ##.... ##:
 ##:::: ##: ####'####: ####'####: ##:::: ##: ##:::: ##: ##:::: ##:
 ##:::: ##: ## ### ##: ## ### ##: ########:: ##:::: ##: ########::
. ##:: ##:: ##. #: ##: ##. #: ##: ##.. ##::: ##:::: ##: ##.... ##:
:. ## ##::: ##:.:: ##: ##:.:: ##: ##::. ##:: ##:::: ##: ##:::: ##:
::. ###:::: ##:::: ##: ##:::: ##: ##:::. ##: ########:: ########::
:::...:::::..:::::..::..:::::..::..:::::..::........:::........:::
`)

	return nil
}

func (t *vmmrdbCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"vmmrdb": vmmrdbPlugin("vmmrdb"), //OP
	}
}

var Plugins vmmrdbCommands
