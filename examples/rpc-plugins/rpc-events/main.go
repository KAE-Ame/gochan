package main

import (
	"fmt"
	"os"

	"github.com/gochan-org/gochan/pkg/events"
	"github.com/gochan-org/gochan/pkg/gcplugin/rpcplugin"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

type EventsPlugin struct {
	logger hclog.Logger
}

func (ep EventsPlugin) Register(triggers []string) {
	ep.logger.Info("Register called from plugin for " + fmt.Sprint(triggers))
}

func (ep EventsPlugin) Trigger(trigger string, data ...interface{}) (bool, error, bool) {
	args := data[0].(events.EventData)
	ep.logger.Info("Trigger() called in plugin", "trigger", trigger, "args", args)
	for id, val := range args {
		ep.logger.Info("Element in args", "id", id, "val", val)
	}
	return false, nil, false
}

func main() {
	logFile, err := os.OpenFile("/var/log/gochan/gochan_rpc.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		hclog.Default().Error(err.Error())
		return
	}
	defer logFile.Close()

	eventplug := &EventsPlugin{
		logger: hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Trace,
			Output:     logFile,
			JSONFormat: true,
		}),
	}
	defer func() {
		a := recover()
		if a != nil {
			eventplug.logger.Error("Panic:", "cause", a)
		}
	}()
	eventplug.logger.Debug("Hello from rpc-events plugin!")

	var plugins = plugin.PluginSet{
		"eventplugin": &events.EventPlugin{
			Impl: eventplug,
		},
	}

	rpcplugin.SetupRPCPluginLogger("rpc-events")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: rpcplugin.RPCHandshakeConfig,
		Logger:          eventplug.logger,
		Plugins:         plugins,
	})
}
