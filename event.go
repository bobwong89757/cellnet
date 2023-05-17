package cellnet

// RecvMsgEvent
// @Description: 接收到消息
type RecvMsgEvent struct {
	Ses Session
	Msg interface{}
}

func (self *RecvMsgEvent) Session() Session {
	return self.Ses
}

func (self *RecvMsgEvent) Message() interface{} {
	return self.Msg
}

func (self *RecvMsgEvent) Send(msg interface{}) {
	self.Ses.Send(msg)
}

// Reply
//
//	@Description: 兼容relay和rpc的回消息接口
//	@receiver self
//	@param msg
func (self *RecvMsgEvent) Reply(msg interface{}) {
	self.Ses.Send(msg)
}

// SendMsgEvent
// @Description: 会话开始发送数据事件
type SendMsgEvent struct {
	Ses Session
	Msg interface{} // 用户需要发送的消息
}

func (self *SendMsgEvent) Message() interface{} {
	return self.Msg
}

func (self *SendMsgEvent) Session() Session {
	return self.Ses
}

// ReplyEvent
// @Description: rpc, relay, 普通消息
type ReplyEvent interface {
	Reply(msg interface{})
}
