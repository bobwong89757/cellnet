package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"net"
	"time"
)

// tcpSyncConnector
// @Description: TCP同步连接器
type tcpSyncConnector struct {
	peer.SessionManager

	peer.CorePeerProperty
	peer.CoreContextSet
	peer.CoreProcBundle
	peer.CoreTCPSocketOption

	defaultSes *tcpSession
}

// Port
//
//	@Description: 端口号
//	@receiver self
//	@return int
func (self *tcpSyncConnector) Port() int {
	conn := self.defaultSes.Conn()

	if conn == nil {
		return 0
	}

	return conn.LocalAddr().(*net.TCPAddr).Port
}

// Start
//
//	@Description: 开始连接
//	@receiver self
//	@return cellnet.Peer
func (self *tcpSyncConnector) Start() cellnet.Peer {

	// 尝试用Socket连接地址
	conn, err := net.Dial("tcp", self.Address())

	// 发生错误时退出
	if err != nil {

		log.GetLog().Debug("#tcp.connect failed(%s)@%d address: %s", self.Name(), self.defaultSes.ID(), self.Address())

		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnectError{}})
		return self
	}

	self.defaultSes.setConn(conn)

	self.ApplySocketOption(conn)

	self.defaultSes.Start()

	self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnected{}})

	return self
}

// Session
//
//	@Description: 获取会话
//	@receiver self
//	@return cellnet.Session
func (self *tcpSyncConnector) Session() cellnet.Session {
	return self.defaultSes
}

// SetSessionManager
//
//	@Description: 设置会话管理器
//	@receiver self
//	@param raw
func (self *tcpSyncConnector) SetSessionManager(raw interface{}) {
	self.SessionManager = raw.(peer.SessionManager)
}

// ReconnectDuration
//
//	@Description: 获取重连间隔
//	@receiver self
//	@return time.Duration
func (self *tcpSyncConnector) ReconnectDuration() time.Duration {
	return 0
}

// SetReconnectDuration
//
//	@Description: 设置重连间隔
//	@receiver self
//	@param v
func (self *tcpSyncConnector) SetReconnectDuration(v time.Duration) {

}

// Stop
//
//	@Description: 关闭连接
//	@receiver self
func (self *tcpSyncConnector) Stop() {

	if self.defaultSes != nil {
		self.defaultSes.Close()
	}

}

// IsReady
//
//	@Description: 是否准备好
//	@receiver self
//	@return bool
func (self *tcpSyncConnector) IsReady() bool {

	return self.SessionCount() != 0
}

// TypeName
//
//	@Description: 获取类型名字
//	@receiver self
//	@return string
func (self *tcpSyncConnector) TypeName() string {
	return "tcp.SyncConnector"
}

func init() {
	//  注册端
	peer.RegisterPeerCreator(func() cellnet.Peer {
		self := &tcpSyncConnector{
			SessionManager: new(peer.CoreSessionManager),
		}

		self.defaultSes = newSession(nil, self, nil)

		self.CoreTCPSocketOption.Init()

		return self
	})
}
