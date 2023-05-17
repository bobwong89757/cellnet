package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/cellnet/util"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// tcpSession
// @Description: Socket会话
type tcpSession struct {
	peer.CoreContextSet
	peer.CoreSessionIdentify
	*peer.CoreProcBundle

	pInterface cellnet.Peer

	// Socket原始连接
	conn      net.Conn
	connGuard sync.RWMutex

	// 退出同步器
	exitSync sync.WaitGroup

	// 发送队列
	sendQueue *cellnet.Pipe

	cleanupGuard sync.Mutex

	endNotify func()

	closing int64
}

// setConn
//
//	@Description: 设置连接
//	@receiver self
//	@param conn
func (self *tcpSession) setConn(conn net.Conn) {
	self.connGuard.Lock()
	self.conn = conn
	self.connGuard.Unlock()
}

// Conn
//
//	@Description: 连接
//	@receiver self
//	@return net.Conn
func (self *tcpSession) Conn() net.Conn {
	self.connGuard.RLock()
	defer self.connGuard.RUnlock()
	return self.conn
}

// Peer
//
//	@Description: 获取端
//	@receiver self
//	@return cellnet.Peer
func (self *tcpSession) Peer() cellnet.Peer {
	return self.pInterface
}

// Raw
//
//	@Description: 取原始连接
//	@receiver self
//	@return interface{}
func (self *tcpSession) Raw() interface{} {
	return self.Conn()
}

// Close
//
//	@Description: 关闭连接器
//	@receiver self
func (self *tcpSession) Close() {

	closing := atomic.SwapInt64(&self.closing, 1)
	if closing != 0 {
		return
	}

	conn := self.Conn()

	if conn != nil {
		// 关闭读
		tcpConn := conn.(*net.TCPConn)
		// 关闭读
		tcpConn.CloseRead()
		// 手动读超时
		tcpConn.SetReadDeadline(time.Now())
	}
}

// Send
//
//	@Description: 发送封包
//	@receiver self
//	@param msg
func (self *tcpSession) Send(msg interface{}) {

	// 只能通过Close关闭连接
	if msg == nil {
		return
	}

	// 已经关闭，不再发送
	if self.IsManualClosed() {
		return
	}

	self.sendQueue.Add(msg)
}

// IsManualClosed
//
//	@Description: 是否手动关闭
//	@receiver self
//	@return bool
func (self *tcpSession) IsManualClosed() bool {
	return atomic.LoadInt64(&self.closing) != 0
}

// protectedReadMessage
//
//	@Description: 读取消息
//	@receiver self
//	@return msg
//	@return err
func (self *tcpSession) protectedReadMessage() (msg interface{}, err error) {

	defer func() {

		if err := recover(); err != nil {
			log.GetLog().Error("IO panic: %s", err)
			self.Conn().Close()
		}

	}()

	msg, err = self.ReadMessage(self)

	return
}

// recvLoop
//
//	@Description: 接收循环
//	@receiver self
func (self *tcpSession) recvLoop() {

	var capturePanic bool

	if i, ok := self.Peer().(cellnet.PeerCaptureIOPanic); ok {
		capturePanic = i.CaptureIOPanic()
	}

	for self.Conn() != nil {

		var msg interface{}
		var err error

		if capturePanic {
			msg, err = self.protectedReadMessage()
		} else {
			msg, err = self.ReadMessage(self)
		}

		if err != nil {
			if !util.IsEOFOrNetReadError(err) {

				var ip string
				if self.conn != nil {
					addr := self.conn.RemoteAddr()
					if addr != nil {
						ip = addr.String()
					}
				}

				log.GetLog().Error("session closed, sesid: %d, err: %s ip: %s", self.ID(), err, ip)
			}

			self.sendQueue.Add(nil)

			// 标记为手动关闭原因
			closedMsg := &cellnet.SessionClosed{}
			if self.IsManualClosed() {
				closedMsg.Reason = cellnet.CloseReason_Manual
			}

			self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: closedMsg})
			break
		}

		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: msg})
	}

	// 通知完成
	self.exitSync.Done()
}

func (self *tcpSession) protectedSendMessage(ev cellnet.Event) {

	defer func() {
		if err := recover(); err != nil {
			log.GetLog().Error("IO send panic: %s %s", err, cellnet.MessageToName(ev.Message()))
		}

	}()

	self.SendMessage(ev)
}

// 发送循环
func (self *tcpSession) sendLoop() {

	var writeList []interface{}

	var capturePanic bool

	if i, ok := self.Peer().(cellnet.PeerCaptureIOPanic); ok {
		capturePanic = i.CaptureIOPanic()
	}

	for {
		writeList = writeList[0:0]
		exit := self.sendQueue.Pick(&writeList)

		// 遍历要发送的数据
		for _, msg := range writeList {

			if capturePanic {
				self.protectedSendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
			} else {
				self.SendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
			}
		}

		if exit {
			break
		}
	}

	// 完整关闭
	conn := self.Conn()
	if conn != nil {
		conn.Close()
	}

	// 通知完成
	self.exitSync.Done()
}

// Start
//
//	@Description: 启动会话的各种资源
//	@receiver self
func (self *tcpSession) Start() {

	atomic.StoreInt64(&self.closing, 0)

	// connector复用session时，上一次发送队列未释放可能造成问题
	self.sendQueue.Reset()

	// 需要接收和发送线程同时完成时才算真正的完成
	self.exitSync.Add(2)

	// 将会话添加到管理器, 在线程处理前添加到管理器(分配id), 避免ID还未分配,就开始使用id的竞态问题
	self.Peer().(peer.SessionManager).Add(self)

	go func() {

		// 等待2个任务结束
		self.exitSync.Wait()

		// 将会话从管理器移除
		self.Peer().(peer.SessionManager).Remove(self)

		if self.endNotify != nil {
			self.endNotify()
		}

	}()

	// 启动并发接收goroutine
	go self.recvLoop()

	// 启动并发发送goroutine
	go self.sendLoop()
}

// newSession
//
//	@Description: 新建会话
//	@param conn
//	@param p
//	@param endNotify
//	@return *tcpSession
func newSession(conn net.Conn, p cellnet.Peer, endNotify func()) *tcpSession {
	self := &tcpSession{
		conn:       conn,
		endNotify:  endNotify,
		sendQueue:  cellnet.NewPipe(),
		pInterface: p,
		CoreProcBundle: p.(interface {
			GetBundle() *peer.CoreProcBundle
		}).GetBundle(),
	}

	return self
}
