package udp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/cellnet/util"
	"net"
	"time"
)

// MaxUDPRecvBuffer UDP 接收缓冲区最大大小
// 用于接收 UDP 数据包
const MaxUDPRecvBuffer = 2048

// udpAcceptor UDP 接受器实现
// 用于创建 UDP 服务器，接受客户端数据包
// UDP 是无连接协议，需要基于地址管理 Session
type udpAcceptor struct {
	peer.CoreSessionManager  // 会话管理器，用于管理所有客户端连接
	peer.CorePeerProperty    // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreContextSet      // 上下文数据存储
	peer.CoreRunningTag      // 运行状态标记
	peer.CoreProcBundle      // 消息处理组件（编码器、钩子、回调等）
	peer.CoreCaptureIOPanic  // IO 层 panic 捕获控制

	// conn UDP 连接
	// 用于接收和发送 UDP 数据包
	conn *net.UDPConn

	// sesTimeout Session 生存时间
	// 如果在此时间内没有收到来自该地址的数据包，Session 会被标记为超时
	sesTimeout time.Duration

	// sesCleanTimeout Session 清理检查间隔
	// 每隔此时间检查一次超时的 Session 并清理
	sesCleanTimeout time.Duration

	// sesCleanLastTime 上次清理 Session 的时间
	// 用于判断是否需要执行清理操作
	sesCleanLastTime time.Time

	// sesByConnTrack 基于连接跟踪键的 Session 映射
	// 使用地址（IP+端口）作为键来管理 Session
	sesByConnTrack map[connTrackKey]*udpSession
}

// IsReady 检查接受器是否已准备好
// 返回 true 表示接受器正在运行并可以接受数据包
// 返回 false 表示接受器未运行
func (self *udpAcceptor) IsReady() bool {

	return self.IsRunning()
}

// Port 获取当前侦听的端口号
// 如果连接未初始化，返回 0
// 返回当前 UDP 连接绑定的端口号
func (self *udpAcceptor) Port() int {
	if self.conn == nil {
		return 0
	}

	return self.conn.LocalAddr().(*net.UDPAddr).Port
}

// Start 异步开始侦听 UDP 数据包
// 如果接受器已经在运行，直接返回
// 如果地址中包含端口 0，会自动分配可用端口
// 启动成功后会在后台 goroutine 中接受数据包
// 返回自身以支持链式调用
func (self *udpAcceptor) Start() cellnet.Peer {

	var finalAddr *util.Address
	// 尝试监听指定地址，如果端口为 0 则自动分配
	ln, err := util.DetectPort(self.Address(), func(a *util.Address, port int) (interface{}, error) {

		// 解析 UDP 地址
		addr, err := net.ResolveUDPAddr("udp", a.HostPortString(port))
		if err != nil {
			return nil, err
		}

		finalAddr = a

		// 创建 UDP 监听连接
		return net.ListenUDP("udp", addr)
	})

	if err != nil {
		// 监听失败，记录错误
		log.GetLog().Errorf("#udp.listen failed(%s) %v", self.Name(), err.Error())
		return self
	}

	self.conn = ln.(*net.UDPConn)

	log.GetLog().Infof("#udp.listen(%s) %s", self.Name(), finalAddr.String(self.Port()))

	// 在后台 goroutine 中接受数据包
	go self.accept()

	return self
}

// protectedRecvPacket 受保护的数据包接收
// ses: 接收数据包的会话
// data: 接收到的数据
// 捕获接收过程中的 panic，防止整个程序崩溃
// 如果发生 panic，记录日志并关闭连接
func (self *udpAcceptor) protectedRecvPacket(ses *udpSession, data []byte) {
	defer func() {

		if err := recover(); err != nil {
			log.GetLog().Errorf("IO panic: %s", err)
			// 发生 panic 时关闭连接
			self.conn.Close()
		}

	}()

	ses.Recv(data)
}

