package cellnet

import "fmt"

// SessionInit 表示会话初始化事件
// 在 Session 创建时触发，用于初始化会话
type SessionInit struct {
}

// SessionAccepted 表示服务器接受新连接的事件
// 当服务器端（Acceptor）接受一个新的客户端连接时触发
// 在用户回调中可以通过此事件处理新连接
type SessionAccepted struct {
}

// SessionConnected 表示客户端连接成功的事件
// 当客户端（Connector）成功连接到服务器时触发
// 在用户回调中可以通过此事件处理连接成功后的逻辑
type SessionConnected struct {
}

// SessionConnectError 表示连接错误的事件
// 当客户端连接失败时触发
// 在用户回调中可以通过此事件处理连接失败的情况
type SessionConnectError struct {
}

// CloseReason 表示连接关闭的原因
type CloseReason int32

const (
	// CloseReason_IO 表示普通的 IO 断开
	// 通常是由于网络错误、对端关闭等原因导致的断开
	CloseReason_IO CloseReason = iota

	// CloseReason_Manual 表示手动关闭
	// 在关闭前调用过 Session.Close() 方法
	CloseReason_Manual
)

// String 返回关闭原因的字符串表示
// 用于日志记录和调试
func (self CloseReason) String() string {
	switch self {
	case CloseReason_IO:
		return "IO"
	case CloseReason_Manual:
		return "Manual"
	}

	return "Unknown"
}

// SessionClosed 表示会话关闭的事件
// 当 Session 连接断开时触发，包含断开原因
type SessionClosed struct {
	// Reason 断开原因
	// 可以是 CloseReason_IO（IO 断开）或 CloseReason_Manual（手动关闭）
	Reason CloseReason
}

// SessionCloseNotify 用于 UDP 通知关闭，内部使用
// 用于 UDP 协议的特殊关闭通知机制
type SessionCloseNotify struct {
}

// String 方法实现 fmt.Stringer 接口，用于格式化输出
func (self *SessionInit) String() string         { return fmt.Sprintf("%+v", *self) }
func (self *SessionAccepted) String() string     { return fmt.Sprintf("%+v", *self) }
func (self *SessionConnected) String() string    { return fmt.Sprintf("%+v", *self) }
func (self *SessionConnectError) String() string { return fmt.Sprintf("%+v", *self) }
func (self *SessionClosed) String() string       { return fmt.Sprintf("%+v", *self) }
func (self *SessionCloseNotify) String() string  { return fmt.Sprintf("%+v", *self) }

// SystemMessage 方法标记这些消息为系统消息
// 系统消息是框架内部使用的消息，不会通过正常的消息注册流程
// 可以通过类型断言 SystemMessageIdentifier 来判断是否为系统消息
func (self *SessionInit) SystemMessage()         {}
func (self *SessionAccepted) SystemMessage()     {}
func (self *SessionConnected) SystemMessage()    {}
func (self *SessionConnectError) SystemMessage() {}
func (self *SessionClosed) SystemMessage()       {}
func (self *SessionCloseNotify) SystemMessage()  {}

// SystemMessageIdentifier 是系统消息的标识接口
// 所有系统消息都实现了此接口
// 可以使用类型断言来判断一个消息是否为系统消息：
//   if _, ok := msg.(SystemMessageIdentifier); ok {
//       // 这是系统消息
//   }
type SystemMessageIdentifier interface {
	SystemMessage()
}
