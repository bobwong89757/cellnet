package proc

import (
	"github.com/bobwong89757/cellnet"
)

// ProcessorBundle
// @Description: 处理器设置接口，由各Peer实现
type ProcessorBundle interface {

	// 设置 传输器，负责收发消息
	SetTransmitter(v cellnet.MessageTransmitter)

	// 设置 接收后，发送前的事件处理流程
	SetHooker(v cellnet.EventHooker)

	// 设置 接收后最终处理回调
	SetCallback(v cellnet.EventCallback)
}

// NewQueuedEventCallback
//
//	@Description: 让EventCallback保证放在ses的队列里，而不是并发的
//	@param callback
//	@return cellnet.EventCallback
func NewQueuedEventCallback(callback cellnet.EventCallback) cellnet.EventCallback {

	return func(ev cellnet.Event) {
		if callback != nil {
			cellnet.SessionQueuedCall(ev.Session(), func() {

				callback(ev)
			})
		}
	}
}

// 当需要多个Hooker时，使用NewMultiHooker将多个hooker合并成1个hooker处理
type MultiHooker []cellnet.EventHooker

// OnInboundEvent
//
//	@Description: 消息入口
//	@receiver self
//	@param input
//	@return output
func (self MultiHooker) OnInboundEvent(input cellnet.Event) (output cellnet.Event) {

	for _, h := range self {

		input = h.OnInboundEvent(input)

		if input == nil {
			break
		}
	}

	return input
}

// OnOutboundEvent
//
//	@Description: 消息出口
//	@receiver self
//	@param input
//	@return output
func (self MultiHooker) OnOutboundEvent(input cellnet.Event) (output cellnet.Event) {

	for _, h := range self {

		input = h.OnOutboundEvent(input)

		if input == nil {
			break
		}
	}

	return input
}

// NewMultiHooker
//
//	@Description: 新建MultiHooker
//	@param h
//	@return cellnet.EventHooker
func NewMultiHooker(h ...cellnet.EventHooker) cellnet.EventHooker {

	return MultiHooker(h)
}
