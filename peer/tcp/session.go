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

// tcpSession TCP 会话实现
// 表示一个 TCP 连接，负责消息的接收和发送
// 使用独立的 goroutine 处理接收和发送，实现异步通信
type tcpSession struct {
	peer.CoreContextSet      // 上下文数据存储
	peer.CoreSessionIdentify // 会话 ID 管理
	*peer.CoreProcBundle     // 消息处理组件（编码器、钩子、回调等）

	// pInterface 所属的 Peer
	// 用于访问 Peer 的配置和功能
	pInterface cellnet.Peer

	// conn Socket 原始连接
	// connGuard 连接读写锁，保护 conn 的并发访问
	conn      net.Conn
	connGuard sync.RWMutex

	// exitSync 退出同步器
	// 用于等待接收和发送 goroutine 结束
	exitSync sync.WaitGroup

	// sendQueue 发送队列
	// 使用 Pipe 实现非阻塞的消息发送
	sendQueue *cellnet.Pipe

	// cleanupGuard 清理保护锁
	// 用于保护清理操作的并发安全
	cleanupGuard sync.Mutex

	// endNotify 会话结束回调
	// 当会话结束时会被调用
	endNotify func()

	// closing 关闭标记
	// 使用原子操作，1 表示正在关闭或已关闭，0 表示正常
	closing int64
}

// setConn 设置连接
// conn: TCP 连接对象
// 使用写锁保护，确保并发安全
func (self *tcpSession) setConn(conn net.Conn) {
	self.connGuard.Lock()
	self.conn = conn
	self.connGuard.Unlock()
}

// Conn 获取连接
// 返回当前的 TCP 连接
// 使用读锁保护，允许多个 goroutine 同时读取
func (self *tcpSession) Conn() net.Conn {
	self.connGuard.RLock()
	defer self.connGuard.RUnlock()
	return self.conn
}

// Peer 获取所属的 Peer
// 返回会话所属的 Peer 对象
func (self *tcpSession) Peer() cellnet.Peer {
	return self.pInterface
}

// Raw 获取原始连接
// 返回底层的 TCP 连接对象
// 用于需要直接访问底层连接的高级操作
func (self *tcpSession) Raw() interface{} {
	return self.Conn()
}

// Close 关闭会话
// 标记会话为关闭状态，并关闭连接的读端
// 使用原子操作确保只执行一次关闭操作
// 关闭读端会触发接收循环退出，发送循环会在发送完队列中的消息后退出
func (self *tcpSession) Close() {

	// 原子交换，如果已经是关闭状态，直接返回
	closing := atomic.SwapInt64(&self.closing, 1)
	if closing != 0 {
		return
	}

	conn := self.Conn()

	if conn != nil {
		// 转换为 TCP 连接以使用 TCP 特定方法
		tcpConn := conn.(*net.TCPConn)
		// 关闭读端，触发接收循环退出
		tcpConn.CloseRead()
		// 设置读超时为当前时间，确保阻塞的读操作立即返回
		tcpConn.SetReadDeadline(time.Now())
	}
}

// Send 发送消息
// msg: 要发送的消息对象
// 将消息添加到发送队列，由发送循环异步发送
// 如果会话已关闭，消息将被丢弃
func (self *tcpSession) Send(msg interface{}) {

	// nil 消息不发送
	if msg == nil {
		return
	}

	// 如果已经关闭，不再发送
	if self.IsManualClosed() {
		return
	}

	// 添加到发送队列
	self.sendQueue.Add(msg)
}

// IsManualClosed 检查会话是否已手动关闭
// 返回 true 表示会话已关闭，false 表示会话正常
// 使用原子操作读取关闭标记
func (self *tcpSession) IsManualClosed() bool {
	return atomic.LoadInt64(&self.closing) != 0
}

// protectedReadMessage 受保护的读取消息
// 捕获读取过程中的 panic，防止整个程序崩溃
// 如果发生 panic，记录日志并关闭连接
// 返回读取的消息和错误
func (self *tcpSession) protectedReadMessage() (msg interface{}, err error) {

	defer func() {

		if err := recover(); err != nil {
			log.GetLog().Errorf("IO panic: %s", err)
			// 发生 panic 时关闭连接
			self.Conn().Close()
		}

	}()

	msg, err = self.ReadMessage(self)

	return
}

