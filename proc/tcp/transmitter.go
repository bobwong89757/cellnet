package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/util"
	"io"
	"net"
)

// TCPMessageTransmitter
// @Description: TCP消息传输器
type TCPMessageTransmitter struct {
}

type socketOpt interface {
	MaxPacketSize() int
	ApplySocketReadTimeout(conn net.Conn, callback func())
	ApplySocketWriteTimeout(conn net.Conn, callback func())
}

// OnRecvMessage
//
//	@Description: 接收消息
//	@receiver TCPMessageTransmitter
//	@param ses
//	@return msg
//	@return err
func (TCPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {

	reader, ok := ses.Raw().(io.Reader)

	// 转换错误，或者连接已经关闭时退出
	if !ok || reader == nil {
		return nil, nil
	}

	opt := ses.Peer().(socketOpt)

	if conn, ok := reader.(net.Conn); ok {

		// 有读超时时，设置超时
		opt.ApplySocketReadTimeout(conn, func() {

			msg, err = util.RecvLTVPacket(reader, opt.MaxPacketSize())

		})
	}

	return
}

// OnSendMessage
//
//	@Description: 发送消息
//	@receiver TCPMessageTransmitter
//	@param ses
//	@param msg
//	@return err
func (TCPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) (err error) {

	writer, ok := ses.Raw().(io.Writer)

	// 转换错误，或者连接已经关闭时退出
	if !ok || writer == nil {
		return nil
	}

	opt := ses.Peer().(socketOpt)

	// 有写超时时，设置超时
	opt.ApplySocketWriteTimeout(writer.(net.Conn), func() {

		err = util.SendLTVPacket(writer, ses.(cellnet.ContextSet), msg)

	})

	return
}
