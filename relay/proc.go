package relay

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/msglog"
)

// PassthroughContent 表示透传内容
// 用于在 relay 消息中传递额外的数据
type PassthroughContent struct {
	// Int64 透传的 int64 数据
	Int64 int64

	// Int64Slice 透传的 int64 切片数据
	Int64Slice []int64

	// Str 透传的字符串数据
	Str string
}

// ResoleveInboundEvent 处理入站的 relay 消息
// inputEvent: 输入的接收事件
// 返回处理后的输出事件、是否已处理、错误信息
// 如果消息是 RelayACK 类型，会解码消息并调用广播函数
func ResoleveInboundEvent(inputEvent cellnet.Event) (ouputEvent cellnet.Event, handled bool, err error) {
	switch relayMsg := inputEvent.Message().(type) {
	case *RelayACK:
		// 创建接收事件
		ev := &RecvMsgEvent{
			Ses: inputEvent.Session(),
			ack: relayMsg,
		}

		// 如果有消息 ID，解码消息
		if relayMsg.MsgID != 0 {
			ev.Msg, _, err = codec.DecodeMessage(int(relayMsg.MsgID), relayMsg.Msg)
			if err != nil {
				return
			}
		}

		// 如果消息日志有效，记录接收日志
		if msglog.IsMsgLogValid(int(relayMsg.MsgID)) {
			peerInfo := inputEvent.Session().Peer().(cellnet.PeerProperty)

			log.GetLog().Debugf("#relay.recv(%s)@%d len: %d %s {%s}| %s",
				peerInfo.Name(),
				inputEvent.Session().ID(),
				cellnet.MessageSize(ev.Message()),
				cellnet.MessageToName(ev.Message()),
				cellnet.MessageToString(relayMsg),
				cellnet.MessageToString(ev.Message()))
		}

		// 如果有广播函数，在队列中调用
		if bcFunc != nil {
			// 转到对应线程中调用，保证线程安全
			cellnet.SessionQueuedCall(inputEvent.Session(), func() {
				bcFunc(ev)
			})
		}

		return ev, true, nil
	}

	// 不是 relay 消息，不处理
	return inputEvent, false, nil
}

// ResolveOutboundEvent 处理 relay.Relay 出站消息的日志
// inputEvent: 输入的发送事件
// 返回是否已处理、错误信息
// 如果消息是 RelayACK 类型，会记录发送日志
func ResolveOutboundEvent(inputEvent cellnet.Event) (handled bool, err error) {
	switch relayMsg := inputEvent.Message().(type) {
	case *RelayACK:
		// 如果消息日志有效，记录发送日志
		if msglog.IsMsgLogValid(int(relayMsg.MsgID)) {
			var payload interface{}
			// 如果有消息 ID，解码消息用于日志
			if relayMsg.MsgID != 0 {
				payload, _, err = codec.DecodeMessage(int(relayMsg.MsgID), relayMsg.Msg)
				if err != nil {
					return
				}
			}

			peerInfo := inputEvent.Session().Peer().(cellnet.PeerProperty)

			log.GetLog().Debugf("#relay.send(%s)@%d len: %d %s {%s}| %s",
				peerInfo.Name(),
				inputEvent.Session().ID(),
				cellnet.MessageSize(payload),
				cellnet.MessageToName(payload),
				cellnet.MessageToString(relayMsg),
				cellnet.MessageToString(payload))
		}

		return true, nil
	}

	return
}
