package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"net"
	"sync"
	"time"
)

// tcpConnector
// @Description: TCP连接器
type tcpConnector struct {
	peer.SessionManager

	peer.CorePeerProperty
	peer.CoreContextSet
	peer.CoreRunningTag
	peer.CoreProcBundle
	peer.CoreTCPSocketOption
	peer.CoreCaptureIOPanic

	defaultSes *tcpSession

	tryConnTimes int // 尝试连接次数

	sesEndSignal sync.WaitGroup

	reconDur time.Duration
}

// Start
//
//	@Description: 开始
//	@receiver self
//	@return cellnet.Peer
func (self *tcpConnector) Start() cellnet.Peer {

	self.WaitStopFinished()

	if self.IsRunning() {
		return self
	}

	go self.connect(self.Address())

	return self
}

// Session
//
//	@Description: 获取当前会话
//	@receiver self
//	@return cellnet.Session
func (self *tcpConnector) Session() cellnet.Session {
	return self.defaultSes
}

// SetSessionManager
//
//	@Description: 设置会话管理器
//	@receiver self
//	@param raw
func (self *tcpConnector) SetSessionManager(raw interface{}) {
	self.SessionManager = raw.(peer.SessionManager)
}

// Stop
//
//	@Description: 停止连接器
//	@receiver self
func (self *tcpConnector) Stop() {
	if !self.IsRunning() {
		return
	}

	if self.IsStopping() {
		return
	}

	self.StartStopping()

	// 通知发送关闭
	self.defaultSes.Close()

	// 等待线程结束
	self.WaitStopFinished()

}

// ReconnectDuration
//
//	@Description: 获取重连间隔
//	@receiver self
//	@return time.Duration
func (self *tcpConnector) ReconnectDuration() time.Duration {

	return self.reconDur
}

// SetReconnectDuration
//
//	@Description: 设置重连间隔
//	@receiver self
//	@param v
func (self *tcpConnector) SetReconnectDuration(v time.Duration) {
	self.reconDur = v
}

// Port
//
//	@Description: 连接端口号
//	@receiver self
//	@return int
func (self *tcpConnector) Port() int {

	conn := self.defaultSes.Conn()

	if conn == nil {
		return 0
	}

	return conn.LocalAddr().(*net.TCPAddr).Port
}

const reportConnectFailedLimitTimes = 3

// connect
//
//	@Description: 连接器，传入连接地址和发送封包次数
//	@receiver self
//	@param address
func (self *tcpConnector) connect(address string) {

	self.SetRunning(true)

	for {
		self.tryConnTimes++

		// 尝试用Socket连接地址
		conn, err := net.Dial("tcp", address)

		self.defaultSes.setConn(conn)

		// 发生错误时退出
		if err != nil {

			if self.tryConnTimes <= reportConnectFailedLimitTimes {
				log.GetLog().Error("#tcp.connect failed(%s) %v", self.Name(), err.Error())

				if self.tryConnTimes == reportConnectFailedLimitTimes {
					log.GetLog().Error("(%s) continue reconnecting, but mute log", self.Name())
				}
			}

			// 没重连就退出
			if self.ReconnectDuration() == 0 || self.IsStopping() {

				self.ProcEvent(&cellnet.RecvMsgEvent{
					Ses: self.defaultSes,
					Msg: &cellnet.SessionConnectError{},
				})
				break
			}

			// 有重连就等待
			time.Sleep(self.ReconnectDuration())

			// 继续连接
			continue
		}

		self.sesEndSignal.Add(1)

		self.ApplySocketOption(conn)

		self.defaultSes.Start()

		self.tryConnTimes = 0

		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnected{}})

		self.sesEndSignal.Wait()

		self.defaultSes.setConn(nil)

		// 没重连就退出/主动退出
		if self.IsStopping() || self.ReconnectDuration() == 0 {
			break
		}

		// 有重连就等待
		time.Sleep(self.ReconnectDuration())

		// 继续连接
		continue

	}

	self.SetRunning(false)

	self.EndStopping()
}

// IsReady
//
//	@Description: 是否准备好
//	@receiver self
//	@return bool
func (self *tcpConnector) IsReady() bool {

	return self.SessionCount() != 0
}

// TypeName
//
//	@Description: 获取名字
//	@receiver self
//	@return string
func (self *tcpConnector) TypeName() string {
	return "tcp.Connector"
}

func init() {
	//  注册端
	peer.RegisterPeerCreator(func() cellnet.Peer {
		self := &tcpConnector{
			SessionManager: new(peer.CoreSessionManager),
		}

		self.defaultSes = newSession(nil, self, func() {
			self.sesEndSignal.Done()
		})

		self.CoreTCPSocketOption.Init()

		return self
	})
}
