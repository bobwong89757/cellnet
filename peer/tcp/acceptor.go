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

// tcpAcceptor TCP 接受器实现
// 用于创建 TCP 服务器，接受客户端连接
// 组合了多个核心组件以提供完整的功能
type tcpAcceptor struct {
	peer.SessionManager      // 会话管理器，用于管理所有客户端连接
	peer.CorePeerProperty    // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreContextSet      // 上下文数据存储
	peer.CoreRunningTag      // 运行状态标记
	peer.CoreProcBundle      // 消息处理组件（编码器、钩子、回调等）
	peer.CoreTCPSocketOption // TCP Socket 选项（缓冲区、超时等）
	peer.CoreCaptureIOPanic  // IO 层 panic 捕获控制

	// listener 保存 TCP 侦听器
	// 用于接受客户端连接
	listener net.Listener
}

// Port 获取当前侦听的端口号
// 如果 listener 未初始化，返回 0
// 返回当前 TCP 侦听器绑定的端口号
func (self *tcpAcceptor) Port() int {
	if self.listener == nil {
		return 0
	}

	return self.listener.Addr().(*net.TCPAddr).Port
}

// IsReady 检查接受器是否已准备好
// 返回 true 表示接受器正在运行并可以接受连接
// 返回 false 表示接受器未运行
func (self *tcpAcceptor) IsReady() bool {

	return self.IsRunning()
}

// Start 异步开始侦听客户端连接
// 如果接受器已经在运行，直接返回
// 如果地址中包含端口 0，会自动分配可用端口
// 启动成功后会在后台 goroutine 中接受连接
// 返回自身以支持链式调用
func (self *tcpAcceptor) Start() cellnet.Peer {

	// 等待上一次停止完成
	self.WaitStopFinished()

	// 如果已经在运行，直接返回
	if self.IsRunning() {
		return self
	}

	// 尝试监听指定地址，如果端口为 0 则自动分配
	ln, err := util.DetectPort(self.Address(), func(a *util.Address, port int) (interface{}, error) {
		return net.Listen("tcp", a.HostPortString(port))
	})

	if err != nil {
		// 监听失败，记录错误并设置运行状态为 false
		log.GetLog().Errorf("#tcp.listen failed(%s) %v", self.Name(), err.Error())

		self.SetRunning(false)

		return self
	}

	self.listener = ln.(net.Listener)

	log.GetLog().Infof("#tcp.listen(%s) %s", self.Name(), self.ListenAddress())

	// 在后台 goroutine 中接受连接
	go self.accept()

	return self
}

// ListenAddress 获取完整的监听地址（包含实际端口）
// 如果原始地址中没有端口，直接返回原始地址
// 否则返回 "host:实际端口" 格式的地址
// 用于显示实际监听的地址（特别是当使用端口 0 自动分配时）
func (self *tcpAcceptor) ListenAddress() string {

	pos := strings.Index(self.Address(), ":")
	if pos == -1 {
		return self.Address()
	}

	host := self.Address()[:pos]

	return util.JoinAddress(host, self.Port())
}

// accept 接受连接的循环
// 在后台 goroutine 中运行，持续接受客户端连接
// 每个新连接会在独立的 goroutine 中处理，避免阻塞 accept 循环
// 当接受器停止时，退出循环
func (self *tcpAcceptor) accept() {
	self.SetRunning(true)

	for {
		conn, err := self.listener.Accept()

		// 如果正在停止，退出循环
		if self.IsStopping() {
			break
		}

		if err == nil {
			// 处理连接进入独立线程，防止 accept 无法响应
			// 这样可以快速接受下一个连接，提高并发处理能力
			go self.onNewSession(conn)

		} else {
			// 处理临时错误，短暂等待后重试
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}

			// 非临时错误，记录日志并退出
			log.GetLog().Errorf("#tcp.accept failed(%s) %v", self.Name(), err.Error())
			break
		}
	}

	self.SetRunning(false)

	// 标记停止完成
	self.EndStopping()

}

// onNewSession 处理新建立的连接会话
// conn: 新接受的 TCP 连接
// 应用 Socket 选项，创建会话，启动会话，并发送 SessionAccepted 事件
func (self *tcpAcceptor) onNewSession(conn net.Conn) {

	// 应用配置的 Socket 选项（缓冲区大小、Nagle 算法等）
	self.ApplySocketOption(conn)

	// 创建新的会话对象
	ses := newSession(conn, self, nil)

	// 启动会话（启动接收和发送循环）
	ses.Start()

	// 发送会话已接受事件，通知上层应用
	self.ProcEvent(&cellnet.RecvMsgEvent{
		Ses: ses,
		Msg: &cellnet.SessionAccepted{},
	})
}

// Stop 停止接受器
// 关闭监听器，断开所有客户端连接，并等待所有 goroutine 结束
// 如果接受器未运行或正在停止，直接返回
func (self *tcpAcceptor) Stop() {
	if !self.IsRunning() {
		return
	}

	if self.IsStopping() {
		return
	}

	// 标记开始停止
	self.StartStopping()

	// 关闭监听器，停止接受新连接
	self.listener.Close()

	// 断开所有已建立的连接
	self.CloseAllSession()

	// 等待所有 goroutine 结束
	self.WaitStopFinished()
}

// TypeName 返回接受器的类型名称
// 用于标识和日志记录
func (self *tcpAcceptor) TypeName() string {
	return "tcp.Acceptor"
}

// init 包初始化函数
// 自动注册 TCP 接受器的创建函数
// 当调用 cellnet.NewPeer("tcp.Acceptor", ...) 时会使用此函数创建实例
func init() {
	// 注册 Peer 创建函数
	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &tcpAcceptor{
			SessionManager: new(peer.CoreSessionManager),
		}

		// 初始化 TCP Socket 选项
		p.CoreTCPSocketOption.Init()

		return p
	})
}
