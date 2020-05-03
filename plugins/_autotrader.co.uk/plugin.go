package main

import (
	"context"
	"fmt"

	"github.com/qor/admin"

	adm "github.com/lucmichalski/cars-dataset/autotrader.co.uk/admin"
	"github.com/lucmichalski/cars-dataset/autotrader.co.uk/crawler"
	"github.com/lucmichalski/cars-dataset/autotrader.co.uk/models"
	"github.com/lucmichalski/cars-dataset/pkg/config"
	"github.com/lucmichalski/cars-dataset/pkg/plugins"
)

var Tables = []interface{}{
	&models.SettingPneus1001{}, &models.CatalogPneus1001{}, &models.ImagePneus1001{}, &models.CategoryPneus1001{},
}

var Resources = []interface{}{
	&models.SettingPneus1001{},
}

var Catalog models.CatalogPneus1001

type pneus1001Plugin string

func (o pneus1001Plugin) Name() string      { return string(o) }
func (o pneus1001Plugin) Section() string   { return `1001pneus.fr` }
func (o pneus1001Plugin) Usage() string     { return `hello` }
func (o pneus1001Plugin) ShortDesc() string { return `1001pneus.fr crawler"` }
func (o pneus1001Plugin) LongDesc() string  { return o.ShortDesc() }

func (o pneus1001Plugin) Migrate() []interface{} {
	return Tables
}

func (o pneus1001Plugin) Resources(Admin *admin.Admin) {
	adm.ConfigureAdmin(Admin)
}

func (o pneus1001Plugin) Crawl(cfg *config.Config) error {
	return crawler.Extract(cfg)
}

func (o pneus1001Plugin) Config() *config.Config {
	cfg := &config.Config{
		// AllowedDomains: []string{"www.1001pneus.fr", "1001pneus.fr", "m.1001pneus.fr"},
		URLs: []string{
			"https://www.1001pneus.fr/sitemap/cms.xml",
			"https://www.1001pneus.fr/sitemap/station-montage-ville.xml",
			"https://www.1001pneus.fr/sitemap/station-montage.xml",
			"https://www.1001pneus.fr/sitemap/recherche-vehicule.xml",
			"https://www.1001pneus.fr/sitemap/recherche-moto-iciv.xml",
			"https://www.1001pneus.fr/sitemap/recherche-moto.xml",
			"https://www.1001pneus.fr/sitemap/recherche-auto-iciv.xml",
			"https://www.1001pneus.fr/sitemap/recherche-auto.xml",
			"https://www.1001pneus.fr/sitemap/profils-moto.xml",
			"https://www.1001pneus.fr/sitemap/promos.xml",
			"https://www.1001pneus.fr/sitemap/pneus-moto.xml",
			"https://www.1001pneus.fr/sitemap/pneus-auto.xml",
			"https://www.1001pneus.fr/sitemap/marques-moto.xml",
			"https://www.1001pneus.fr/sitemap/marques-auto.xml",
			"https://www.1001pneus.fr/sitemap/general.xml",
		},
		QueueMaxSize:    1000000,
		ConsumerThreads: 35,
	}
	return cfg
}

type pneus1001Commands struct{}

func (t *pneus1001Commands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
--------------------------------------------------------------------------------------------------------------------------------------------------------------
:::'###::::'##::::'##:'########::'#######::'########:'########:::::'###::::'########::'########:'########::::::::'######:::'#######:::::::'##::::'##:'##:::'##:
::'## ##::: ##:::: ##:... ##..::'##.... ##:... ##..:: ##.... ##:::'## ##::: ##.... ##: ##.....:: ##.... ##::::::'##... ##:'##.... ##:::::: ##:::: ##: ##::'##::
:'##:. ##:: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##::'##:. ##:: ##:::: ##: ##::::::: ##:::: ##:::::: ##:::..:: ##:::: ##:::::: ##:::: ##: ##:'##:::
'##:::. ##: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ########::'##:::. ##: ##:::: ##: ######::: ########::::::: ##::::::: ##:::: ##:::::: ##:::: ##: #####::::
 #########: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##.. ##::: #########: ##:::: ##: ##...:::: ##.. ##:::::::: ##::::::: ##:::: ##:::::: ##:::: ##: ##. ##:::
 ##.... ##: ##:::: ##:::: ##:::: ##:::: ##:::: ##:::: ##::. ##:: ##.... ##: ##:::: ##: ##::::::: ##::. ##::'###: ##::: ##: ##:::: ##:'###: ##:::: ##: ##:. ##::
 ##:::: ##:. #######::::: ##::::. #######::::: ##:::: ##:::. ##: ##:::: ##: ########:: ########: ##:::. ##: ###:. ######::. #######:: ###:. #######:: ##::. ##:
..:::::..:::.......::::::..::::::.......::::::..:::::..:::::..::..:::::..::........:::........::..:::::..::...:::......::::.......:::...:::.......:::..::::..::
`)

	return nil
}

func (t *pneus1001Commands) Registry() map[string]plugins.Plugin {
	return map[string]plugins.Plugin{
		"1001pneus": pneus1001Plugin("1001pneus"), //OP
	}
}

var Plugins pneus1001Commands
