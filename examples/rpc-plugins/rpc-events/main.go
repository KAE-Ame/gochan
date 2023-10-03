package main

import (
	"fmt"
	"os"

	"github.com/gochan-org/gochan/pkg/events"
	"github.com/hashicorp/go-hclog"
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
	fi, err := os.OpenFile("rpc.log", os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "gochan-rpc",
			MagicCookieValue: "gochan-rpc",
		},
		Logger: hclog.New(&hclog.LoggerOptions{
			Output:     fi,
			JSONFormat: true,
		}),
		Plugins: pluginMap,
	})
}
