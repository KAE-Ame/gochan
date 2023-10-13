package main

import (
	"fmt"

	"github.com/gochan-org/gochan/pkg/events"
	"github.com/gochan-org/gochan/pkg/gcplugin"
	"github.com/hashicorp/go-plugin"
)

type EventsPlugin struct {
}

func (ep EventsPlugin) Register(triggers []string, handler func(string, ...interface{})) {
	fmt.Println("Register called from plugin")
}

func (ep EventsPlugin) Trigger(trigger string, data ...interface{}) (bool, error, bool) {
	fmt.Println("Trigger called from plugin")
	return false, nil, false
}

func main() {
	pluginMap := map[string]plugin.Plugin{
		"eventplug": &events.EventPlugin{
			Impl: EventsPlugin{},
		},
	}
	gcplugin.SetupRPCPluginLogger("rpc-events")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: gcplugin.RPCHandshakeConfig,
		Logger:          gcplugin.PluginLogger(),
		Plugins:         pluginMap,
	})
}
