package tcp

import (
	"net"
	"strings"
	"time"

	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/cellnet/util"
)

// tcpAcceptor
// @Description: TCP消息接收器
type tcpAcceptor struct {
	peer.SessionManager
	peer.CorePeerProperty
	peer.CoreContextSet
	peer.CoreRunningTag
	peer.CoreProcBundle
	peer.CoreTCPSocketOption
	peer.CoreCaptureIOPanic

	// 保存侦听器
	listener net.Listener
}

// Port
//
//	@Description: 获取端口
//	@receiver self
//	@return int
func (self *tcpAcceptor) Port() int {
	if self.listener == nil {
		return 0
	}

	return self.listener.Addr().(*net.TCPAddr).Port
}

// IsReady
//
//	@Description: 检查是否准备好
//	@receiver self
//	@return bool
func (self *tcpAcceptor) IsReady() bool {

	return self.IsRunning()
}

// Start
//
//	@Description: 异步开始侦听
//	@receiver self
//	@return cellnet.Peer
func (self *tcpAcceptor) Start() cellnet.Peer {

	self.WaitStopFinished()

	if self.IsRunning() {
		return self
	}

	ln, err := util.DetectPort(self.Address(), func(a *util.Address, port int) (interface{}, error) {
		return net.Listen("tcp", a.HostPortString(port))
	})

	if err != nil {

		log.GetLog().Errorf("#tcp.listen failed(%s) %v", self.Name(), err.Error())

		self.SetRunning(false)

		return self
	}

	self.listener = ln.(net.Listener)

	log.GetLog().Infof("#tcp.listen(%s) %s", self.Name(), self.ListenAddress())

	go self.accept()

	return self
}

// ListenAddress
//
//	@Description: 监听地址
//	@receiver self
//	@return string
func (self *tcpAcceptor) ListenAddress() string {

	pos := strings.Index(self.Address(), ":")
	if pos == -1 {
		return self.Address()
	}

	host := self.Address()[:pos]

	return util.JoinAddress(host, self.Port())
}

// accept
//
//	@Description: 开始侦听
//	@receiver self
func (self *tcpAcceptor) accept() {
	self.SetRunning(true)

	for {
		conn, err := self.listener.Accept()

		if self.IsStopping() {
			break
		}

		if err == nil {
			// 处理连接进入独立线程, 防止accept无法响应
			go self.onNewSession(conn)

		} else {

			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}

			// 调试状态时, 才打出accept的具体错误
			log.GetLog().Errorf("#tcp.accept failed(%s) %v", self.Name(), err.Error())
			break
		}
	}

	self.SetRunning(false)

	self.EndStopping()

}

// onNewSession
//
//	@Description: 新会话钩子
//	@receiver self
//	@param conn
func (self *tcpAcceptor) onNewSession(conn net.Conn) {

	self.ApplySocketOption(conn)

	ses := newSession(conn, self, nil)

	ses.Start()

	self.ProcEvent(&cellnet.RecvMsgEvent{
		Ses: ses,
		Msg: &cellnet.SessionAccepted{},
	})
}

// Stop
//
//	@Description: 停止侦听器
//	@receiver self
func (self *tcpAcceptor) Stop() {
	if !self.IsRunning() {
		return
	}

	if self.IsStopping() {
		return
	}

	self.StartStopping()

	self.listener.Close()

	// 断开所有连接
	self.CloseAllSession()

	// 等待线程结束
	self.WaitStopFinished()
}

// TypeName
//
//	@Description: 名称
//	@receiver self
//	@return string
func (self *tcpAcceptor) TypeName() string {
	return "tcp.Acceptor"
}

// init
//
//	@Description: 自动执行的init
func init() {
	//  注册端
	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &tcpAcceptor{
			SessionManager: new(peer.CoreSessionManager),
		}

		p.CoreTCPSocketOption.Init()

		return p
	})
}
