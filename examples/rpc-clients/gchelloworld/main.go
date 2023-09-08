package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

func main() {
	var network string
	var address string
	var useTLS bool
	flag.StringVar(&network, "network", "tcp", "unix or tcp")
	flag.StringVar(&address, "address", "192.168.56.3:80", "if network is tcp, this should be <ip>:<port>, otherwise /path/to/socket")
	flag.BoolVar(&useTLS, "tls", false, "whether or not to use TLS for better security")
	flag.Parse()

	var conn net.Conn
	var err error
	if useTLS {
		conn, err = tls.Dial(network, address, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = net.Dial(network, address)
	}
	log.Println("Dialed")
	if err != nil {
		panic("Error opening connection: " + err.Error())
	}
	var client *rpc.Client

	defer func() {
		if client != nil {
			if err = client.Close(); err != nil {
				fmt.Println("Error closing RPC client:", client.Close())
			}
		}
	}()

	// io.WriteString(conn, "CONNECT "+rpc.DefaultRPCPath+" HTTP/1.0\n\n")
	// log.Println("Wrote CONNECT")
	// resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	// log.Println(err)
	// if err != nil {
	// 	panic("Error reading response: " + err.Error())
	// } else if resp.Status != "200 Connected to Go RPC" {
	// 	panic("unexpected HTTP response: " + resp.Status)
	// } else {
	client = rpc.NewClient(conn)
	// }

	log.Println("Calling hello.HelloWorld")
	var args int
	var result int
	if err = client.Call("hello.HelloWorld", &args, &result); err != nil {
		log.Fatalln("Error calling hello.HelloWorld:", err.Error())
	}
	log.Println(args, result)
}
