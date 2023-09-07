package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/gochan-org/gochan/pkg/config"
)

func main() {
	config.InitConfig("3.8.0")
	rpcConfig := config.GetSystemCriticalConfig().RPC
	var conn net.Conn
	if rpcConfig == nil {
		panic("RPC is not configured in the Gochan configuration")
	}
	var err error
	if rpcConfig.UseTLS {
		conn, err = tls.Dial(rpcConfig.Network, rpcConfig.Address, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = net.Dial(rpcConfig.Network, rpcConfig.Address)
	}

	if err != nil {
		panic("Error opening connection: " + err.Error())
	}
	var client *rpc.Client

	defer func() {
		if client == nil {
			return
		}
		if err = client.Close(); err != nil {
			fmt.Println("Error closing RPC client:", client.Close())
		}
	}()

	io.WriteString(conn, "CONNECT "+rpc.DefaultRPCPath+" HTTP/1.0\n\n")

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		panic("Error reading response: " + err.Error())
	} else if resp.Status != "200 Connected to Go RPC" {
		panic("unexpected HTTP response: " + resp.Status)
	} else {
		client = rpc.NewClient(conn)
	}

	log.Println("Calling hello.HelloWorld")
	var args int
	var result int
	if err = client.Call("hello.HelloWorld", &args, &result); err != nil {
		log.Fatalln("Error calling hello.HelloWorld:", err.Error())
	}
	log.Println(args, result)
}
