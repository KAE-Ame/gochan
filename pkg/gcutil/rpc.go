package gcutil

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/rpc"
)

// NewRPCClient acts as a wrapper around rpc.Dial(), allowing the client to connect to a TLS-secured socket
func DialRPC(network string, address string, tlsConfig *tls.Config) (*rpc.Client, error) {
	var conn net.Conn
	var err error
	if tlsConfig == nil {
		conn, err = net.Dial(network, address)
	} else {
		conn, err = tls.Dial(network, address, tlsConfig)
	}
	if err != nil {
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+rpc.DefaultRPCPath+" HTTP/1.0\n\n")
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		return nil, err
	} else if resp.Status != "200 Connected to Go RPC" {
		return nil, errors.New("unexpected HTTP response: " + resp.Status)
	}
	return rpc.NewClient(conn), nil
}
