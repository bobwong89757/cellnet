package proc

import (
	"github.com/bobwong89757/cellnet"
	"reflect"
	"sync"
)

// MessageDispatcher 消息派发器，可选组件
// 兼容 v3 以前的注册及派发消息方式
// 在没有代码生成框架及工具时，这是一个较方便的接收处理接口
// 支持根据消息类型自动派发到对应的处理函数
type MessageDispatcher struct {
	// handlerByType 存储消息类型到处理回调函数的映射
	// 键为消息类型（reflect.Type），值为处理回调函数列表
	handlerByType map[reflect.Type][]cellnet.EventCallback

	// handlerByTypeGuard 保护 handlerByType 的读写锁
	// 用于并发安全地访问处理器映射
	handlerByTypeGuard sync.RWMutex
}

// OnEvent 处理事件
// ev: 要处理的事件
// 根据事件中消息的类型，查找对应的处理回调函数并调用
// 如果消息类型未注册，则不处理
func (self *MessageDispatcher) OnEvent(ev cellnet.Event) {
	// 获取消息的类型
	msgType := reflect.TypeOf(ev.Message())

	// 如果消息为 nil，直接返回
	if msgType == nil {
		return
	}

	// 查找该消息类型对应的处理回调函数
	self.handlerByTypeGuard.RLock()
	handlers, ok := self.handlerByType[msgType.Elem()]
	self.handlerByTypeGuard.RUnlock()

	if ok {
		// 调用所有注册的处理回调函数
		for _, callback := range handlers {
			callback(ev)
		}
	}
}

// Exists 检查消息是否已注册处理函数
// msgName: 消息的完整名称，格式为 "包名.类型名"
// 返回 true 表示已注册处理函数，false 表示未注册
func (self *MessageDispatcher) Exists(msgName string) bool {
	// 根据消息名称获取消息元信息
	meta := cellnet.MessageMetaByFullName(msgName)
	if meta == nil {
		return false
	}

	self.handlerByTypeGuard.Lock()
	defer self.handlerByTypeGuard.Unlock()

	// 检查是否有注册的处理函数
	handlers, _ := self.handlerByType[meta.Type]
	return len(handlers) > 0
}

// RegisterMessage 注册消息处理函数
// msgName: 消息的完整名称，格式为 "包名.类型名"
// userCallback: 处理该消息的回调函数
// 如果消息未注册到消息元信息表，会触发 panic
// 支持为同一消息类型注册多个处理函数
func (self *MessageDispatcher) RegisterMessage(msgName string, userCallback cellnet.EventCallback) {
	// 根据消息名称获取消息元信息
	meta := cellnet.MessageMetaByFullName(msgName)
	if meta == nil {
		panic("message not found:" + msgName)
	}

	self.handlerByTypeGuard.Lock()
	// 获取该消息类型已有的处理函数列表
	handlers, _ := self.handlerByType[meta.Type]
	// 追加新的处理函数
	handlers = append(handlers, userCallback)
	// 更新映射表
	self.handlerByType[meta.Type] = handlers
	self.handlerByTypeGuard.Unlock()
}

// NewMessageDispatcher 创建一个新的消息派发器
// 返回初始化好的 MessageDispatcher
func NewMessageDispatcher() *MessageDispatcher {
	return &MessageDispatcher{
		handlerByType: make(map[reflect.Type][]cellnet.EventCallback),
	}
}

// NewMessageDispatcherBindPeer 创建消息派发器并绑定到 Peer
// peer: 要绑定的 Peer
// processorName: 处理器的名称，如 "tcp.ltv"、"udp.ltv"
// 返回创建并绑定好的 MessageDispatcher
// 这是一个便捷函数，创建派发器并自动绑定到 Peer 的处理器
func NewMessageDispatcherBindPeer(peer cellnet.Peer, processorName string) *MessageDispatcher {
	// 创建消息派发器
	self := NewMessageDispatcher()

	// 将派发器的 OnEvent 方法绑定到 Peer 的处理器
	BindProcessorHandler(peer, processorName, self.OnEvent)

	return self
}
