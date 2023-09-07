package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"

	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcutil"
)

var (
	rpcListener net.Listener
	rpcServer   *http.Server
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
	fatalEv := gcutil.LogFatal()
	defer fatalEv.Discard()

	var err error
	if rpcCfg.Network == "unix" {
		// using a socket file, make the directory if it doesn't already exist
		socketDir := path.Dir(rpcCfg.Address)
		if err = os.MkdirAll(socketDir, config.GC_DIR_MODE); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Unable to create socket directory %s: %s\n", socketDir, err.Error())
			fatalEv.Err(err).Caller().Str("socketDir", socketDir).Send()
		}
	}
	if err = rpc.RegisterName("hello", &hello); err != nil {
		fmt.Println("Error registering hello receiver:", err)
		fatalEv.Err(err).Caller().Str("rpcName", "hello").Msg("Unable to register receiver")
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
			fatalEv.Err(err).Caller().
				Str("network", rpcCfg.Network).
				Str("address", rpcCfg.Address).
				Send()
		}
		rpcServer = &http.Server{ErrorLog: log.New(gcutil.Logger(), "", 0)}
		if rpcCfg.UseTLS {
			err = rpcServer.ServeTLS(rpcListener, rpcCfg.CertFile, rpcCfg.KeyFile)
		} else {
			err = rpcServer.Serve(rpcListener)
		}
		if err != nil {
			fatalEv.Err(err).Caller().Send()
		}
	}
}

func closeRPC() {
	if rpcListener == nil || rpcListener.Addr() == serverListener.Addr() {
		return
	}
	rpcListener.Close()
}
