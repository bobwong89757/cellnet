package cellnet

// Peer 表示一个网络端点，可以是服务器（Acceptor）或客户端（Connector）
// Peer 是 cellnet 框架的核心接口，代表一个网络通信端点
// 可以通过类型断言查询更多接口支持，如 PeerProperty、ContextSet、SessionAccessor
//
// 使用示例:
//
//	peerIns := peer.NewGenericPeer("tcp.Acceptor", "server", "127.0.0.1:17701", queue)
//	peerIns.Start()
type Peer interface {
	// Start 启动网络端点
	// 对于 Acceptor，会开始监听指定地址
	// 对于 Connector，会尝试连接到指定地址
	// 返回自身以便链式调用
	Start() Peer

	// Stop 停止网络端点
	// 停止监听或断开连接，并清理相关资源
	Stop()

	// TypeName 返回 Peer 的类型名称
	// 格式为 "protocol.type"，例如 "tcp.Acceptor"、"udp.Connector"
	TypeName() string
}

// PeerProperty 提供 Peer 的基础属性访问接口
// 包括名称、地址和事件队列等基本属性
type PeerProperty interface {
	// Name 返回 Peer 的名称
	// 名称用于标识和日志记录
	Name() string

	// Address 返回 Peer 的地址
	// 对于 Acceptor，返回监听地址
	// 对于 Connector，返回连接地址
	Address() string

	// Queue 返回 Peer 关联的事件队列
	// 事件队列用于异步处理网络事件
	Queue() EventQueue

	// SetName 设置 Peer 的名称
	// 名称用于标识和日志记录，可选设置
	SetName(v string)

	// SetAddress 设置 Peer 的地址
	// 对于 Acceptor，设置监听地址（如 "127.0.0.1:8080"）
	// 对于 Connector，设置连接地址
	SetAddress(v string)

	// SetQueue 设置 Peer 关联的事件队列
	// 事件队列用于异步处理网络事件，可选设置
	SetQueue(v EventQueue)
}

// GenericPeer 通用的 Peer 接口
// 组合了 Peer 和 PeerProperty 接口，提供完整的 Peer 功能
// 所有具体的 Peer 实现都应该实现此接口
type GenericPeer interface {
	Peer
	PeerProperty
}

// ContextSet 提供自定义属性存储和访问接口
// 允许为对象绑定任意类型的键值对，用于存储上下文信息
// 常用于存储用户数据、配置信息等
type ContextSet interface {
	// SetContext 为对象设置一个自定义属性
	// key: 属性的键，可以是任意类型
	// v: 属性的值，可以是任意类型
	SetContext(key interface{}, v interface{})

	// GetContext 从对象上根据 key 获取一个自定义属性
	// key: 属性的键
	// 返回属性值和是否存在
	GetContext(key interface{}) (interface{}, bool)

	// FetchContext 根据值的类型自动获取上下文并设置到值指针
	// key: 属性的键
	// valuePtr: 指向目标值的指针，类型会自动匹配
	// 返回是否成功获取并设置
	FetchContext(key, valuePtr interface{}) bool
}

// SessionAccessor 提供会话访问接口
// 用于管理和访问 Peer 下的所有 Session
// 主要用于服务器端管理多个客户端连接
type SessionAccessor interface {
	// GetSession 根据会话 ID 获取一个连接
	// sesID: 会话的唯一标识符
	// 返回对应的 Session，如果不存在返回 nil
	GetSession(sesID int64) Session

	// VisitSession 遍历所有连接
	// callback: 遍历回调函数，参数为 Session
	// 如果回调返回 false，则停止遍历
	VisitSession(callback func(Session) bool)

	// SessionCount 返回当前连接数量
	// 返回活跃的 Session 数量
	SessionCount() int

	// CloseAllSession 关闭所有连接
	// 断开所有活跃的 Session 连接
	CloseAllSession()
}

// PeerReadyChecker 提供 Peer 就绪状态检查接口
// 用于检查 Peer 是否已经准备好处理连接
// 对于 Acceptor，检查是否已开始监听
// 对于 Connector，检查是否已成功连接
type PeerReadyChecker interface {
	// IsReady 检查 Peer 是否正常工作
	// 返回 true 表示 Peer 已就绪，可以处理连接
	IsReady() bool
}

// PeerCaptureIOPanic 提供 IO 层异常捕获控制接口
// 用于控制是否捕获 IO 操作中的 panic，避免程序崩溃
// 在生产环境的对外端口应该启用此设置
type PeerCaptureIOPanic interface {
	// EnableCaptureIOPanic 开启或关闭 IO 层异常捕获
	// v: true 表示启用异常捕获，false 表示禁用
	EnableCaptureIOPanic(v bool)

	// CaptureIOPanic 获取当前异常捕获设置
	// 返回 true 表示已启用异常捕获
	CaptureIOPanic() bool
}
