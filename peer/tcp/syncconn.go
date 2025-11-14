package tcp

import (
	"net"
	"time"

	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
)

// tcpSyncConnector TCP 同步连接器实现
// 用于创建 TCP 客户端，同步连接到服务器
// 与普通连接器不同，同步连接器在 Start() 时同步等待连接完成，不支持自动重连
type tcpSyncConnector struct {
	peer.SessionManager      // 会话管理器
	peer.CorePeerProperty    // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreContextSet      // 上下文数据存储
	peer.CoreProcBundle      // 消息处理组件（编码器、钩子、回调等）
	peer.CoreTCPSocketOption // TCP Socket 选项（缓冲区、超时等）

	// defaultSes 默认会话
	// 连接器通常只有一个会话
	defaultSes *tcpSession
}

// Port 获取本地端口号
// 返回当前连接使用的本地端口号
// 如果未连接，返回 0
func (self *tcpSyncConnector) Port() int {
	conn := self.defaultSes.Conn()

	if conn == nil {
		return 0
	}

	return conn.LocalAddr().(*net.TCPAddr).Port
}

// Start 同步开始连接
// 同步等待连接完成，连接失败时发送 SessionConnectError 事件
// 连接成功后启动会话并发送 SessionConnected 事件
// 不支持自动重连
// 返回自身以支持链式调用
func (self *tcpSyncConnector) Start() cellnet.Peer {

	// 尝试用 Socket 连接地址（同步阻塞）
	conn, err := net.Dial("tcp", self.Address())

	// 发生错误时退出
	if err != nil {
		// 记录连接失败日志
		log.GetLog().Debugf("#tcp.connect failed(%s)@%d address: %s", self.Name(), self.defaultSes.ID(), self.Address())

		// 发送连接错误事件
		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnectError{}})
		return self
	}

	// 设置连接
	self.defaultSes.setConn(conn)

	// 应用 Socket 选项
	self.ApplySocketOption(conn)

	// 启动会话（启动接收和发送循环）
	self.defaultSes.Start()

	// 发送连接成功事件
	self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnected{}})

	return self
}

// Session 获取当前会话
// 返回连接器的默认会话
// 如果未连接，会话可能为 nil
func (self *tcpSyncConnector) Session() cellnet.Session {
	return self.defaultSes
}

// SetSessionManager 设置会话管理器
// raw: 实现 peer.SessionManager 接口的对象
// 用于自定义会话管理逻辑
func (self *tcpSyncConnector) SetSessionManager(raw interface{}) {
	self.SessionManager = raw.(peer.SessionManager)
}

// ReconnectDuration 获取重连时间间隔
// 同步连接器不支持重连，始终返回 0
func (self *tcpSyncConnector) ReconnectDuration() time.Duration {
	return 0
}

// SetReconnectDuration 设置重连时间间隔
// v: 重连时间间隔
// 同步连接器不支持重连，此方法为空操作
func (self *tcpSyncConnector) SetReconnectDuration(v time.Duration) {

}

// Stop 停止连接器
// 关闭会话，断开连接
func (self *tcpSyncConnector) Stop() {

	if self.defaultSes != nil {
		self.defaultSes.Close()
	}

}

// IsReady 检查连接器是否已准备好
// 返回 true 表示已成功连接到服务器
// 返回 false 表示未连接
func (self *tcpSyncConnector) IsReady() bool {

	return self.SessionCount() != 0
}

// TypeName 返回连接器的类型名称
// 用于标识和日志记录
func (self *tcpSyncConnector) TypeName() string {
	return "tcp.SyncConnector"
}

// init 包初始化函数
// 自动注册 TCP 同步连接器的创建函数
// 当调用 cellnet.NewPeer("tcp.SyncConnector", ...) 时会使用此函数创建实例
func init() {
	// 注册 Peer 创建函数
	peer.RegisterPeerCreator(func() cellnet.Peer {
		self := &tcpSyncConnector{
			SessionManager: new(peer.CoreSessionManager),
		}

		// 创建默认会话
		self.defaultSes = newSession(nil, self, nil)

		// 初始化 TCP Socket 选项
		self.CoreTCPSocketOption.Init()

		return self
	})
}
