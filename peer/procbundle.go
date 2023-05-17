package peer

import (
	"errors"
	"github.com/bobwong89757/cellnet"
)

// MessagePoster
// @Description: 手动消息投递
type MessagePoster interface {
	// 投递一个消息到Hooker之前
	ProcEvent(ev cellnet.Event)
}

// CoreProcBundle
// @Description: 消息处理资源包
type CoreProcBundle struct {
	// 消息传输器
	transmit cellnet.MessageTransmitter
	// 事件截获
	hooker cellnet.EventHooker
	// 事件回调
	callback cellnet.EventCallback
}

// GetBundle
//
//	@Description: 获取核心协议包
//	@receiver self
//	@return *CoreProcBundle
func (self *CoreProcBundle) GetBundle() *CoreProcBundle {
	return self
}

// SetTransmitter
//
//	@Description: 设置消息传输器
//	@receiver self
//	@param v
func (self *CoreProcBundle) SetTransmitter(v cellnet.MessageTransmitter) {
	self.transmit = v
}

// SetHooker
//
//	@Description: 设置钩子
//	@receiver self
//	@param v
func (self *CoreProcBundle) SetHooker(v cellnet.EventHooker) {
	self.hooker = v
}

// SetCallback
//
//	@Description: 设置回调
//	@receiver self
//	@param v
func (self *CoreProcBundle) SetCallback(v cellnet.EventCallback) {
	self.callback = v
}

var notHandled = errors.New("Processor: Transimitter nil")

// ReadMessage
//
//	@Description: 读取消息
//	@receiver self
//	@param ses
//	@return msg
//	@return err
func (self *CoreProcBundle) ReadMessage(ses cellnet.Session) (msg interface{}, err error) {

	if self.transmit != nil {
		return self.transmit.OnRecvMessage(ses)
	}

	return nil, notHandled
}

// SendMessage
//
//	@Description: 发送消息
//	@receiver self
//	@param ev
func (self *CoreProcBundle) SendMessage(ev cellnet.Event) {

	if self.hooker != nil {
		ev = self.hooker.OnOutboundEvent(ev)
	}

	if self.transmit != nil && ev != nil {
		self.transmit.OnSendMessage(ev.Session(), ev.Message())
	}
}

// ProcEvent
//
//	@Description: 处理事件
//	@receiver self
//	@param ev
func (self *CoreProcBundle) ProcEvent(ev cellnet.Event) {

	if self.hooker != nil {
		ev = self.hooker.OnInboundEvent(ev)
	}

	if self.callback != nil && ev != nil {
		self.callback(ev)
	}
}
