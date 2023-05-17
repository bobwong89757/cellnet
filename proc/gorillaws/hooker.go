package gorillaws

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/msglog"
)

// MsgHooker
// @Description: 带有RPC和relay功能
type MsgHooker struct {
}

func (self MsgHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	msglog.WriteRecvLogger("ws", inputEvent.Session(), inputEvent.Message())

	return inputEvent
}

func (self MsgHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	msglog.WriteSendLogger("ws", inputEvent.Session(), inputEvent.Message())

	return inputEvent
}
