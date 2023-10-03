package events

import (
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type EventRPC struct {
	client *rpc.Client
}

func (er *EventRPC) Register(triggers []string, handler func(string, ...interface{}) error) {
	err := er.client.Call("Plugin.Register", new(interface{}), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Register called from gochan main")
}
func (er *EventRPC) Trigger(trigger string, data ...interface{}) (bool, error, bool) {
	err := er.client.Call("Plugin.Trigger", new(interface{}), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Trigger called from gochan main")
	return false, nil, false
}

type EventRPCServer struct {
	Impl EventsInterface
}

func (er *EventRPCServer) Register(args interface{}, resp *string) error {
	fmt.Println("register args:", args)
	return nil
}
func (er *EventRPCServer) Trigger(args interface{}, resp *string) error {
	fmt.Println("trigger args:", args)
	return nil
}

type EventPlugin struct {
	Impl EventsInterface
}

func (p *EventPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	fmt.Println("EventPlugin.Server called")
	return &EventRPCServer{Impl: p.Impl}, nil
}

func (p *EventPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	fmt.Println("EventPlugin.Client called")
	return &EventRPC{client: c}, nil
}
