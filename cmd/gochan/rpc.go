package main

import (
	"fmt"
	"os/exec"

	"github.com/gochan-org/gochan/pkg/events"
	"github.com/gochan-org/gochan/pkg/gcplugin"
	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginLog    hclog.Logger
	pluginClient *plugin.Client
)

func initRPC() {
	fatalEv := gcutil.LogFatal()
	defer fatalEv.Discard()

	gcplugin.SetupRPCPluginLogger("rpc")

	pluginClient = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "gochan-rpc",
			MagicCookieValue: "gochan-rpc",
		},
		Plugins: map[string]plugin.Plugin{
			"eventplug": &events.EventPlugin{},
		},
		Cmd: exec.Command("/vagrant/gochan-events-rpc"),
	})
	defer pluginClient.Kill()

	rpcClient, err := pluginClient.Client()
	if err != nil {
		fmt.Println("Unable to initialize plugin client:", err.Error())
		fatalEv.Err(err).Caller().Msg("Unable to initialize plugin client")
	}

	raw, err := rpcClient.Dispense("eventplug")
	if err != nil {
		fmt.Println("Unable to dispense eventplug:", err)
		fatalEv.Err(err).Caller().Str("plugin", "eventplug").Send()
	}
	ev := raw.(*events.EventRPC)
	fmt.Println("EventsRPC:", ev)
}

func closeRPC() {

	// if rpcServer != nil {
	// 	if err := rpcServer.Close(); err != nil {
	// 		gcutil.LogError(err).Msg("Failed closing RPC server")
	// 	}
	// }
}
