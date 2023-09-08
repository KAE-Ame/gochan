package main

import (
	"crypto/tls"
	"flag"
	"log"

	"github.com/gochan-org/gochan/pkg/gcutil"
)

func main() {
	var socket string
	var useTLS bool
	flag.StringVar(&socket, "socket", "/var/run/gochan/rpc.sock", "Path to the RPC socket")
	flag.BoolVar(&useTLS, "tls", false, "whether or not to use TLS for better security")
	flag.Parse()

	var tlsConfig *tls.Config
	if useTLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client, err := gcutil.DialRPC("unix", socket, tlsConfig)
	if err != nil {
		log.Fatalln("Error creating RPC client:", err)
	}

	log.Println("Calling hello.HelloWorld")
	var args int
	var result int
	if err = client.Call("hello.HelloWorld", &args, &result); err != nil {
		log.Fatalln("Error calling hello.HelloWorld:", err.Error())
	}
	log.Println("Calling hello.Sum")
	arr := []int{1, 2, 3, 4}
	if err = client.Call("hello.Sum", arr, &result); err != nil {
		log.Fatalln("Erro calling hello.Sum:", err.Error())
	}
	log.Println("Sum of elements of", arr, "=", result)
}