// accept 接受数据包的循环
// 在后台 goroutine 中运行，持续从 UDP 连接读取数据包
// 根据数据包的源地址获取或创建 Session，并处理数据包
// 定期检查并清理超时的 Session
func (self *udpAcceptor) accept() {

	self.SetRunning(true)

	// 创建接收缓冲区
	recvBuff := make([]byte, MaxUDPRecvBuffer)

	for {

		// 从 UDP 连接读取数据包
		n, remoteAddr, err := self.conn.ReadFromUDP(recvBuff)
		if err != nil {
			// 读取错误，退出循环
			break
		}

		// 检查并清理超时的 Session
		self.checkTimeoutSession()

		if n > 0 {
			// 根据源地址获取或创建 Session
			ses := self.getSession(remoteAddr)

			// 根据配置决定是否使用受保护的接收
			if self.CaptureIOPanic() {
				self.protectedRecvPacket(ses, recvBuff[:n])
			} else {
				ses.Recv(recvBuff[:n])
			}

		}

	}

	self.SetRunning(false)

}

// checkTimeoutSession 检查并清理超时的 Session
// 定期执行清理操作，移除已经超时的 Session 以释放内存
// 清理间隔由 sesCleanTimeout 控制
func (self *udpAcceptor) checkTimeoutSession() {
	now := time.Now()

	// 定时清理超时的 Session
	// 如果距离上次清理时间超过清理间隔，执行清理
	if now.After(self.sesCleanLastTime.Add(self.sesCleanTimeout)) {
		// 收集需要删除的 Session
		sesToDelete := make([]*udpSession, 0, 10)
		for _, ses := range self.sesByConnTrack {
			if !ses.IsAlive() {
				sesToDelete = append(sesToDelete, ses)
			}
		}

		// 删除超时的 Session
		for _, ses := range sesToDelete {
			delete(self.sesByConnTrack, *ses.key)
		}

		// 更新上次清理时间
		self.sesCleanLastTime = now
	}
}

// getSession 根据地址获取或创建 Session
// addr: 数据包的源地址
// 如果该地址的 Session 不存在，创建一个新的 Session
// 如果已存在，更新其超时时间（续租）
// 返回对应的 Session
func (self *udpAcceptor) getSession(addr *net.UDPAddr) *udpSession {

	// 根据地址创建连接跟踪键
	key := newConnTrackKey(addr)

	// 查找是否已存在该地址的 Session
	ses := self.sesByConnTrack[*key]

	if ses == nil {
		// 不存在，创建新的 Session
		ses = &udpSession{}
		ses.conn = self.conn
		ses.remote = addr
		ses.pInterface = self
		ses.CoreProcBundle = &self.CoreProcBundle
		ses.key = key
		// 添加到映射中
		self.sesByConnTrack[*key] = ses
	}

	// 续租：更新 Session 的超时时间
	// 每次收到来自该地址的数据包时，都会延长 Session 的生存时间
	ses.timeOutTick = time.Now().Add(self.sesTimeout)

	return ses
}

// SetSessionTTL 设置 Session 的生存时间（TTL）
// dur: Session 的生存时间
// 如果在此时间内没有收到来自该地址的数据包，Session 会被标记为超时
func (self *udpAcceptor) SetSessionTTL(dur time.Duration) {
	self.sesTimeout = dur
}

// SetSessionCleanTimeout 设置 Session 清理检查间隔
// dur: 清理检查间隔
// 每隔此时间检查一次超时的 Session 并清理
func (self *udpAcceptor) SetSessionCleanTimeout(dur time.Duration) {
	self.sesCleanTimeout = dur
}

// Stop 停止接受器
// 关闭 UDP 连接，停止接受数据包
// 注意：当前实现不等待 accept goroutine 结束
func (self *udpAcceptor) Stop() {

	if self.conn != nil {
		self.conn.Close()
	}

	// TODO 等待 accept 线程结束
	self.SetRunning(false)
}

// TypeName 返回接受器的类型名称
// 用于标识和日志记录
func (self *udpAcceptor) TypeName() string {
	return "udp.Acceptor"
}

// init 包初始化函数
// 自动注册 UDP 接受器的创建函数
// 当调用 cellnet.NewPeer("udp.Acceptor", ...) 时会使用此函数创建实例
func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &udpAcceptor{
			// 默认 Session 生存时间为 1 分钟
			sesTimeout: time.Minute,
			// 默认清理检查间隔为 1 分钟
			sesCleanTimeout:  time.Minute,
			sesCleanLastTime: time.Now(),
			sesByConnTrack:   make(map[connTrackKey]*udpSession),
		}

		return p
	})
}
