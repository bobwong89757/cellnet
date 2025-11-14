package udp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"net"
)

// udpConnector UDP 连接器实现
// 用于创建 UDP 客户端，连接到服务器
// UDP 是无连接协议，连接器维护一个默认的 Session
type udpConnector struct {
	peer.CoreSessionManager // 会话管理器
	peer.CorePeerProperty   // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreContextSet     // 上下文数据存储
	peer.CoreRunningTag     // 运行状态标记
	peer.CoreProcBundle     // 消息处理组件（编码器、钩子、回调等）

	// remoteAddr 远程服务器地址
	// 连接器会向此地址发送数据包
	remoteAddr *net.UDPAddr

	// defaultSes 默认会话
	// UDP 连接器通常只有一个 Session
	defaultSes *udpSession
}

// Start 开始连接
// 解析服务器地址，并在后台 goroutine 中执行连接逻辑
// 返回自身以支持链式调用
func (self *udpConnector) Start() cellnet.Peer {

	var err error
	// 解析 UDP 地址
	self.remoteAddr, err = net.ResolveUDPAddr("udp", self.Address())

	if err != nil {
		// 地址解析失败，记录错误
		log.GetLog().Errorf("#resolve udp address failed(%s) %v", self.Name(), err.Error())
		return self
	}

	// 在后台 goroutine 中执行连接
	go self.connect()

	return self
}

// Session 获取当前会话
// 返回连接器的默认会话
// 如果未连接，会话可能为 nil
func (self *udpConnector) Session() cellnet.Session {
	return self.defaultSes
}

// IsReady 检查连接器是否已准备好
// 返回 true 表示已成功连接到服务器
// 返回 false 表示未连接
func (self *udpConnector) IsReady() bool {

	return self.defaultSes.Conn() != nil
}

// connect 连接循环
// 在后台 goroutine 中运行，连接到服务器并持续接收数据包
// UDP 是无连接协议，连接实际上是创建了一个 UDP socket
func (self *udpConnector) connect() {

	// 创建 UDP 连接
	conn, err := net.DialUDP("udp", nil, self.remoteAddr)
	if err != nil {
		// 连接失败，记录错误
		log.GetLog().Errorf("#udp.connect failed(%s) %v", self.Name(), err.Error())
		return
	}

	// 设置会话的连接
	self.defaultSes.setConn(conn)

	ses := self.defaultSes

	// 发送连接成功事件
	self.ProcEvent(&cellnet.RecvMsgEvent{ses, &cellnet.SessionConnected{}})

	// 创建接收缓冲区
	recvBuff := make([]byte, MaxUDPRecvBuffer)

	self.SetRunning(true)

	// 持续接收数据包
	for self.IsRunning() {

		n, _, err := conn.ReadFromUDP(recvBuff)
		if err != nil {
			// 读取错误，退出循环
			break
		}

		if n > 0 {
			// 处理接收到的数据包
			ses.Recv(recvBuff[:n])
		}

	}
}

// Stop 停止连接器
// 停止接收循环，关闭 UDP 连接
func (self *udpConnector) Stop() {

	self.SetRunning(false)

	// 关闭连接
	if c := self.defaultSes.Conn(); c != nil {
		c.Close()
	}
}

// TypeName 返回连接器的类型名称
// 用于标识和日志记录
func (self *udpConnector) TypeName() string {
	return "udp.Connector"
}

// init 包初始化函数
// 自动注册 UDP 连接器的创建函数
// 当调用 cellnet.NewPeer("udp.Connector", ...) 时会使用此函数创建实例
func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &udpConnector{}

		// 创建默认会话
		p.defaultSes = &udpSession{
			pInterface:     p,
			CoreProcBundle: &p.CoreProcBundle,
		}

		return p
	})
}
