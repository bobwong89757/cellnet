package cellnet

import "time"

// TCPSocketOption 定义 TCP Socket 选项接口
// 用于配置 TCP 连接的底层选项
type TCPSocketOption interface {
	// SetSocketBuffer 设置收发缓冲区大小
	// readBufferSize: 接收缓冲区大小，默认 -1 表示使用系统默认值
	// writeBufferSize: 发送缓冲区大小，默认 -1 表示使用系统默认值
	// noDelay: 是否禁用 Nagle 算法，true 表示禁用（立即发送），false 表示启用（延迟发送）
	SetSocketBuffer(readBufferSize, writeBufferSize int, noDelay bool)

	// SetMaxPacketSize 设置最大的封包大小
	// maxSize: 最大封包大小（字节），超过此大小的消息会被拒绝
	SetMaxPacketSize(maxSize int)

	// SetSocketDeadline 设置读写超时时间
	// read: 读取超时时间，默认 0 表示不超时
	// write: 写入超时时间，默认 0 表示不超时
	SetSocketDeadline(read, write time.Duration)
}

// TCPAcceptor 定义 TCP 接受器接口
// 用于创建 TCP 服务器，接受客户端连接
// 具备会话访问功能，可以管理多个客户端连接
type TCPAcceptor interface {
	GenericPeer

	// SessionAccessor 访问会话
	// 可以获取、遍历和管理所有客户端连接
	SessionAccessor

	// TCPSocketOption TCP Socket 选项
	// 可以配置 TCP 连接的底层选项
	TCPSocketOption

	// Port 查看当前侦听端口
	// 返回当前监听的端口号
	// 如果使用 "host:0" 作为 Address，socket 底层会自动分配侦听端口
	Port() int
}

// TCPConnector 定义 TCP 连接器接口
// 用于创建 TCP 客户端，连接到服务器
type TCPConnector interface {
	GenericPeer

	// TCPSocketOption TCP Socket 选项
	// 可以配置 TCP 连接的底层选项
	TCPSocketOption

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
