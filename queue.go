package cellnet

import (
	"fmt"
	"github.com/bobwong89757/cellnet/log"
	"runtime/debug"
	"sync"
	"time"
)

// EventQueue 定义事件队列接口
// 事件队列是 cellnet 事件驱动模型的核心，用于异步处理网络事件
// 所有网络事件（消息接收、连接建立、连接断开等）都会通过事件队列处理
//
// 事件队列支持单线程异步、多线程同步、多线程并发等多种处理模型
type EventQueue interface {
	// StartLoop 启动事件循环
	// 在独立的 goroutine 中开始处理事件队列中的事件
	// 返回自身以便链式调用
	StartLoop() EventQueue

	// StopLoop 停止事件循环
	// 向队列发送停止信号，事件循环会在处理完当前事件后退出
	// 返回自身以便链式调用
	StopLoop() EventQueue

	// Wait 等待事件循环退出
	// 阻塞当前 goroutine，直到事件循环完全退出
	// 通常与 StopLoop() 配合使用，确保优雅关闭
	Wait()

	// Post 投递事件到队列
	// callback: 要执行的回调函数
	// 事件会异步处理，不会阻塞调用者
	Post(callback func())

	// EnableCapturePanic 启用或禁用异常捕获
	// v: true 表示启用异常捕获，false 表示禁用
	// 启用后，如果事件处理中发生 panic，会被捕获并通知，避免程序崩溃
	EnableCapturePanic(v bool)

	// Count 返回队列中当前事件的数量
	// 返回待处理的事件数量
	Count() int
}

// CapturePanicNotifyFunc 是 panic 捕获通知函数的类型
// 当事件处理中发生 panic 时，会调用此函数进行通知
// raw: panic 的值
// queue: 发生 panic 的事件队列
type CapturePanicNotifyFunc func(interface{}, EventQueue)

// eventQueue 是 EventQueue 接口的默认实现
// 基于 Pipe 实现，提供线程安全的事件队列功能
type eventQueue struct {
	// Pipe 底层的数据管道，用于存储和传递事件
	*Pipe

	// endSignal 用于等待事件循环退出的同步信号
	endSignal sync.WaitGroup

	// capturePanic 是否启用异常捕获
	capturePanic bool

	// onPanic panic 捕获通知函数
	// 当事件处理中发生 panic 时，会调用此函数
	onPanic CapturePanicNotifyFunc
}

// EnableCapturePanic 启用或禁用异常捕获
// v: true 表示启用异常捕获，false 表示禁用
// 启用后，事件处理中的 panic 会被捕获，避免程序崩溃
func (self *eventQueue) EnableCapturePanic(v bool) {
	self.capturePanic = v
}

// SetCapturePanicNotify 设置 panic 捕获通知函数
// callback: 当发生 panic 时调用的通知函数
// 可以自定义 panic 处理逻辑，如记录日志、发送告警等
func (self *eventQueue) SetCapturePanicNotify(callback CapturePanicNotifyFunc) {
	self.onPanic = callback
}

// Post 投递事件到队列
// callback: 要执行的回调函数
// 如果 callback 为 nil，则忽略
// 事件会异步处理，不会阻塞调用者
func (self *eventQueue) Post(callback func()) {
	if callback == nil {
		return
	}

	// 将回调函数添加到队列
	self.Add(callback)
}

// protectedCall 保护调用用户函数
// callback: 要执行的回调函数
// 如果启用了异常捕获，会捕获 panic 并通知
// 如果未启用异常捕获，直接调用函数，panic 会向上传播
func (self *eventQueue) protectedCall(callback func()) {
	if self.capturePanic {
		// 启用异常捕获，使用 defer recover 捕获 panic
		defer func() {
			if err := recover(); err != nil {
				// 调用通知函数处理 panic
				self.onPanic(err, self)
			}
		}()
	}

	// 执行回调函数
	callback()
}

// StartLoop 启动事件循环
// 在独立的 goroutine 中开始处理事件队列中的事件
// 事件循环会持续运行，直到收到停止信号
// 返回自身以便链式调用
func (self *eventQueue) StartLoop() EventQueue {
	// 增加等待计数，用于 Wait() 方法等待退出
	self.endSignal.Add(1)

	// 在独立的 goroutine 中运行事件循环
	go func() {
		var writeList []interface{}

		// 事件循环主循环
		for {
			// 清空列表，复用切片
			writeList = writeList[0:0]
			// 从队列中取出所有待处理的事件
			exit := self.Pick(&writeList)

			// 遍历处理所有事件
			for _, msg := range writeList {
				switch t := msg.(type) {
				case func():
					// 如果是函数类型，执行回调
					self.protectedCall(t)
				case nil:
					// nil 表示退出信号，跳出循环
					break
				default:
					// 其他类型，记录警告日志
					log.GetLog().Warnf("unexpected type %T", t)
				}
			}

			// 如果收到退出信号，退出循环
			if exit {
				break
			}
		}

		// 通知等待者事件循环已退出
		self.endSignal.Done()
	}()

	return self
}

// StopLoop 停止事件循环
// 向队列发送 nil 作为退出信号
// 事件循环会在处理完当前事件后退出
// 返回自身以便链式调用
func (self *eventQueue) StopLoop() EventQueue {
	// 添加 nil 作为退出信号
	self.Add(nil)
	return self
}

// Wait 等待事件循环退出
// 阻塞当前 goroutine，直到事件循环完全退出
// 通常与 StopLoop() 配合使用，确保优雅关闭
func (self *eventQueue) Wait() {
	self.endSignal.Wait()
}

// NewEventQueue 创建一个新的事件队列
// 返回初始化好的 EventQueue，包含默认的 panic 处理
// 默认启用异常捕获，panic 时会打印堆栈信息
func NewEventQueue() EventQueue {
	return &eventQueue{
		Pipe: NewPipe(),

		// 默认的 panic 处理：打印时间和堆栈信息
		onPanic: func(raw interface{}, queue EventQueue) {
			fmt.Printf("%s: %v \n%s\n", time.Now().Format("2006-01-02 15:04:05"), raw, string(debug.Stack()))
			debug.PrintStack()
		},
	}
}

// SessionQueuedCall 在会话对应的 Peer 的事件队列中执行回调
// ses: 会话对象
// callback: 要执行的回调函数
// 如果 Session 为 nil，则不执行
// 如果 Peer 有事件队列，则在队列中执行；如果没有队列，则立即执行
func SessionQueuedCall(ses Session, callback func()) {
	if ses == nil {
		return
	}
	// 获取 Peer 的事件队列
	q := ses.Peer().(interface {
		Queue() EventQueue
	}).Queue()

	QueuedCall(q, callback)
}

// QueuedCall 根据是否有队列决定执行方式
// queue: 事件队列，可以为 nil
// callback: 要执行的回调函数
// 如果 queue 为 nil，立即执行 callback
// 如果 queue 不为 nil，将 callback 投递到队列中异步执行
func QueuedCall(queue EventQueue, callback func()) {
	if queue == nil {
		// 没有队列，立即执行
		callback()
	} else {
		// 有队列，投递到队列中
		queue.Post(callback)
	}
}
