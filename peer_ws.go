package cellnet

import "time"

// WSAcceptor 定义 WebSocket 接受器接口
// 用于创建 WebSocket 服务器，接受客户端连接
// 具备会话访问功能，可以管理多个客户端连接
type WSAcceptor interface {
	GenericPeer

	// SessionAccessor 访问会话
	// 可以获取、遍历和管理所有客户端连接
	SessionAccessor

	// SetHttps 设置 HTTPS 证书
	// certfile: 证书文件路径
	// keyfile: 私钥文件路径
	// 用于启用 WSS（WebSocket Secure）支持
	SetHttps(certfile, keyfile string)

	// SetUpgrader 设置 WebSocket 升级器
	// upgrader: WebSocket 升级器对象
	// 用于自定义 WebSocket 连接升级逻辑
	SetUpgrader(upgrader interface{})

	// Port 查看当前侦听端口
	// 返回当前监听的端口号
	// 如果使用 "host:0" 作为 Address，socket 底层会自动分配侦听端口
	Port() int
}

// WSConnector 定义 WebSocket 连接器接口
// 用于创建 WebSocket 客户端，连接到服务器
type WSConnector interface {
	GenericPeer

	// SetReconnectDuration 设置重连时间间隔
	// 当连接断开时，会在此时间后尝试重新连接
	SetReconnectDuration(time.Duration)

	// ReconnectDuration 获取重连时间间隔
	// 返回当前设置的重连时间间隔
	ReconnectDuration() time.Duration

	// Session 获取默认会话
	// 返回当前连接的 Session，如果未连接返回 nil
	Session() Session

	// SetSessionManager 设置会话管理器
	// raw: 实现 peer.SessionManager 接口的对象
	// 用于自定义会话管理逻辑
	SetSessionManager(raw interface{})

	// Port 查看当前连接使用的端口
	// 返回本地端口号
	Port() int
}
