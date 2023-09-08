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
	fmt.Println("Hello, RPC!")
	return nil
}

func (h *HelloWorld) Sum(args []int, result *int) error {
	*result = 0
	for _, i := range args {
		*result += i
	}
	return nil
}

func initRPC() {
	systemCritical := config.GetSystemCriticalConfig()
	rpcCfg := systemCritical.RPC
	fatalEv := gcutil.LogFatal()
	defer fatalEv.Discard()

	var err error
	if err = rpc.RegisterName("hello", &hello); err != nil {
		fmt.Println("Error registering hello receiver:", err)
		fatalEv.Err(err).Caller().Str("rpcName", "hello").Msg("Unable to register receiver")
	}

	// using a socket file, make the directory if it doesn't already exist
	socketDir := path.Dir(rpcCfg.Socket)
	if err = os.MkdirAll(socketDir, config.GC_DIR_MODE); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Unable to create socket directory %s: %s\n", socketDir, err.Error())
		fatalEv.Err(err).Caller().Str("socketDir", socketDir).Send()
	}

	rpcListener, err = net.Listen(rpcCfg.Network, rpcCfg.Address)
	if err != nil {
		if !systemCritical.DebugMode {
			fmt.Printf("Failed listening to socket %q: %s\n",
				rpcCfg.Socket, err.Error())
		}
		fatalEv.Err(err).Caller().
			Str("socket", rpcCfg.Socket).
			Send()
	}
	rpcServer = &http.Server{
		Addr:     rpcCfg.Address,
		ErrorLog: log.New(gcutil.Logger(), "", 0),
	}
	http.Handle(rpc.DefaultRPCPath, rpc.DefaultServer)
	if rpcCfg.UseTLS {
		err = rpcServer.ServeTLS(rpcListener, rpcCfg.CertFile, rpcCfg.KeyFile)
	} else {
		err = rpcServer.Serve(rpcListener)
	}
	if err != nil {
		fatalEv.Err(err).Caller().Send()
	}
}

func closeRPC() {
	if rpcServer != nil {
		if err := rpcServer.Close(); err != nil {
			gcutil.LogError(err).Msg("Failed closing RPC server")
		}
	}
}
