package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucmichalski/cars-contrib/carvana-kaggle/catalog"
	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{}
var Resources = []interface{}{}

type carvanaKagglePlugin string

func (o carvanaKagglePlugin) Name() string      { return string(o) }
func (o carvanaKagglePlugin) Section() string   { return `carvana-kaggle` }
func (o carvanaKagglePlugin) Usage() string     { return `hello` }
func (o carvanaKagglePlugin) ShortDesc() string { return `carvana-kaggle data importer"` }
func (o carvanaKagglePlugin) LongDesc() string  { return o.ShortDesc() }

func (o carvanaKagglePlugin) Migrate() []interface{} {
	return Tables
}

func (o carvanaKagglePlugin) Resources(Admin *admin.Admin) {}

func (o carvanaKagglePlugin) Crawl(cfg *config.Config) error {
	return errors.New("not implemented")
}

func (o carvanaKagglePlugin) Catalog(cfg *config.Config) error {
	return catalog.ImportFromURL(cfg)
}

func (o carvanaKagglePlugin) Config() *config.Config {
	cfg := &config.Config{
		AnalyzerURL: "http://localhost:9003/crop?url=%s",
		CatalogURL:  "./shared/datasets/kaggle/metadata.csv",
		ImageDirs:   []string{"./shared/datasets/kaggle/train_hq", "./shared/datasets/kaggle/test_hq"},
	}
	return cfg
}

type carvanaKaggleCommands struct{}

func (t *carvanaKaggleCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
--------------------------------------------------------------------------------------------------------------------------------------------------
:'######:::::'###::::'########::'##::::'##::::'###::::'##::: ##::::'###:::::::::::::'##:::'##::::'###:::::'######::::'######:::'##:::::::'########:
'##... ##:::'## ##::: ##.... ##: ##:::: ##:::'## ##::: ###:: ##:::'## ##:::::::::::: ##::'##::::'## ##:::'##... ##::'##... ##:: ##::::::: ##.....::
 ##:::..:::'##:. ##:: ##:::: ##: ##:::: ##::'##:. ##:: ####: ##::'##:. ##::::::::::: ##:'##::::'##:. ##:: ##:::..::: ##:::..::: ##::::::: ##:::::::
 ##:::::::'##:::. ##: ########:: ##:::: ##:'##:::. ##: ## ## ##:'##:::. ##:'#######: #####::::'##:::. ##: ##::'####: ##::'####: ##::::::: ######:::
 ##::::::: #########: ##.. ##:::. ##:: ##:: #########: ##. ####: #########:........: ##. ##::: #########: ##::: ##:: ##::: ##:: ##::::::: ##...::::
 ##::: ##: ##.... ##: ##::. ##:::. ## ##::: ##.... ##: ##:. ###: ##.... ##:::::::::: ##:. ##:: ##.... ##: ##::: ##:: ##::: ##:: ##::::::: ##:::::::
. ######:: ##:::: ##: ##:::. ##:::. ###:::: ##:::: ##: ##::. ##: ##:::: ##:::::::::: ##::. ##: ##:::: ##:. ######:::. ######::: ########: ########:
:......:::..:::::..::..:::::..:::::...:::::..:::::..::..::::..::..:::::..:::::::::::..::::..::..:::::..:::......:::::......::::........::........::
`)

	return nil
}

func (t *carvanaKaggleCommands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"carvanaKaggle": carvanaKagglePlugin("carvanaKaggle"), //OP
	}
}

var Plugins carvanaKaggleCommands
