package main

import (
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.DialHTTP("unix", "/var/run/gochan/gochanrpc.sock")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer client.Close()
	log.Println("Calling hello.HelloWorld")
	var args int
	var result int
	if err = client.Call("hello.HelloWorld", &args, &result); err != nil {
		log.Fatalln("Error calling hello.HelloWorld:", err.Error())
	}
	log.Println(args, result)
}
