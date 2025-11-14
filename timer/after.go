package timer

import (
	"github.com/bobwong89757/cellnet"
	"time"
)

// AfterStopper 定义定时器停止接口
// 用于取消已设置的定时器
type AfterStopper interface {
	// Stop 停止定时器
	// 返回 true 表示定时器已停止，false 表示定时器已执行或已停止
	Stop() bool
}

// After 在指定的持续时间后执行回调函数
// q: 事件队列，回调函数会在指定的队列 goroutine 中执行
//    如果 q 为 nil，则直接在当前 goroutine 中执行
// duration: 延迟时间，回调函数会在此时间后执行
// callbackObj: 回调函数对象，支持两种类型：
//   - func(): 无参数的回调函数
//   - func(interface{}): 带一个参数的回调函数，参数为 context
// context: 上下文信息，会传递给带参数的回调函数
// 返回 AfterStopper，可用于取消定时器
func After(q cellnet.EventQueue, duration time.Duration, callbackObj interface{}, context interface{}) AfterStopper {
	// 使用 time.AfterFunc 创建定时器
	return time.AfterFunc(duration, func() {
		switch callback := callbackObj.(type) {
		case func():
			// 无参数的回调函数
			if callback != nil {
				cellnet.QueuedCall(q, callback)
			}

		case func(interface{}):
			// 带参数的回调函数
			if callback != nil {
				cellnet.QueuedCall(q, func() {
					callback(context)
				})
			}
		default:
			// 不支持的回调函数类型
			panic("timer.After: require func() or func(interface{})")
		}
	})
}
