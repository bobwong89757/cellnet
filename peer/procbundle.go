package peer

import (
	"errors"

	"github.com/bobwong89757/cellnet"
)

// MessagePoster 定义手动消息投递接口
// 用于手动投递消息到处理流程
// 消息会经过 Hooker 处理后再到达用户回调
type MessagePoster interface {
	// ProcEvent 投递一个事件到处理流程
	// ev: 要处理的事件
	// 事件会先经过 Hooker 处理，然后到达用户回调
	ProcEvent(ev cellnet.Event)
}

// CoreProcBundle 提供消息处理资源包的核心实现
// 包含消息传输器、事件钩子和用户回调
// 所有 Peer 实现都可以嵌入此结构体来获得消息处理功能
type CoreProcBundle struct {
	// transmit 消息传输器
	// 负责从网络连接读取消息和向网络连接写入消息
	transmit cellnet.MessageTransmitter

	// hooker 事件钩子
	// 用于在消息处理流程中插入自定义逻辑
	hooker cellnet.EventHooker

	// callback 事件回调函数
	// 当消息到达或系统事件发生时调用
	callback cellnet.EventCallback
}

// GetBundle 获取核心协议包
// 返回自身，用于类型转换
func (self *CoreProcBundle) GetBundle() *CoreProcBundle {
	return self
}

// SetTransmitter 设置消息传输器
// v: 消息传输器，负责从网络连接读取消息和向网络连接写入消息
func (self *CoreProcBundle) SetTransmitter(v cellnet.MessageTransmitter) {
	self.transmit = v
}

// SetHooker 设置事件钩子
// v: 事件钩子，用于在消息处理流程中插入自定义逻辑
func (self *CoreProcBundle) SetHooker(v cellnet.EventHooker) {
	self.hooker = v
}

// SetCallback 设置事件回调函数
// v: 事件处理回调函数，当消息到达或系统事件发生时调用
func (self *CoreProcBundle) SetCallback(v cellnet.EventCallback) {
	self.callback = v
}

// notHandled 表示传输器未设置的错误
var notHandled = errors.New("Processor: Transimitter nil")

// ReadMessage 从 Session 读取消息
// ses: 要读取消息的 Session
// 返回接收到的消息对象和错误信息
// 如果传输器未设置，返回错误
func (self *CoreProcBundle) ReadMessage(ses cellnet.Session) (msg interface{}, err error) {
	if self.transmit != nil {
		// 使用传输器读取消息
		return self.transmit.OnRecvMessage(ses)
	}

	return nil, notHandled
}

// SendMessage 向 Session 发送消息
// ev: 要发送的事件
// 消息会先经过 Hooker 处理，然后由传输器发送
// 如果 Hooker 返回 nil，则停止发送
func (self *CoreProcBundle) SendMessage(ev cellnet.Event) {
	// 如果有钩子，先处理出站事件
	if self.hooker != nil {
		ev = self.hooker.OnOutboundEvent(ev)
	}

	// 如果传输器已设置且事件不为 nil，发送消息
	if self.transmit != nil && ev != nil {
		self.transmit.OnSendMessage(ev.Session(), ev.Message())
	}
}

// ProcEvent 处理事件
// ev: 要处理的事件
// 事件会先经过 Hooker 处理，然后到达用户回调
// 如果 Hooker 返回 nil，则停止处理
func (self *CoreProcBundle) ProcEvent(ev cellnet.Event) {
	// 如果有钩子，先处理入站事件
	if self.hooker != nil {
		ev = self.hooker.OnInboundEvent(ev)
	}

	// 如果有回调函数且事件不为 nil，调用回调
	if self.callback != nil && ev != nil {
		self.callback(ev)
	}
}
