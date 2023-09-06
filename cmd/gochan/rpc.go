package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcutil"
)

var (
	rpcListener net.Listener
	hello       HelloWorld
)

type HelloWorld int

func (h *HelloWorld) HelloWorld(args *int, reply *int) error {
	fmt.Println("hw.HelloWorld called by RPC client")
	return nil
}

func initRPC() {
	systemCritical := config.GetSystemCriticalConfig()
	rpcCfg := systemCritical.RPC
	err := rpc.RegisterName("hello", &hello)
	if err != nil {
		fmt.Println("Error registering hello rcvr:", err)
		os.Exit(1)
	}

	rpc.HandleHTTP()

	if rpcCfg.Address == systemCritical.HostAndPort() {
		// listen on the same port as the main server
		rpcListener = serverListener
	} else {
		rpcListener, err = net.Listen(rpcCfg.Network, rpcCfg.Address)
		if err != nil {
			if !systemCritical.DebugMode {
				fmt.Printf("Failed listening on network %q, address %q: %s\n",
					rpcCfg.Network, rpcCfg.Address, err.Error())
			}
			gcutil.LogFatal().Err(err).Caller().
				Str("network", rpcCfg.Network).
				Str("address", rpcCfg.Address).
				Send()
		}
		if err = http.Serve(rpcListener, nil); err != nil {
			gcutil.LogFatal().Err(err).Caller().Send()
		}
	}
}

func closeRPC() {
	if rpcListener == nil || rpcListener.Addr() == serverListener.Addr() {
		return
	}
	rpcListener.Close()
}