// recvLoop 接收循环
// 在独立的 goroutine 中运行，持续从连接读取消息
// 读取到的消息会通过事件系统分发给上层应用
// 当连接关闭或发生错误时，退出循环
func (self *tcpSession) recvLoop() {

	// 检查是否需要捕获 panic
	var capturePanic bool

	if i, ok := self.Peer().(cellnet.PeerCaptureIOPanic); ok {
		capturePanic = i.CaptureIOPanic()
	}

	// 持续读取消息
	for self.Conn() != nil {

		var msg interface{}
		var err error

		// 根据配置决定是否使用受保护的读取
		if capturePanic {
			msg, err = self.protectedReadMessage()
		} else {
			msg, err = self.ReadMessage(self)
		}

		// 处理读取错误
		if err != nil {
			// 如果不是正常的 EOF 或网络读取错误，记录日志
			if !util.IsEOFOrNetReadError(err) {

				var ip string
				if self.conn != nil {
					addr := self.conn.RemoteAddr()
					if addr != nil {
						ip = addr.String()
					}
				}

				log.GetLog().Errorf("session closed, sesid: %d, err: %s ip: %s", self.ID(), err, ip)
			}

			// 向发送队列添加 nil，触发发送循环退出
			self.sendQueue.Add(nil)

			// 创建关闭消息，标记关闭原因
			closedMsg := &cellnet.SessionClosed{}
			if self.IsManualClosed() {
				closedMsg.Reason = cellnet.CloseReason_Manual
			}

			// 发送关闭事件
			self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: closedMsg})
			break
		}

		// 发送接收到的消息事件
		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: msg})
	}

	// 通知接收循环已完成
	self.exitSync.Done()
}

// protectedSendMessage 受保护的发送消息
// ev: 发送事件
// 捕获发送过程中的 panic，防止整个程序崩溃
// 如果发生 panic，记录日志（不关闭连接，因为可能只是单个消息的问题）
func (self *tcpSession) protectedSendMessage(ev cellnet.Event) {

	defer func() {
		if err := recover(); err != nil {
			log.GetLog().Errorf("IO send panic: %s %s", err, cellnet.MessageToName(ev.Message()))
		}

	}()

	self.SendMessage(ev)
}

// sendLoop 发送循环
// 在独立的 goroutine 中运行，持续从发送队列取出消息并发送
// 当收到 nil 消息时，退出循环并关闭连接
func (self *tcpSession) sendLoop() {

	var writeList []interface{}

	// 检查是否需要捕获 panic
	var capturePanic bool

	if i, ok := self.Peer().(cellnet.PeerCaptureIOPanic); ok {
		capturePanic = i.CaptureIOPanic()
	}

	// 持续从队列取出消息并发送
	for {
		// 清空列表，复用切片
		writeList = writeList[0:0]
		// 从队列中批量取出消息
		exit := self.sendQueue.Pick(&writeList)

		// 遍历要发送的消息
		for _, msg := range writeList {

			// 根据配置决定是否使用受保护的发送
			if capturePanic {
				self.protectedSendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
			} else {
				self.SendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
			}
		}

		// 如果收到退出信号（nil 消息），退出循环
		if exit {
			break
		}
	}

	// 完整关闭连接
	conn := self.Conn()
	if conn != nil {
		conn.Close()
	}

	// 通知发送循环已完成
	self.exitSync.Done()
}

// Start 启动会话
// 初始化会话状态，启动接收和发送 goroutine
// 会话会被添加到会话管理器，分配 ID
func (self *tcpSession) Start() {

	// 重置关闭标记
	atomic.StoreInt64(&self.closing, 0)

	// connector 复用 session 时，上一次发送队列未释放可能造成问题
	// 重置发送队列，清空之前的消息
	self.sendQueue.Reset()

	// 需要接收和发送线程同时完成时才算真正的完成
	// 设置等待计数为 2（接收循环和发送循环）
	self.exitSync.Add(2)

	// 将会话添加到管理器
	// 在线程处理前添加到管理器（分配 ID），避免 ID 还未分配就开始使用 ID 的竞态问题
	self.Peer().(peer.SessionManager).Add(self)

	// 启动清理 goroutine
	go func() {

		// 等待接收和发送循环都结束
		self.exitSync.Wait()

		// 将会话从管理器移除
		self.Peer().(peer.SessionManager).Remove(self)

		// 调用结束回调
		if self.endNotify != nil {
			self.endNotify()
		}

	}()

	// 启动并发接收 goroutine
	go self.recvLoop()

	// 启动并发发送 goroutine
	go self.sendLoop()
}

// newSession 创建新的 TCP 会话
// conn: TCP 连接对象，可以为 nil（连接器会在连接成功后再设置）
// p: 所属的 Peer 对象
// endNotify: 会话结束时的回调函数，可以为 nil
// 返回新创建的会话对象
func newSession(conn net.Conn, p cellnet.Peer, endNotify func()) *tcpSession {
	self := &tcpSession{
		conn:       conn,
		endNotify:  endNotify,
		sendQueue:  cellnet.NewPipe(),
		pInterface: p,
		// 从 Peer 获取消息处理组件
		CoreProcBundle: p.(interface {
			GetBundle() *peer.CoreProcBundle
		}).GetBundle(),
	}

	return self
}
