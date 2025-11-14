package tcp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/util"
	"io"
	"net"
)

// TCPMessageTransmitter TCP 消息传输器
// 实现 MessageTransmitter 接口，负责 TCP 消息的接收和发送
// 使用 LTV（Length-Type-Value）格式进行消息封装
type TCPMessageTransmitter struct {
}

// socketOpt Socket 选项接口
// 用于获取 Socket 配置选项
type socketOpt interface {
	// MaxPacketSize 获取最大封包大小
	MaxPacketSize() int
	// ApplySocketReadTimeout 应用读取超时并执行回调
	ApplySocketReadTimeout(conn net.Conn, callback func())
	// ApplySocketWriteTimeout 应用写入超时并执行回调
	ApplySocketWriteTimeout(conn net.Conn, callback func())
}

// OnRecvMessage 接收消息
// ses: 会话对象
// 从 TCP 连接读取 LTV 格式的数据包并解码为消息
// 支持读取超时配置
// 返回解码后的消息和错误
func (TCPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {

	// 获取原始连接的 Reader 接口
	reader, ok := ses.Raw().(io.Reader)

	// 转换错误，或者连接已经关闭时退出
	if !ok || reader == nil {
		return nil, nil
	}

	// 获取 Socket 选项
	opt := ses.Peer().(socketOpt)

	// 转换为网络连接以应用超时
	if conn, ok := reader.(net.Conn); ok {

		// 有读超时时，设置超时
		opt.ApplySocketReadTimeout(conn, func() {
			// 接收 LTV 格式的数据包并解码
			msg, err = util.RecvLTVPacket(reader, opt.MaxPacketSize())

		})
	}

	return
}

// OnSendMessage 发送消息
// ses: 会话对象
// msg: 要发送的消息
// 将消息编码为 LTV 格式并写入 TCP 连接
// 支持写入超时配置
// 返回发送错误
func (TCPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) (err error) {

	// 获取原始连接的 Writer 接口
	writer, ok := ses.Raw().(io.Writer)

	// 转换错误，或者连接已经关闭时退出
	if !ok || writer == nil {
		return nil
	}

	// 获取 Socket 选项
	opt := ses.Peer().(socketOpt)

	// 有写超时时，设置超时
	opt.ApplySocketWriteTimeout(writer.(net.Conn), func() {
		// 编码消息为 LTV 格式并发送
		err = util.SendLTVPacket(writer, ses.(cellnet.ContextSet), msg)

	})

	return
}
