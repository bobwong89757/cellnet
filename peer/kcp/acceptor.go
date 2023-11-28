package kcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/cellnet/util"
	"github.com/bobwong89757/kcp-go/v6"
	"net"
	"strings"
	"time"
)

const MaxUDPRecvBuffer = 4096

type kcpAcceptor struct {
	peer.SessionManager
	peer.CorePeerProperty
	peer.CoreContextSet
	peer.CoreRunningTag
	peer.CoreProcBundle
	peer.CoreCaptureIOPanic

	conn *net.UDPConn

	listener *kcp.Listener

	sesTimeout       time.Duration
	sesCleanTimeout  time.Duration
	sesCleanLastTime time.Time

	sesByConnTrack map[connTrackKey]*KcpSession
}

func (self *kcpAcceptor) IsReady() bool {

	return self.IsRunning()
}

func (self *kcpAcceptor) Port() int {
	if self.conn == nil {
		return 0
	}

	return self.conn.LocalAddr().(*net.UDPAddr).Port
}

func (self *kcpAcceptor) Start() cellnet.Peer {

	self.WaitStopFinished()
	if self.IsRunning() {
		return self
	}

	var finalAddr *util.Address
	ln, err := util.DetectPort(self.Address(), func(a *util.Address, port int) (interface{}, error) {

		addr, err := net.ResolveUDPAddr("udp", a.HostPortString(port))
		if err != nil {
			return nil, err
		}

		finalAddr = a

		self.listener, err = kcp.ListenWithOptions(addr.String(), nil, 0, 0)
		if err != nil {
			log.GetLog().Errorf("#kcp.listen failed(%s) %v", self.Name(), err.Error())
			return nil, err
		}
		//return net.ListenUDP("udp", addr)
		return self.listener.GetConn(), nil

	})

	if err != nil {

		log.GetLog().Errorf("#kcp.listen failed(%s) %v", self.Name(), err.Error())
		return self
	}

	self.conn = ln.(*net.UDPConn)

	log.GetLog().Infof("#kcp.listen(%s) %s", self.Name(), finalAddr.String(self.Port()))

	go self.accept()
	return self
}

//func (self *kcpAcceptor) protectedRecvPacket(ses *KcpSession, data []byte) {
//	defer func() {
//
//		if err := recover(); err != nil {
//			log.GetLog().Errorf("IO panic: %s", err)
//			self.conn.Close()
//		}
//
//	}()
//
//	ses.Recv(data)
//}

func (self *kcpAcceptor) ListenAddress() string {

	pos := strings.Index(self.Address(), ":")
	if pos == -1 {
		return self.Address()
	}

	host := self.Address()[:pos]

	return util.JoinAddress(host, self.Port())
}

func (self *kcpAcceptor) accept() {

	self.SetRunning(true)

	//recvBuff := make([]byte, MaxUDPRecvBuffer)

	for self.IsRunning() {
		udpSession, err := self.listener.AcceptKCP()

		if self.IsStopping() {
			break
		}

		if err == nil {
			// 处理连接进入独立线程, 防止accept无法响应
			go self.onNewSession(udpSession)

		} else {

			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}

			// 调试状态时, 才打出accept的具体错误
			log.GetLog().Errorf("#kcp.accept failed(%s) %v", self.Name(), err.Error())
			break
		}

		//if err != nil {
		//	break
		//}
		//
		self.checkTimeoutSession()
		//
		//ses := self.getSession(udpSession.RemoteAddr().(*net.UDPAddr))
		//ses.KcpSession = udpSession
		//self.ProcEvent(&cellnet.RecvMsgEvent{
		//	Ses: ses,
		//	Msg: &cellnet.SessionAccepted{},
		//})
		//
		//ses.Recv(recvBuff)
	}

	self.SetRunning(false)
	self.EndStopping()
}

func (self *kcpAcceptor) onNewSession(kcpSession *kcp.UDPSession) {

	self.getSession(kcpSession.RemoteAddr().(*net.UDPAddr), kcpSession)
}

// 检查超时session
func (self *kcpAcceptor) checkTimeoutSession() {
	now := time.Now()

	// 定时清理超时的session
	if now.After(self.sesCleanLastTime.Add(self.sesCleanTimeout)) {
		sesToDelete := make([]*KcpSession, 0, 10)
		for _, ses := range self.sesByConnTrack {
			if !ses.IsAlive() {
				sesToDelete = append(sesToDelete, ses)
			}
		}

		for _, ses := range sesToDelete {
			delete(self.sesByConnTrack, *ses.key)
		}

		self.sesCleanLastTime = now
	}
}

func (self *kcpAcceptor) getSession(addr *net.UDPAddr, kcpSession *kcp.UDPSession) *KcpSession {

	key := newConnTrackKey(addr)

	ses := self.sesByConnTrack[*key]

	if ses == nil {
		ses = newSession(kcpSession, self, nil)
		ses.key = key
		ses.Start()
		self.ProcEvent(&cellnet.RecvMsgEvent{
			Ses: ses,
			Msg: &cellnet.SessionAccepted{},
		})
	} else {
		ses.pInterface = self
	}

	self.sesByConnTrack[*key] = ses

	// 续租
	ses.timeOutTick = time.Now().Add(self.sesTimeout)

	return ses
}

func (self *kcpAcceptor) SetSessionTTL(dur time.Duration) {
	self.sesTimeout = dur
}

func (self *kcpAcceptor) SetSessionCleanTimeout(dur time.Duration) {
	self.sesCleanTimeout = dur
}

func (self *kcpAcceptor) Stop() {

	if !self.IsRunning() {
		return
	}

	if self.IsStopping() {
		return
	}

	self.StartStopping()

	self.listener.Close()

	self.CloseAllSession()

	//// TODO 等待accept线程结束
	//self.SetRunning(false)
	self.WaitStopFinished()
}

func (self *kcpAcceptor) TypeName() string {
	return "kcp.Acceptor"
}

func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &kcpAcceptor{
			SessionManager:   new(peer.CoreSessionManager),
			sesTimeout:       time.Minute,
			sesCleanTimeout:  time.Minute,
			sesCleanLastTime: time.Now(),
			sesByConnTrack:   make(map[connTrackKey]*KcpSession),
		}

		return p
	})
}
