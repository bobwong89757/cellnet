package peer

import "github.com/bobwong89757/cellnet"

// CorePeerProperty 提供 Peer 基础属性的核心实现
// 包括名称、地址和事件队列等基本属性
// 所有 Peer 实现都可以嵌入此结构体来获得基础属性功能
type CorePeerProperty struct {
	// name Peer 的名称
	// 用于标识和日志记录
	name string

	// queue Peer 关联的事件队列
	// 用于异步处理网络事件
	queue cellnet.EventQueue

	// addr Peer 的地址
	// 对于 Acceptor，是监听地址（如 "127.0.0.1:8080"）
	// 对于 Connector，是连接地址
	addr string
}

// Name 获取 Peer 的名称
// 返回 Peer 的名称，用于标识和日志记录
func (self *CorePeerProperty) Name() string {
	return self.name
}

// Queue 获取 Peer 关联的事件队列
// 返回事件队列，用于异步处理网络事件
func (self *CorePeerProperty) Queue() cellnet.EventQueue {
	return self.queue
}

// Address 获取 Peer 的地址
// 返回在 SetAddress 中设置的侦听或连接地址
// 对于 Acceptor，返回监听地址
// 对于 Connector，返回连接地址
func (self *CorePeerProperty) Address() string {
	return self.addr
}

// SetName 设置 Peer 的名称
// v: Peer 的名称，用于标识和日志记录
func (self *CorePeerProperty) SetName(v string) {
	self.name = v
}

// SetQueue 设置 Peer 关联的事件队列
// v: 事件队列，用于异步处理网络事件
func (self *CorePeerProperty) SetQueue(v cellnet.EventQueue) {
	self.queue = v
}

// SetAddress 设置 Peer 的地址
// v: 地址字符串
// 对于 Acceptor，设置监听地址（如 "127.0.0.1:8080"）
// 对于 Connector，设置连接地址
func (self *CorePeerProperty) SetAddress(v string) {
	self.addr = v
}
