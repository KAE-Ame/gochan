package gcplugin

import (
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
)

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
