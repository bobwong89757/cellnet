package peer

import (
	"net"
	"time"
)

// CoreTCPSocketOption TCP Socket 选项核心实现
// 用于配置 TCP 连接的底层选项
// 包括缓冲区大小、Nagle 算法、超时时间等
type CoreTCPSocketOption struct {
	// readBufferSize 接收缓冲区大小
	// -1 表示使用系统默认值，>= 0 表示自定义大小
	readBufferSize int

	// writeBufferSize 发送缓冲区大小
	// -1 表示使用系统默认值，>= 0 表示自定义大小
	writeBufferSize int

	// noDelay 是否禁用 Nagle 算法
	// true 表示禁用（立即发送），false 表示启用（延迟发送以合并小包）
	noDelay bool

	// maxPacketSize 最大封包大小
	// 超过此大小的消息会被拒绝
	maxPacketSize int

	// readTimeout 读取超时时间
	// 0 表示不超时，> 0 表示读取操作的超时时间
	readTimeout time.Duration

	// writeTimeout 写入超时时间
	// 0 表示不超时，> 0 表示写入操作的超时时间
	writeTimeout time.Duration
}

// SetSocketBuffer 设置 Socket 缓冲区大小和 Nagle 算法
// readBufferSize: 接收缓冲区大小，-1 表示使用系统默认值
// writeBufferSize: 发送缓冲区大小，-1 表示使用系统默认值
// noDelay: 是否禁用 Nagle 算法，true 表示禁用
func (self *CoreTCPSocketOption) SetSocketBuffer(readBufferSize, writeBufferSize int, noDelay bool) {
	self.readBufferSize = readBufferSize
	self.writeBufferSize = writeBufferSize
	self.noDelay = noDelay
}

// SetSocketDeadline 设置 Socket 读写超时时间
// read: 读取超时时间，0 表示不超时
// write: 写入超时时间，0 表示不超时
func (self *CoreTCPSocketOption) SetSocketDeadline(read, write time.Duration) {

	self.readTimeout = read
	self.writeTimeout = write
}

// SetMaxPacketSize 设置最大封包大小
// maxSize: 最大封包大小（字节），超过此大小的消息会被拒绝
func (self *CoreTCPSocketOption) SetMaxPacketSize(maxSize int) {
	self.maxPacketSize = maxSize
}

// MaxPacketSize 获取最大封包大小
// 返回当前设置的最大封包大小
func (self *CoreTCPSocketOption) MaxPacketSize() int {

	return self.maxPacketSize
}

// ApplySocketOption 应用 Socket 选项到连接
// conn: TCP 连接对象
// 设置缓冲区大小和 Nagle 算法
func (self *CoreTCPSocketOption) ApplySocketOption(conn net.Conn) {

	if cc, ok := conn.(*net.TCPConn); ok {

		// 设置接收缓冲区大小
		if self.readBufferSize >= 0 {
			cc.SetReadBuffer(self.readBufferSize)
		}

		// 设置发送缓冲区大小
		if self.writeBufferSize >= 0 {
			cc.SetWriteBuffer(self.writeBufferSize)
		}

		// 设置 Nagle 算法
		cc.SetNoDelay(self.noDelay)
	}

}

// ApplySocketReadTimeout 应用读取超时并执行回调
// conn: 网络连接对象
// callback: 要执行的回调函数
// 如果设置了读取超时，会在回调执行前后设置和清除超时
// 参考: http://blog.GetLog().sina.com.cn/s/blog_9be3b8f10101lhiq.html
func (self *CoreTCPSocketOption) ApplySocketReadTimeout(conn net.Conn, callback func()) {

	if self.readTimeout > 0 {
		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(self.readTimeout))
		callback()
		// 清除读取超时（设置为零值表示不超时）
		conn.SetReadDeadline(time.Time{})

	} else {
		// 没有设置超时，直接执行回调
		callback()
	}
}

// ApplySocketWriteTimeout 应用写入超时并执行回调
// conn: 网络连接对象
// callback: 要执行的回调函数
// 如果设置了写入超时，会在回调执行前后设置和清除超时
func (self *CoreTCPSocketOption) ApplySocketWriteTimeout(conn net.Conn, callback func()) {

	if self.writeTimeout > 0 {
		// 设置写入超时
		conn.SetWriteDeadline(time.Now().Add(self.writeTimeout))
		callback()
		// 清除写入超时（设置为零值表示不超时）
		conn.SetWriteDeadline(time.Time{})

	} else {
		// 没有设置超时，直接执行回调
		callback()
	}
}

// Init 初始化 Socket 选项
// 将缓冲区大小设置为 -1（使用系统默认值）
func (self *CoreTCPSocketOption) Init() {
	self.readBufferSize = -1
	self.writeBufferSize = -1
}
