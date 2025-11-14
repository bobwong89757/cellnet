package proc

import (
	"github.com/bobwong89757/cellnet"
)

// ProcessorBundle 定义处理器设置接口
// 由各 Peer 实现，用于配置消息处理流程
// 包括设置消息传输器、事件钩子和用户回调
type ProcessorBundle interface {
	// SetTransmitter 设置消息传输器
	// v: 消息传输器，负责从网络连接读取消息和向网络连接写入消息
	SetTransmitter(v cellnet.MessageTransmitter)

	// SetHooker 设置事件钩子
	// v: 事件钩子，用于在消息接收后、发送前进行处理
	//    可以在消息处理流程中插入自定义逻辑，如日志记录、消息过滤、加密解密等
	SetHooker(v cellnet.EventHooker)

	// SetCallback 设置用户回调函数
	// v: 事件处理回调函数，当消息到达或系统事件发生时调用
	SetCallback(v cellnet.EventCallback)
}

// NewQueuedEventCallback 创建一个队列化的事件回调函数
// callback: 原始的事件回调函数
// 返回一个新的 EventCallback，保证回调在 Session 的队列中执行，而不是并发执行
// 这样可以确保事件处理的顺序性和线程安全性
func NewQueuedEventCallback(callback cellnet.EventCallback) cellnet.EventCallback {
	return func(ev cellnet.Event) {
		if callback != nil {
			// 将回调放入 Session 对应 Peer 的事件队列中执行
			cellnet.SessionQueuedCall(ev.Session(), func() {
				callback(ev)
			})
		}
	}
}

// MultiHooker 组合多个事件钩子
// 当需要多个 Hooker 时，使用 NewMultiHooker 将多个 hooker 合并成 1 个 hooker 处理
// 钩子会按顺序执行，如果某个钩子返回 nil，则停止后续钩子的执行
type MultiHooker []cellnet.EventHooker

// OnInboundEvent 处理入站（接收）事件
// input: 输入的接收事件
// 返回处理后的输出事件
// 按顺序调用所有钩子的 OnInboundEvent 方法
// 如果某个钩子返回 nil，则停止处理并返回 nil
func (self MultiHooker) OnInboundEvent(input cellnet.Event) (output cellnet.Event) {
	// 按顺序处理所有钩子
	for _, h := range self {
		// 调用钩子的入站事件处理方法
		input = h.OnInboundEvent(input)

		// 如果钩子返回 nil，停止处理
		if input == nil {
			break
		}
	}

	return input
}

// OnOutboundEvent 处理出站（发送）事件
// input: 输入的发送事件
// 返回处理后的输出事件
// 按顺序调用所有钩子的 OnOutboundEvent 方法
// 如果某个钩子返回 nil，则停止处理并返回 nil
func (self MultiHooker) OnOutboundEvent(input cellnet.Event) (output cellnet.Event) {
	// 按顺序处理所有钩子
	for _, h := range self {
		// 调用钩子的出站事件处理方法
		input = h.OnOutboundEvent(input)

		// 如果钩子返回 nil，停止处理
		if input == nil {
			break
		}
	}

	return input
}

// NewMultiHooker 创建组合多个事件钩子的 MultiHooker
// h: 要组合的事件钩子列表
// 返回一个实现了 EventHooker 接口的 MultiHooker
// 钩子会按传入的顺序执行
func NewMultiHooker(h ...cellnet.EventHooker) cellnet.EventHooker {
	return MultiHooker(h)
}
