package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/proc"
)

// init 包初始化函数
// 自动注册 TCP LTV（Length-Type-Value）消息处理器
// 当调用 proc.BindProcessorHandler(peer, "tcp.ltv", callback) 时会使用此处理器
func init() {
	// 注册消息处理器为 tcp.ltv
	proc.RegisterProcessor("tcp.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		// 设置消息传输器，负责消息的编码、解码和网络传输
		bundle.SetTransmitter(new(TCPMessageTransmitter))
		// 设置事件钩子，用于拦截和处理事件
		bundle.SetHooker(new(MsgHooker))
		// 设置事件回调，使用队列化的回调以确保事件在正确的 goroutine 中处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))

	})
}
