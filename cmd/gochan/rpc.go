package main

import (
	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcplugin/rpcplugin"
	"github.com/gochan-org/gochan/pkg/gcutil"
)

func initRPC() {
	fatalEv := gcutil.LogFatal()
	defer fatalEv.Discard()

	rpcConfig := config.GetSystemCriticalConfig().RPC

	rpcplugin.SetupRPCPluginLogger("rpc")

	for _, info := range rpcConfig.RPCPlugins {
		rpcplugin.LoadPlugin(&info, rpcConfig.AutoMTLS)
	}
}
