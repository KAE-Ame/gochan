package main

import (
	"fmt"
	"os/exec"

	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/events"
	"github.com/gochan-org/gochan/pkg/gcplugin"
	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginLog     hclog.Logger
	pluginClient  *plugin.Client
	pluginClients []plugin.ClientProtocol
)

func initRPC() {
	fatalEv := gcutil.LogFatal()
	defer fatalEv.Discard()

	rpcConfig := config.GetSystemCriticalConfig().RPC

	gcplugin.SetupRPCPluginLogger("rpc")

	for _, pluginCmd := range rpcConfig.RPCPlugins {
		clientCfg := &plugin.ClientConfig{
			Cmd:             exec.Command(pluginCmd),
			HandshakeConfig: gcplugin.RPCHandshakeConfig,

			Plugins: map[string]plugin.Plugin{
				"eventplug": &events.EventPlugin{},
			},
			Managed:  true,
			AutoMTLS: rpcConfig.AutoMTLS,
			Logger:   gcplugin.PluginLogger(),
		}

		pluginClient = plugin.NewClient(clientCfg)

		rpcClient, err := pluginClient.Client()
		if err != nil {
			fmt.Println("Unable to initialize plugin client:", err.Error())
			fatalEv.Err(err).Caller().Msg("Unable to initialize plugin client")
		}
		pluginClients = append(pluginClients, rpcClient)

		raw, err := rpcClient.Dispense("eventplug")
		if err != nil {
			fmt.Println("Unable to dispense eventplug:", err)
			fatalEv.Err(err).Caller().Str("plugin", "eventplug").Send()
		}
		ev := raw.(*events.EventClient)
		ev.Trigger("rpc-events", "plugin event triggered from gochan")
	}
}
