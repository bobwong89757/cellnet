package udp

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/msglog"
	"github.com/bobwong89757/cellnet/peer/udp"
	"github.com/bobwong89757/cellnet/proc"
)

// UDPMessageTransmitter UDP 消息传输器
// 实现 MessageTransmitter 接口，负责 UDP 消息的接收和发送
// 使用自定义的 UDP 数据包格式进行消息封装
type UDPMessageTransmitter struct {
}

// OnRecvMessage 接收消息
// ses: 会话对象
// 从 UDP Session 读取数据包并解码为消息
// 记录接收日志
// 返回解码后的消息和错误
func (UDPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {

	// 从 Session 读取原始数据
	data := ses.Raw().(udp.DataReader).ReadData()

	// 解码数据包为消息
	msg, err = RecvPacket(data)

	// 记录接收日志
	msglog.WriteRecvLogger("udp", ses, msg)

	return
}

// OnSendMessage 发送消息
// ses: 会话对象
// msg: 要发送的消息
// 将消息编码为 UDP 数据包格式并发送
// 记录发送日志
// 返回发送错误
func (UDPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {

	// 获取数据写入器
	writer := ses.(udp.DataWriter)

	// 记录发送日志
	msglog.WriteSendLogger("udp", ses, msg)

	// Session 不再被复用，所以使用 Session 自己的 ContextSet 做内存池，避免串台
	return sendPacket(writer, ses.(cellnet.ContextSet), msg)
}

// init 包初始化函数
// 自动注册 UDP LTV 消息处理器
// 当调用 proc.BindProcessorHandler(peer, "udp.ltv", callback) 时会使用此处理器
func init() {
	// 注册 udp.ltv 包处理器
	proc.RegisterProcessor("udp.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		// 设置消息传输器，负责消息的编码、解码和网络传输
		bundle.SetTransmitter(new(UDPMessageTransmitter))
		// 设置事件回调（UDP 不使用队列化回调，直接使用用户回调）
		bundle.SetCallback(userCallback)

	})
}
