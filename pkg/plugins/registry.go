package plugins

import (
	"context"
)

type CarPlugin struct {
	Ctx      context.Context
	Commands map[string]Plugin
	Closed   chan struct{}
}

func New() *CarPlugin {
	return &CarPlugin{
		// pluginsDir: plug.PluginsDir,
		Ctx:      context.Background(),
		Commands: make(map[string]Plugin),
		Closed:   make(chan struct{}),
	}
}
