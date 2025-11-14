package msglog

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
)

// PacketMessagePeeker 定义消息内容提取接口
// 用于从包装的消息中提取实际的消息内容
// 例如，从 RawPacket 中提取实际的消息对象
type PacketMessagePeeker interface {
	// Message 提取消息中的实际内容
	Message() interface{}
}

// WriteRecvLogger 写入接收消息的日志
// protocol: 协议名称，如 "tcp"、"udp"、"kcp" 等
// ses: 接收消息的 Session
// msg: 接收到的消息对象
// 如果消息实现了 PacketMessagePeeker 接口，会提取实际消息内容
// 如果消息日志有效，会记录详细的接收日志
func WriteRecvLogger(protocol string, ses cellnet.Session, msg interface{}) {
	// 如果消息实现了 PacketMessagePeeker 接口，提取实际消息内容
	if peeker, ok := msg.(PacketMessagePeeker); ok {
		msg = peeker.Message()
	}

	// 检查消息日志是否有效
	if IsMsgLogValid(cellnet.MessageToID(msg)) {
		peerInfo := ses.Peer().(cellnet.PeerProperty)

		// 记录接收日志
		log.GetLog().Debugf("#%s.recv(%s)@%d len: %d %s | %s",
			protocol,
			peerInfo.Name(),
			ses.ID(),
			cellnet.MessageSize(msg),
			cellnet.MessageToName(msg),
			cellnet.MessageToString(msg))
	}
}

// WriteSendLogger 写入发送消息的日志
// protocol: 协议名称，如 "tcp"、"udp"、"kcp" 等
// ses: 发送消息的 Session
// msg: 要发送的消息对象
// 如果消息实现了 PacketMessagePeeker 接口，会提取实际消息内容
// 如果消息日志有效，会记录详细的发送日志
func WriteSendLogger(protocol string, ses cellnet.Session, msg interface{}) {
	// 如果消息实现了 PacketMessagePeeker 接口，提取实际消息内容
	if peeker, ok := msg.(PacketMessagePeeker); ok {
		msg = peeker.Message()
	}

	// 检查消息日志是否有效
	if IsMsgLogValid(cellnet.MessageToID(msg)) {
		peerInfo := ses.Peer().(cellnet.PeerProperty)

		// 记录发送日志
		log.GetLog().Debugf("#%s.send(%s)@%d len: %d %s | %s",
			protocol,
			peerInfo.Name(),
			ses.ID(),
			cellnet.MessageSize(msg),
			cellnet.MessageToName(msg),
			cellnet.MessageToString(msg))
	}
}
