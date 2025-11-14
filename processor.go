package cellnet

// Event 表示一个网络事件
// Event 是 cellnet 事件驱动模型的核心，所有网络操作都通过 Event 传递
// 包括消息接收、连接建立、连接断开等事件
type Event interface {
	// Session 返回事件对应的会话
	// 每个事件都关联一个 Session，表示事件发生的连接
	Session() Session

	// Message 返回事件携带的消息
	// 对于消息接收事件，返回接收到的消息对象
	// 对于系统事件（如连接建立、断开），返回对应的系统消息
	Message() interface{}
}

// MessageTransmitter 定义消息传输器接口
// 负责从网络连接读取消息和向网络连接写入消息
// 不同的协议（TCP、UDP、WebSocket 等）有不同的实现
type MessageTransmitter interface {
	// OnRecvMessage 从 Session 接收消息
	// ses: 要读取消息的 Session
	// 返回接收到的消息对象和错误信息
	// 如果连接关闭，返回 nil, nil
	OnRecvMessage(ses Session) (msg interface{}, err error)

	// OnSendMessage 向 Session 发送消息
	// ses: 要发送消息的 Session
	// msg: 要发送的消息对象
	// 返回发送错误，如果成功则返回 nil
	OnSendMessage(ses Session, msg interface{}) error
}

// EventHooker 定义事件钩子接口
// 用于在消息处理流程中插入自定义逻辑
// 可以在消息接收和发送前后进行处理，如日志记录、消息过滤、加密解密等
//
// 注意：如果不想让消息继续处理，可以在钩子中将 Event 设置为 nil
type EventHooker interface {
	// OnInboundEvent 处理入站（接收）事件
	// input: 输入的接收事件
	// 返回处理后的输出事件，如果返回 nil 则停止处理
	OnInboundEvent(input Event) (output Event)

	// OnOutboundEvent 处理出站（发送）事件
	// input: 输入的发送事件
	// 返回处理后的输出事件，如果返回 nil 则停止发送
	OnOutboundEvent(input Event) (output Event)
}

// EventCallback 是用户定义的事件处理回调函数
// 当消息到达或系统事件发生时，会调用此回调函数
// ev: 触发的事件，包含 Session 和 Message 信息
type EventCallback func(ev Event)
