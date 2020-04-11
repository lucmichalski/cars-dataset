package plugins

import (
	"context"

	"github.com/qor/admin"

	"github.com/lucmichalski/cars-dataset/pkg/config"
)

// Module a plugin that can be initialized
type Module interface {
	Init(context.Context) error
}

type Plugin interface {
	Name() string
	Usage() string
	Section() string
	ShortDesc() string
	LongDesc() string
	Migrate() []interface{}
	Resources(Admin *admin.Admin)
	Config() *config.Config
	Crawl(cfg *config.Config) error
}

// Plugins a plugin that contains one or more command
type Plugins interface {
	Module
	Registry() map[string]Plugin
}
