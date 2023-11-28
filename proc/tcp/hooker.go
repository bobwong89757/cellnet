package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/msglog"
	"github.com/bobwong89757/cellnet/relay"
	"github.com/bobwong89757/cellnet/rpc"
)

// MsgHooker
// @Description: 消息钩子,带有RPC和relay功能
type MsgHooker struct {
}

// OnInboundEvent
//
//	@Description: 消息入口监听
//	@receiver self
//	@param inputEvent
//	@return outputEvent
func (self MsgHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	var handled bool
	var err error

	inputEvent, handled, err = rpc.ResolveInboundEvent(inputEvent)

	if err != nil {
		log.GetLog().Errorf("rpc.ResolveInboundEvent:", err)
		return
	}

	if !handled {

		inputEvent, handled, err = relay.ResoleveInboundEvent(inputEvent)

		if err != nil {
			log.GetLog().Errorf("relay.ResoleveInboundEvent:", err)
			return
		}

		if !handled {
			msglog.WriteRecvLogger("tcp", inputEvent.Session(), inputEvent.Message())
		}
	}

	return inputEvent
}

// OnOutboundEvent
//
//	@Description: 消息出口监听
//	@receiver self
//	@param inputEvent
//	@return outputEvent
func (self MsgHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	handled, err := rpc.ResolveOutboundEvent(inputEvent)

	if err != nil {
		log.GetLog().Errorf("rpc.ResolveOutboundEvent:", err)
		return nil
	}

	if !handled {

		handled, err = relay.ResolveOutboundEvent(inputEvent)

		if err != nil {
			log.GetLog().Errorf("relay.ResolveOutboundEvent:", err)
			return nil
		}

		if !handled {
			msglog.WriteSendLogger("tcp", inputEvent.Session(), inputEvent.Message())
		}
	}

	return inputEvent
}
