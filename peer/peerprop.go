package peer

import "github.com/bobwong89757/cellnet"

// CorePeerProperty
// @Description: 核心端口属性
type CorePeerProperty struct {
	// 名字
	name string
	// 队列
	queue cellnet.EventQueue
	// 地址
	addr string
}

// Name
//
//	@Description: 获取通讯端的名称
//	@receiver self
//	@return string
func (self *CorePeerProperty) Name() string {
	return self.name
}

// Queue
//
//	@Description: 获取队列
//	@receiver self
//	@return cellnet.EventQueue
func (self *CorePeerProperty) Queue() cellnet.EventQueue {
	return self.queue
}

// Address
//
//	@Description: 获取SetAddress中的侦听或者连接地址
//	@receiver self
//	@return string
func (self *CorePeerProperty) Address() string {

	return self.addr
}

// SetName
//
//	@Description: 设置通讯端的名称
//	@receiver self
//	@param v
func (self *CorePeerProperty) SetName(v string) {
	self.name = v
}

// SetQueue
//
//	@Description: 设置队列
//	@receiver self
//	@param v
func (self *CorePeerProperty) SetQueue(v cellnet.EventQueue) {
	self.queue = v
}

// SetAddress
//
//	@Description: 设置侦听或者连接地址
//	@receiver self
//	@param v
func (self *CorePeerProperty) SetAddress(v string) {
	self.addr = v
}
