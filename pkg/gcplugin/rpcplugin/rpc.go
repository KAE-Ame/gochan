package rpcplugin

import (
	"os/exec"

	"github.com/gochan-org/gochan/pkg/events"
	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var (
	RPCHandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "gochan-rpc",
		MagicCookieValue: "gochan-rpc",
	}
	rpcPluginLogger hclog.Logger
	clients         []*RPCPluginInfo
)

type RPCPluginInfo struct {
	Path     string
	Events   []string
	client   *plugin.Client
	protocol plugin.ClientProtocol
}

func SetupRPCPluginLogger(name string) {
	rpcPluginLogger = hclog.New(&hclog.LoggerOptions{
		Name:            name,
		Output:          gcutil.RPCLogger(),
		JSONFormat:      false,
		DisableTime:     true,
		IncludeLocation: true,
	})
	hclog.SetDefault(rpcPluginLogger)
}

func PluginLogger() hclog.Logger {
	return rpcPluginLogger
}

func LoadPlugin(info *RPCPluginInfo, autoMTLS bool) error {
	fatalEv := gcutil.LogFatal()
	defer fatalEv.Discard()

	clientCfg := &plugin.ClientConfig{
		Cmd:             exec.Command(info.Path),
		HandshakeConfig: RPCHandshakeConfig,

		Plugins: map[string]plugin.Plugin{
			"eventplugin": &events.EventPlugin{},
		},
		Managed:  true,
		AutoMTLS: autoMTLS,
		Logger:   PluginLogger(),
	}
	info.client = plugin.NewClient(clientCfg)

	var err error
	if info.protocol, err = info.client.Client(); err != nil {
		return err
	}

	clients = append(clients, info)

	raw, err := info.protocol.Dispense("eventplugin")
	if err != nil {
		return err
	}
	ev := raw.(*events.EventClient)
	ev.Trigger("rpc-init", "plugin event triggered from gochan", 1, 2, 3)
	return nil
}
