package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/msglog"
	"github.com/bobwong89757/cellnet/relay"
	"github.com/bobwong89757/cellnet/rpc"
)

// MsgHooker 消息钩子实现
// 实现 EventHooker 接口，带有 RPC 和 relay 功能
// 在消息进入和离开时进行拦截和处理
type MsgHooker struct {
}

// OnInboundEvent 消息入口监听（实现 EventHooker 接口）
// inputEvent: 输入事件
// 处理流程：1. 尝试 RPC 处理 2. 尝试 Relay 处理 3. 记录接收日志
// 返回处理后的输出事件
func (self MsgHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	var handled bool
	var err error

	// 尝试 RPC 处理（处理 RPC 请求和响应）
	inputEvent, handled, err = rpc.ResolveInboundEvent(inputEvent)

	if err != nil {
		log.GetLog().Errorf("rpc.ResolveInboundEvent:", err)
		return
	}

	// 如果 RPC 未处理，尝试 Relay 处理
	if !handled {

		inputEvent, handled, err = relay.ResoleveInboundEvent(inputEvent)

		if err != nil {
			log.GetLog().Errorf("relay.ResoleveInboundEvent:", err)
			return
		}

		// 如果都未处理，记录接收日志
		if !handled {
			msglog.WriteRecvLogger("tcp", inputEvent.Session(), inputEvent.Message())
		}
	}

	return inputEvent
}

// OnOutboundEvent 消息出口监听（实现 EventHooker 接口）
// inputEvent: 输入事件
// 处理流程：1. 尝试 RPC 处理 2. 尝试 Relay 处理 3. 记录发送日志
// 返回处理后的输出事件
func (self MsgHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	// 尝试 RPC 处理（处理 RPC 请求和响应）
	handled, err := rpc.ResolveOutboundEvent(inputEvent)

	if err != nil {
		log.GetLog().Errorf("rpc.ResolveOutboundEvent:", err)
		return nil
	}

	// 如果 RPC 未处理，尝试 Relay 处理
	if !handled {

		handled, err = relay.ResolveOutboundEvent(inputEvent)

		if err != nil {
			log.GetLog().Errorf("relay.ResolveOutboundEvent:", err)
			return nil
		}

		// 如果都未处理，记录发送日志
		if !handled {
			msglog.WriteSendLogger("tcp", inputEvent.Session(), inputEvent.Message())
		}
	}

	return inputEvent
}
