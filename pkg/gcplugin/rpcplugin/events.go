package rpcplugin

import (
	"encoding/gob"
	"net/rpc"

	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog"
)

type RPCEvent interface {
	Register([]string)
	Trigger(string, ...interface{}) (bool, error, bool)
}

type EventPlugin struct {
	Impl RPCEvent
}

func (p *EventPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &eventServer{Impl: p.Impl}, nil
}

func (p *EventPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &EventClient{client: c}, nil
}

type EventClient struct {
	client *rpc.Client
}

type EventData []interface{}

type EventTriggeredRequest struct {
	Event string
	Data  []interface{}
}

type EventTriggerResult struct {
	Handled   bool
	Error     error
	Recovered bool
}

func (er *EventClient) Register(triggers []string) {
	var args interface{} = triggers
	err := er.client.Call("Plugin.Register", &args, nil)
	if err != nil {
		gcutil.LogRPC(zerolog.FatalLevel).Err(err).Caller().
			Strs("triggers", triggers).
			Msg("Unable to register plugin")
	}
}
func (er *EventClient) Trigger(trigger string, data ...interface{}) (bool, error, bool) {
	var ed EventData = data
	var args interface{} = ed
	err := er.client.Call("Plugin.Trigger", args, nil)
	if err != nil {
		return false, err, false
	}
	return false, nil, false
}

type eventServer struct {
	Impl RPCEvent
}

func (er *eventServer) Register(args interface{}, resp *string) error {
	gcutil.RPCLogger().Debug().Str("call", "eventServer.Register()").Interface("args", args).Send()
	return nil
}
func (er *eventServer) Trigger(args EventData, resp *EventTriggerResult) error {
	gcutil.RPCLogger().Debug().Str("call", "eventServer.Trigger()").Interface("args", args).Send()
	var res EventTriggerResult
	res.Handled, res.Error, res.Recovered = er.Impl.Trigger("rpc-events", args)
	if resp != nil {
		*resp = res
	}
	return nil
}

func init() {
	gob.Register(EventData{})
}
