package timer

import (
	"sync/atomic"
	"time"

	"github.com/bobwong89757/cellnet"
)

// Loop 轻量级的持续 Tick 循环
// 按照指定的时间间隔持续调用回调函数
// 支持启动、停止、恢复等操作
type Loop struct {
	// Context 上下文信息
	// 可以存储循环相关的数据
	Context interface{}

	// Duration 循环的时间间隔
	// 每次回调之间的时间间隔
	Duration time.Duration

	// notifyCallback 通知回调函数
	// 每次循环时调用的函数，参数为 Loop 自身
	notifyCallback func(*Loop)

	// running 运行状态标识
	// 使用原子操作保证并发安全
	// 0 表示未运行，非 0 表示正在运行
	running int64

	// Queue 事件队列
	// 回调函数会在指定的队列 goroutine 中执行
	Queue cellnet.EventQueue
}

// Running 检查循环是否正在运行
// 返回 true 表示正在运行，false 表示未运行
func (self *Loop) Running() bool {
	return atomic.LoadInt64(&self.running) != 0
}

// setRunning 设置运行状态
// v: true 表示设置为运行状态，false 表示设置为未运行状态
func (self *Loop) setRunning(v bool) {
	if v {
		atomic.StoreInt64(&self.running, 1)
	} else {
		atomic.StoreInt64(&self.running, 0)
	}
}

// Start 开始循环
// 如果循环已经在运行，返回 false
// 如果成功启动，返回 true
func (self *Loop) Start() bool {
	// 如果已经在运行，返回 false
	if self.Running() {
		return false
	}

	// 设置运行状态
	atomic.StoreInt64(&self.running, 1)

	// 开始第一次循环
	self.rawPost()

	return true
}

// rawPost 内部方法，安排下一次循环
// 如果 Duration 为 0，会触发 panic
// 如果循环正在运行，会在 Duration 时间后调用 tick
func (self *Loop) rawPost() {
	// 检查时间间隔是否有效
	if self.Duration == 0 {
		panic("seconds can be zero in loop")
	}

	// 如果循环正在运行，安排下一次循环
	if self.Running() {
		After(self.Queue, self.Duration, func() {
			tick(self, false)
		}, nil)
	}
}

// NextLoop 立即触发下一次循环
// 将下一次循环立即投递到事件队列中
// 不会等待 Duration 时间
func (self *Loop) NextLoop() {
	self.Queue.Post(func() {
		tick(self, true)
	})
}

// Stop 停止循环
// 停止后，循环不会再触发下一次回调
func (self *Loop) Stop() {
	self.setRunning(false)
}

// Resume 恢复循环
// 恢复后，循环会继续按照 Duration 间隔触发回调
// 注意：恢复后不会立即触发回调，需要等待 Duration 时间
func (self *Loop) Resume() {
	self.setRunning(true)
}

// Notify 立即调用一次用户回调
// 不会等待 Duration 时间，立即执行回调函数
// 返回自身以便链式调用
func (self *Loop) Notify() *Loop {
	self.notifyCallback(self)
	return self
}

// SetNotifyFunc 设置通知回调函数
// notifyCallback: 每次循环时调用的函数
// 返回自身以便链式调用
func (self *Loop) SetNotifyFunc(notifyCallback func(*Loop)) *Loop {
	self.notifyCallback = notifyCallback
	return self
}

// NotifyFunc 获取通知回调函数
// 返回当前设置的通知回调函数
func (self *Loop) NotifyFunc() func(*Loop) {
	return self.notifyCallback
}

// tick 循环的 tick 函数
// ctx: Loop 实例
// nextLoop: 是否为立即触发的下一次循环
// 如果 nextLoop 为 false 且循环正在运行，会安排下一次循环
// 即使回调函数中发生 panic，也会继续循环
func tick(ctx interface{}, nextLoop bool) {
	loop := ctx.(*Loop)

	// 如果不是立即触发的循环，且循环正在运行，安排下一次循环
	if !nextLoop && loop.Running() {
		// 使用 defer 确保即使回调函数中发生 panic，也会继续循环
		defer loop.rawPost()
	}

	// 调用用户回调
	loop.Notify()
}

// NewLoop 创建一个新的循环定时器
// q: 事件队列，回调函数会在指定的队列 goroutine 中执行
// duration: 循环的时间间隔
// notifyCallback: 每次循环时调用的函数
// context: 上下文信息，可以存储循环相关的数据
// 返回初始化好的 Loop，需要调用 Start() 方法开始循环
func NewLoop(q cellnet.EventQueue, duration time.Duration, notifyCallback func(*Loop), context interface{}) *Loop {
	self := &Loop{
		Context:        context,
		Duration:       duration,
		notifyCallback: notifyCallback,
		Queue:          q,
	}

	return self
}
