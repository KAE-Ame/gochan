package events

import (
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type EventPlugin struct {
	Impl Event
}

func (p *EventPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	fmt.Println("EventPlugin.Server called")
	return &eventServer{Impl: p.Impl}, nil
}

func (p *EventPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	fmt.Println("EventPlugin.Client called")
	return &EventClient{client: c}, nil
}

type EventClient struct {
	client *rpc.Client
}

func (er *EventClient) Register(triggers []string, handler func(string, ...interface{}) error) {
	err := er.client.Call("Plugin.Register", new(interface{}), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Register called from gochan main")
}
func (er *EventClient) Trigger(trigger string, data ...interface{}) (bool, error, bool) {
	err := er.client.Call("Plugin.Trigger", new(interface{}), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Trigger called from gochan main")
	return false, nil, false
}

type eventServer struct {
	Impl Event
}

func (er *eventServer) Register(args interface{}, resp *string) error {
	fmt.Println("register args:", args)
	return nil
}
func (er *eventServer) Trigger(args interface{}, resp *EventTriggerResult) error {
	var res EventTriggerResult
	res.Handled, res.Error, res.Recovered = er.Impl.Trigger("rpc-events")
	return nil
}
