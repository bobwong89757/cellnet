package msglog

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
)

// PacketMessagePeeker
// @Description: 萃取消息中的内容
type PacketMessagePeeker interface {
	Message() interface{}
}

// WriteRecvLogger
//
//	@Description: 写入接收日志
//	@param protocol
//	@param ses
//	@param msg
func WriteRecvLogger(protocol string, ses cellnet.Session, msg interface{}) {

	if peeker, ok := msg.(PacketMessagePeeker); ok {
		msg = peeker.Message()
	}

	if IsMsgLogValid(cellnet.MessageToID(msg)) {
		peerInfo := ses.Peer().(cellnet.PeerProperty)

		log.GetLog().Debugf("#%s.recv(%s)@%d len: %d %s | %s",
			protocol,
			peerInfo.Name(),
			ses.ID(),
			cellnet.MessageSize(msg),
			cellnet.MessageToName(msg),
			cellnet.MessageToString(msg))
	}

}

// WriteSendLogger
//
//	@Description: 写入发送日志
//	@param protocol
//	@param ses
//	@param msg
func WriteSendLogger(protocol string, ses cellnet.Session, msg interface{}) {

	if peeker, ok := msg.(PacketMessagePeeker); ok {
		msg = peeker.Message()
	}

	if IsMsgLogValid(cellnet.MessageToID(msg)) {
		peerInfo := ses.Peer().(cellnet.PeerProperty)

		log.GetLog().Debugf("#%s.send(%s)@%d len: %d %s | %s",
			protocol,
			peerInfo.Name(),
			ses.ID(),
			cellnet.MessageSize(msg),
			cellnet.MessageToName(msg),
			cellnet.MessageToString(msg))
	}

}
