package udp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/peer"
	"net"
	"sync"
	"time"
)

// DataReader 数据读取接口
// 用于从 Session 中读取原始数据
type DataReader interface {
	// ReadData 读取数据
	// 返回原始数据字节数组
	ReadData() []byte
}

// DataWriter 数据写入接口
// 用于向 Session 写入原始数据
type DataWriter interface {
	// WriteData 写入数据
	// data: 要写入的数据字节数组
	WriteData(data []byte)
}

// udpSession UDP 会话实现
// 表示一个 UDP 连接，负责消息的接收和发送
// 在 Acceptor 中，基于地址管理多个 Session
// 在 Connector 中，通常只有一个 Session
type udpSession struct {
	*peer.CoreProcBundle     // 消息处理组件（编码器、钩子、回调等）
	peer.CoreContextSet      // 上下文数据存储
	peer.CoreSessionIdentify // 会话 ID 管理

	// pInterface 所属的 Peer
	// 用于访问 Peer 的配置和功能
	pInterface cellnet.Peer

	// pkt 当前接收到的数据包
	// 用于实现 DataReader 接口
	pkt []byte

	// remote 远程地址（仅在 Acceptor 中使用）
	// 用于向客户端发送数据包
	remote *net.UDPAddr

	// conn UDP 连接
	// connGuard 连接读写锁，保护 conn 的并发访问
	conn      *net.UDPConn
	connGuard sync.RWMutex

	// timeOutTick 超时时间点
	// 如果当前时间超过此时间点，Session 会被标记为超时
	timeOutTick time.Time

	// key 连接跟踪键（仅在 Acceptor 中使用）
	// 用于在 Session 映射中标识此 Session
	key *connTrackKey
}

// setConn 设置连接
// conn: UDP 连接对象
// 使用写锁保护，确保并发安全
func (self *udpSession) setConn(conn *net.UDPConn) {
	self.connGuard.Lock()
	self.conn = conn
	self.connGuard.Unlock()
}

// Conn 获取连接
// 返回当前的 UDP 连接
// 使用读锁保护，允许多个 goroutine 同时读取
func (self *udpSession) Conn() *net.UDPConn {
	self.connGuard.RLock()
	defer self.connGuard.RUnlock()
	return self.conn
}

// IsAlive 检查 Session 是否仍然存活
// 返回 true 表示 Session 未超时，false 表示已超时
// 用于 Acceptor 中清理超时的 Session
func (self *udpSession) IsAlive() bool {
	return time.Now().Before(self.timeOutTick)
}

// ID 获取会话 ID
// UDP Session 的 ID 始终为 0（UDP 是无连接协议，不需要唯一 ID）
func (self *udpSession) ID() int64 {
	return 0
}

// LocalAddress 获取本地地址
// 返回本地 UDP 地址
func (self *udpSession) LocalAddress() net.Addr {
	return self.Conn().LocalAddr()
}

// Peer 获取所属的 Peer
// 返回会话所属的 Peer 对象
func (self *udpSession) Peer() cellnet.Peer {
	return self.pInterface
}

// Raw 获取原始连接
// 返回 Session 自身（UDP Session 本身就是原始连接）
func (self *udpSession) Raw() interface{} {
	return self
}

// Recv 接收数据包
// data: 接收到的数据包
// 将数据包保存到 pkt，然后解码为消息并分发
func (self *udpSession) Recv(data []byte) {

	// 保存数据包，用于实现 DataReader 接口
	self.pkt = data

	// 解码消息
	msg, err := self.ReadMessage(self)

	// 如果解码成功，发送接收事件
	if msg != nil && err == nil {
		self.ProcEvent(&cellnet.RecvMsgEvent{self, msg})
	}
}

// ReadData 读取数据（实现 DataReader 接口）
// 返回当前接收到的数据包
func (self *udpSession) ReadData() []byte {
	return self.pkt
}

// WriteData 写入数据（实现 DataWriter 接口）
// data: 要写入的数据
// 根据 Session 类型（Connector 或 Acceptor）选择不同的发送方式
func (self *udpSession) WriteData(data []byte) {

	c := self.Conn()
	if c == nil {
		return
	}

	// Connector 中的 Session（remote 为 nil）
	// 直接写入连接，目标地址已在连接时确定
	if self.remote == nil {
		c.Write(data)

		// Acceptor 中的 Session（remote 不为 nil）
		// 需要指定目标地址发送
	} else {
		c.WriteToUDP(data, self.remote)
	}
}

// Send 发送消息
// msg: 要发送的消息对象
// 将消息编码后通过 WriteData 发送
func (self *udpSession) Send(msg interface{}) {

	self.SendMessage(&cellnet.SendMsgEvent{self, msg})
}

// Close 关闭会话
// UDP Session 的关闭是空操作（UDP 是无连接协议）
func (self *udpSession) Close() {

}
