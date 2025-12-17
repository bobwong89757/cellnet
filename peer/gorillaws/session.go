package gorillaws

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/cellnet/util"
	"github.com/gorilla/websocket"
)

// wsSession
// @Description: ws会话
type wsSession struct {
	peer.CoreContextSet
	peer.CoreSessionIdentify
	*peer.CoreProcBundle

	pInterface cellnet.Peer

	conn      *websocket.Conn
	connGuard sync.RWMutex

	// 退出同步器
	exitSync sync.WaitGroup

	// 发送队列
	sendQueue *cellnet.Pipe

	cleanupGuard sync.Mutex

	// closing 关闭标记，使用原子操作，1 表示正在关闭或已关闭，0 表示正常
	closing int64

	endNotify func()
}

func (self *wsSession) Peer() cellnet.Peer {
	return self.pInterface
}

// 取原始连接
func (self *wsSession) Raw() interface{} {
	self.connGuard.RLock()
	defer self.connGuard.RUnlock()
	if self.conn == nil {
		return nil
	}

	return self.conn
}

// getConn 获取连接（内部使用）
func (self *wsSession) getConn() *websocket.Conn {
	self.connGuard.RLock()
	defer self.connGuard.RUnlock()
	return self.conn
}

// closeConn 关闭连接（内部使用，确保只关闭一次）
func (self *wsSession) closeConn() {
	// 先检查关闭标记，如果已经关闭，直接返回
	if atomic.LoadInt64(&self.closing) != 0 {
		return
	}

	// 尝试设置关闭标记，如果已经是关闭状态，直接返回
	if atomic.SwapInt64(&self.closing, 1) != 0 {
		return
	}

	// 获取锁并关闭连接
	self.connGuard.Lock()
	defer self.connGuard.Unlock()

	if self.conn != nil {
		// 设置读写超时为当前时间，确保阻塞的读写操作立即返回
		// 这样可以释放 gorilla/websocket 内部的 bufio reader/writer
		self.conn.SetReadDeadline(time.Now())
		self.conn.SetWriteDeadline(time.Now())
		// 关闭连接，释放 bufio.NewReaderSize 和 bufio.NewWriterSize 创建的缓冲区
		self.conn.Close()
		self.conn = nil
	}
}

func (self *wsSession) Close() {
	// 如果已经关闭，直接返回
	if atomic.LoadInt64(&self.closing) != 0 {
		return
	}

	// 关闭连接，确保 bufio 缓冲区被释放
	self.closeConn()

	// 通知发送循环退出
	self.sendQueue.Add(nil)
}

// 发送封包
func (self *wsSession) Send(msg interface{}) {
	self.sendQueue.Add(msg)
}

func (self *wsSession) protectedReadMessage() (msg interface{}, err error) {

	defer func() {

		if err := recover(); err != nil {
			log.GetLog().Errorf("IO read panic: %s", err)
			self.Close()
		}

	}()

	msg, err = self.ReadMessage(self)

	return
}

// 接收循环
func (self *wsSession) recvLoop() {

	var capturePanic bool

	if i, ok := self.Peer().(cellnet.PeerCaptureIOPanic); ok {
		capturePanic = i.CaptureIOPanic()
	}

	for self.getConn() != nil {

		var msg interface{}
		var err error

		if capturePanic {
			msg, err = self.protectedReadMessage()
		} else {
			msg, err = self.ReadMessage(self)
		}

		if err != nil {

			if !util.IsEOFOrNetReadError(err) {
				log.GetLog().Debugf("session closed: %v", err.Error())
			}

			// 关闭连接，确保 bufio reader 被释放
			self.closeConn()

			self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: &cellnet.SessionClosed{}})
			break
		}

		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: msg})
	}

	// 确保连接被关闭（如果还没有关闭）
	self.closeConn()

	// 通知发送循环退出
	self.sendQueue.Add(nil)

	// 通知完成
	self.exitSync.Done()
}

// 发送循环
func (self *wsSession) sendLoop() {

	var writeList []interface{}

	for {
		writeList = writeList[0:0]
		exit := self.sendQueue.Pick(&writeList)

		// 遍历要发送的数据
		for _, msg := range writeList {
			// 检查连接是否已关闭
			if atomic.LoadInt64(&self.closing) != 0 {
				break
			}

			// 检查连接是否仍然有效
			conn := self.getConn()
			if conn == nil {
				// 连接已被关闭，退出循环
				self.closeConn()
				break
			}

			// TODO SendMsgEvent并不是很有意义
			// 注意：SendMessage 不返回错误，但 OnSendMessage 会返回错误
			// 如果发送失败，会在下次循环中检测到连接问题
			self.SendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
		}

		if exit {
			break
		}
	}

	// 关闭连接，确保 bufio writer 被释放
	// 使用 closeConn 确保只关闭一次，避免竞态条件
	self.closeConn()

	// 通知完成
	self.exitSync.Done()
}

// 启动会话的各种资源
func (self *wsSession) Start() {

	// 重置关闭标记
	atomic.StoreInt64(&self.closing, 0)

	// 将会话添加到管理器
	self.Peer().(peer.SessionManager).Add(self)

	// 需要接收和发送线程同时完成时才算真正的完成
	self.exitSync.Add(2)

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

func newSession(conn *websocket.Conn, p cellnet.Peer, endNotify func()) *wsSession {
	self := &wsSession{
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
