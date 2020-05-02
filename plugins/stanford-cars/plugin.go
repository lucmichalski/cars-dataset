package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucmichalski/cars-contrib/stanford-cars/catalog"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{}

var Resources = []interface{}{}

type stanfordCarsPlugin string

func (o stanfordCarsPlugin) Name() string      { return string(o) }
func (o stanfordCarsPlugin) Section() string   { return `Stanford Cars` }
func (o stanfordCarsPlugin) Usage() string     { return `hello` }
func (o stanfordCarsPlugin) ShortDesc() string { return `Stanford Cars data importer"` }
func (o stanfordCarsPlugin) LongDesc() string  { return o.ShortDesc() }

func (o stanfordCarsPlugin) Migrate() []interface{} {
	return Tables
}

func (o stanfordCarsPlugin) Resources(Admin *admin.Admin) {}

func (o stanfordCarsPlugin) Crawl(cfg *config.Config) error {
	return errors.New("Not implemented")
}

func (o stanfordCarsPlugin) Catalog(cfg *config.Config) error {
	return catalog.ImportFromURL(cfg)
}

func (o stanfordCarsPlugin) Config() *config.Config {
	cfg := &config.Config{
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
		CatalogURL:  "./shared/datasets/stanford-cars/data/cars_data.csv",
		ImageDirs:   []string{"./shared/datasets/stanford-cars/cars_test", "./shared/datasets/stanford-cars/cars_train"},
	}
	return cfg
}

type stanfordCarsCommands struct{}

func (t *stanfordCarsCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
--------------------------------------------------------------------------------------------------------------------------------------
:'######::'########::::'###::::'##::: ##:'########::'#######::'########::'########::::::::::::'######:::::'###::::'########:::'######::
'##... ##:... ##..::::'## ##::: ###:: ##: ##.....::'##.... ##: ##.... ##: ##.... ##::::::::::'##... ##:::'## ##::: ##.... ##:'##... ##:
 ##:::..::::: ##:::::'##:. ##:: ####: ##: ##::::::: ##:::: ##: ##:::: ##: ##:::: ##:::::::::: ##:::..:::'##:. ##:: ##:::: ##: ##:::..::
. ######::::: ##::::'##:::. ##: ## ## ##: ######::: ##:::: ##: ########:: ##:::: ##:'#######: ##:::::::'##:::. ##: ########::. ######::
:..... ##:::: ##:::: #########: ##. ####: ##...:::: ##:::: ##: ##.. ##::: ##:::: ##:........: ##::::::: #########: ##.. ##::::..... ##:
'##::: ##:::: ##:::: ##.... ##: ##:. ###: ##::::::: ##:::: ##: ##::. ##:: ##:::: ##:::::::::: ##::: ##: ##.... ##: ##::. ##::'##::: ##:
. ######::::: ##:::: ##:::: ##: ##::. ##: ##:::::::. #######:: ##:::. ##: ########:::::::::::. ######:: ##:::: ##: ##:::. ##:. ######::
:......::::::..:::::..:::::..::..::::..::..:::::::::.......:::..:::::..::........:::::::::::::......:::..:::::..::..:::::..:::......:::
`)

	return nil
}

func (t *stanfordCarsCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"stanfordCars": stanfordCarsPlugin("stanfordCars"), //OP
	}
}

var Plugins stanfordCarsCommands
