package relay

import (
	"github.com/bobwong89757/cellnet"
)

// RecvMsgEvent 表示接收到的 relay 消息事件
// 实现了 cellnet.Event 接口，用于处理 relay 消息
type RecvMsgEvent struct {
	// Ses 接收消息的 Session
	Ses cellnet.Session

	// ack 接收到的 RelayACK 消息
	ack *RelayACK

	// Msg 解码后的消息对象
	Msg interface{}
}

// PassThroughAsInt64 获取透传的 int64 数据
// 返回透传的 int64 值，如果没有则返回 0
func (self *RecvMsgEvent) PassThroughAsInt64() int64 {
	if self.ack == nil {
		return 0
	}

	return self.ack.Int64
}

// PassThroughAsInt64Slice 获取透传的 int64 切片数据
// 返回透传的 int64 切片，如果没有则返回 nil
func (self *RecvMsgEvent) PassThroughAsInt64Slice() []int64 {
	if self.ack == nil {
		return nil
	}

	return self.ack.Int64Slice
}

// PassThroughAsString 获取透传的字符串数据
// 返回透传的字符串，如果没有则返回空字符串
func (self *RecvMsgEvent) PassThroughAsString() string {
	if self.ack == nil {
		return ""
	}

	return self.ack.Str
}

// Session 返回事件对应的 Session
// 实现 cellnet.Event 接口
func (self *RecvMsgEvent) Session() cellnet.Session {
	return self.Ses
}

// Message 返回事件携带的消息
// 实现 cellnet.Event 接口
func (self *RecvMsgEvent) Message() interface{} {
	return self.Msg
}

// Reply 消息原路返回
// msg: 要回复的消息对象
// 会将消息和透传数据一起发送回原 Session
// 注意：没填的值不会被发送
func (self *RecvMsgEvent) Reply(msg interface{}) {
	// 没填的值不会被发送
	Relay(self.Ses, msg, self.ack.Int64, self.ack.Int64Slice, self.ack.Str)
}
