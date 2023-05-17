package kcp

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/msglog"
	"github.com/bobwong89757/cellnet/peer/kcp"
	"github.com/bobwong89757/cellnet/proc"
)

// KCPMessageTransmitter
// @Description: KCP消息传输器
type KCPMessageTransmitter struct {
}

// OnRecvMessage
//
//	@Description: 接收消息
//	@receiver KCPMessageTransmitter
//	@param ses
//	@return msg
//	@return err
func (KCPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {

	data := ses.Raw().(kcp.DataReader).ReadData()

	if data == nil {
		return
	}

	msg, err = RecvPacket(data)

	msglog.WriteRecvLogger("kcp", ses, msg)

	return
}

// OnSendMessage
//
//	@Description: 发送消息
//	@receiver KCPMessageTransmitter
//	@param ses
//	@param msg
//	@return error
func (KCPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {

	writer := ses.(kcp.DataWriter)

	msglog.WriteSendLogger("kcp", ses, msg)

	// ses不再被复用, 所以使用session自己的contextset做内存池, 避免串台
	return SendPacket(writer, ses.(cellnet.ContextSet), msg)
}

func init() {
	var convid uint32
	binary.Read(rand.Reader, binary.LittleEndian, &convid)
	proc.RegisterProcessor("kcp.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {
		bundle.SetTransmitter(new(KCPMessageTransmitter))
		bundle.SetCallback(userCallback)

	})
}
