package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
)

// tcpConnector TCP 连接器实现
// 用于创建 TCP 客户端，连接到服务器
// 支持自动重连功能
type tcpConnector struct {
	peer.SessionManager      // 会话管理器
	peer.CorePeerProperty    // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreContextSet      // 上下文数据存储
	peer.CoreRunningTag      // 运行状态标记
	peer.CoreProcBundle      // 消息处理组件（编码器、钩子、回调等）
	peer.CoreTCPSocketOption // TCP Socket 选项（缓冲区、超时等）
	peer.CoreCaptureIOPanic  // IO 层 panic 捕获控制

	// defaultSes 默认会话
	// 连接器通常只有一个会话
	defaultSes *tcpSession

	// tryConnTimes 尝试连接次数
	// 用于控制连接失败日志的输出频率
	tryConnTimes int

	// sesEndSignal 会话结束信号
	// 用于等待会话结束
	sesEndSignal sync.WaitGroup

	// reconDur 重连时间间隔
	// 当连接断开时，会在此时间后尝试重新连接
	// 如果为 0，表示不自动重连
	reconDur time.Duration
}

// Start 开始连接
// 如果连接器已经在运行，直接返回
// 在后台 goroutine 中执行连接逻辑
// 返回自身以支持链式调用
func (self *tcpConnector) Start() cellnet.Peer {

	// 等待上一次停止完成
	self.WaitStopFinished()

	// 如果已经在运行，直接返回
	if self.IsRunning() {
		return self
	}

	// 在后台 goroutine 中执行连接
	go self.connect(self.Address())

	return self
}

// Session 获取当前会话
// 返回连接器的默认会话
// 如果未连接，会话可能为 nil
func (self *tcpConnector) Session() cellnet.Session {
	return self.defaultSes
}

// SetSessionManager 设置会话管理器
// raw: 实现 peer.SessionManager 接口的对象
// 用于自定义会话管理逻辑
func (self *tcpConnector) SetSessionManager(raw interface{}) {
	self.SessionManager = raw.(peer.SessionManager)
}

// Stop 停止连接器
// 关闭会话，停止重连，并等待所有 goroutine 结束
// 如果连接器未运行或正在停止，直接返回
func (self *tcpConnector) Stop() {
	if !self.IsRunning() {
		return
	}

	if self.IsStopping() {
		return
	}

	// 标记开始停止
	self.StartStopping()

	// 关闭会话，触发接收和发送循环退出
	self.defaultSes.Close()

	// 等待所有 goroutine 结束
	self.WaitStopFinished()

}

// ReconnectDuration 获取重连时间间隔
// 返回当前设置的重连时间间隔
// 如果为 0，表示不自动重连
func (self *tcpConnector) ReconnectDuration() time.Duration {

	return self.reconDur
}

// SetReconnectDuration 设置重连时间间隔
// v: 重连时间间隔
// 当连接断开时，会在此时间后尝试重新连接
// 如果设置为 0，表示关闭自动重连
func (self *tcpConnector) SetReconnectDuration(v time.Duration) {
	self.reconDur = v
}

// Port 获取本地端口号
// 返回当前连接使用的本地端口号
// 如果未连接，返回 0
func (self *tcpConnector) Port() int {

	conn := self.defaultSes.Conn()

	if conn == nil {
		return 0
	}

	return conn.LocalAddr().(*net.TCPAddr).Port
}

// reportConnectFailedLimitTimes 连接失败日志报告次数限制
// 超过此次数后，连接失败日志将被静默，避免日志刷屏
const reportConnectFailedLimitTimes = 3

// connect 连接循环
// address: 服务器地址（格式：host:port）
// 在后台 goroutine 中运行，持续尝试连接服务器
// 支持自动重连功能，连接断开后会在指定时间后重新连接
func (self *tcpConnector) connect(address string) {

	self.SetRunning(true)

	for {
		self.tryConnTimes++

		// 尝试用 Socket 连接地址
		conn, err := net.Dial("tcp", address)

		// 设置连接（即使失败也设置，以便后续处理）
		self.defaultSes.setConn(conn)

		// 连接失败处理
		if err != nil {

			// 前几次连接失败时记录日志
			if self.tryConnTimes <= reportConnectFailedLimitTimes {
				log.GetLog().Errorf("#tcp.connect failed(%s) %v", self.Name(), err.Error())

				// 达到限制次数时，提示后续日志将被静默
				if self.tryConnTimes == reportConnectFailedLimitTimes {
					log.GetLog().Errorf("(%s) continue reconnecting, but mute log", self.Name())
				}
			}

			// 如果没有设置重连或正在停止，发送连接错误事件并退出
			if self.ReconnectDuration() == 0 || self.IsStopping() {

				self.ProcEvent(&cellnet.RecvMsgEvent{
					Ses: self.defaultSes,
					Msg: &cellnet.SessionConnectError{},
				})
				break
			}

			// 有重连设置，等待后继续尝试
			time.Sleep(self.ReconnectDuration())

			// 继续连接
			continue
		}

		// 连接成功，启动会话
		self.sesEndSignal.Add(1)

		// 应用 Socket 选项
		self.ApplySocketOption(conn)

		// 启动会话（启动接收和发送循环）
		self.defaultSes.Start()

		// 重置连接尝试次数
		self.tryConnTimes = 0

		// 发送连接成功事件
		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self.defaultSes, Msg: &cellnet.SessionConnected{}})

		// 等待会话结束（连接断开）
		self.sesEndSignal.Wait()

		// 清空连接
		self.defaultSes.setConn(nil)

		// 如果正在停止或没有设置重连，退出循环
		if self.IsStopping() || self.ReconnectDuration() == 0 {
			break
		}

		// 有重连设置，等待后继续尝试
		time.Sleep(self.ReconnectDuration())

		// 继续连接
		continue

	}

	self.SetRunning(false)

	// 标记停止完成
	self.EndStopping()
}

// IsReady 检查连接器是否已准备好
// 返回 true 表示已成功连接到服务器
// 返回 false 表示未连接
func (self *tcpConnector) IsReady() bool {

	return self.SessionCount() != 0
}

// TypeName 返回连接器的类型名称
// 用于标识和日志记录
func (self *tcpConnector) TypeName() string {
	return "tcp.Connector"
}

// init 包初始化函数
// 自动注册 TCP 连接器的创建函数
// 当调用 cellnet.NewPeer("tcp.Connector", ...) 时会使用此函数创建实例
func init() {
	// 注册 Peer 创建函数
	peer.RegisterPeerCreator(func() cellnet.Peer {
		self := &tcpConnector{
			SessionManager: new(peer.CoreSessionManager),
		}

		// 创建默认会话，设置结束回调
		self.defaultSes = newSession(nil, self, func() {
			self.sesEndSignal.Done()
		})

		// 初始化 TCP Socket 选项
		self.CoreTCPSocketOption.Init()

		return self
	})
}
