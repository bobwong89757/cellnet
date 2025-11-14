package rpc

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
)

// RecvMsgEvent RPC 接收消息事件
// 用于服务端处理 RPC 请求
// 包含调用 ID，用于关联请求和响应
type RecvMsgEvent struct {
	// ses 会话对象
	ses cellnet.Session

	// Msg 接收到的消息
	Msg interface{}

	// callid 调用 ID
	// 用于关联请求和响应
	callid int64
}

// Session 获取会话对象
// 返回事件关联的会话
func (self *RecvMsgEvent) Session() cellnet.Session {
	return self.ses
}

// Message 获取消息对象
// 返回接收到的消息
func (self *RecvMsgEvent) Message() interface{} {
	return self.Msg
}

// Queue 获取事件队列
// 返回会话所属 Peer 的事件队列
func (self *RecvMsgEvent) Queue() cellnet.EventQueue {
	return self.ses.Peer().(interface {
		Queue() cellnet.EventQueue
	}).Queue()
}

// Reply 回复消息
// msg: 要回复的消息对象
// 将消息编码后通过 RemoteCallACK 发送，使用调用 ID 关联请求和响应
func (self *RecvMsgEvent) Reply(msg interface{}) {

	// 编码消息
	data, meta, err := codec.EncodeMessage(msg, nil)

	if err != nil {
		log.GetLog().Errorf("rpc reply message encode error: %s", err)
		return
	}

	// 发送 RPC 响应，使用调用 ID 关联
	self.ses.Send(&RemoteCallACK{
		MsgID:  uint32(meta.ID),
		Data:   data,
		CallID: self.callid,
	})
}
