package cellnet

// RecvMsgEvent 表示接收到消息的事件
// 当从网络连接接收到消息时，会创建此事件并传递给用户回调
type RecvMsgEvent struct {
	// Ses 接收消息的 Session
	Ses Session

	// Msg 接收到的消息对象
	Msg interface{}
}

// Session 返回事件对应的 Session
func (self *RecvMsgEvent) Session() Session {
	return self.Ses
}

// Message 返回事件携带的消息
func (self *RecvMsgEvent) Message() interface{} {
	return self.Msg
}

// Send 发送消息到对应的 Session
// 这是一个便捷方法，等同于 self.Ses.Send(msg)
func (self *RecvMsgEvent) Send(msg interface{}) {
	self.Ses.Send(msg)
}

// Reply 回复消息到对应的 Session
// 兼容 relay 和 rpc 的回消息接口
// 与 Send 方法功能相同，提供更语义化的接口
func (self *RecvMsgEvent) Reply(msg interface{}) {
	self.Ses.Send(msg)
}

// SendMsgEvent 表示会话开始发送数据的事件
// 当调用 Session.Send() 发送消息时，会创建此事件
// 此事件会经过 EventHooker 处理，然后由 MessageTransmitter 发送
type SendMsgEvent struct {
	// Ses 要发送消息的 Session
	Ses Session

	// Msg 用户需要发送的消息对象
	Msg interface{}
}

// Message 返回要发送的消息
func (self *SendMsgEvent) Message() interface{} {
	return self.Msg
}

// Session 返回事件对应的 Session
func (self *SendMsgEvent) Session() Session {
	return self.Ses
}

// ReplyEvent 定义回复消息接口
// 用于 RPC、Relay 和普通消息的统一回复接口
// 实现了此接口的事件可以直接调用 Reply 方法回复消息
type ReplyEvent interface {
	// Reply 回复消息
	// msg: 要回复的消息对象
	Reply(msg interface{})
}
