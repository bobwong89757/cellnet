package http

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/proc"
)

// init 包初始化函数
// 自动注册 HTTP 消息处理器
// 当调用 proc.BindProcessorHandler(peer, "http", callback) 时会使用此处理器
func init() {

	proc.RegisterProcessor("http", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {
		// 如果 HTTP 的 peer 有队列，依然会在队列中排队执行
		// 使用队列化的回调以确保事件在正确的 goroutine 中处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

}
