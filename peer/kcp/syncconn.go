package kcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/kcp-go/v6"
	"net"
	"time"
)

type kcpSyncConnector struct {
	peer.SessionManager

	peer.CorePeerProperty
	peer.CoreContextSet
	peer.CoreProcBundle
	peer.CoreTCPSocketOption

	defaultSes *KcpSession
}

func (self *kcpSyncConnector) Port() int {
	conn := self.defaultSes.GetKcpSession().GetConn()

	if conn == nil {
		return 0
	}

	return conn.LocalAddr().(*net.TCPAddr).Port
}

func (self *kcpSyncConnector) Start() cellnet.Peer {

	// 尝试用Socket连接地址
	sess, err := kcp.DialWithOptions(self.Address(), nil, 0, 0)
	if err != nil {
		log.GetLog().Errorf("#kcp.connect failed(%s) %v", self.Name(), err.Error())
		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnectError{}})
		return self
	}

	self.defaultSes.SetKcpSession(sess)

	self.defaultSes.Start()

	self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnected{}})

	return self
}

func (self *kcpSyncConnector) Session() cellnet.Session {
	return self.defaultSes
}

func (self *kcpSyncConnector) SetSessionManager(raw interface{}) {
	self.SessionManager = raw.(peer.SessionManager)
}

func (self *kcpSyncConnector) ReconnectDuration() time.Duration {
	return 0
}

func (self *kcpSyncConnector) SetReconnectDuration(v time.Duration) {

}

func (self *kcpSyncConnector) Stop() {
	if self.defaultSes != nil {
		if c := self.defaultSes.GetKcpSession(); c != nil {
			c.Close()
		}
		self.defaultSes.Close()
	}

}

func (self *kcpSyncConnector) IsReady() bool {

	return self.SessionCount() != 0
}

func (self *kcpSyncConnector) TypeName() string {
	return "kcp.SyncConnector"
}

func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		self := &kcpSyncConnector{
			SessionManager: new(peer.CoreSessionManager),
		}

		self.defaultSes = newSession(nil, self, nil)

		self.CoreTCPSocketOption.Init()

		return self
	})
}
