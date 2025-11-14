package cellnet

// Session 表示一个长连接会话
// Session 是网络通信的基本单位，代表一个客户端与服务器之间的连接
// 每个 Session 都有唯一的 ID，用于标识和管理
//
// Session 提供消息发送、连接关闭等基本操作
// 消息发送是异步的，内部维护发送队列
type Session interface {
	// Raw 获得原始的 Socket 连接
	// 返回底层网络连接对象，类型取决于协议（如 *net.TCPConn、*net.UDPConn 等）
	// 可用于获取连接详细信息或进行底层操作
	Raw() interface{}

	// Peer 获得 Session 归属的 Peer
	// 返回创建此 Session 的 Peer 对象
	Peer() Peer

	// Send 发送消息
	// msg: 要发送的消息对象，需要以指针格式传入（如 &MyMessage{}）
	// 消息发送是异步的，会先放入发送队列，由发送循环处理
	// 如果连接已关闭，消息会被丢弃
	Send(msg interface{})

	// Close 关闭连接
	// 断开当前 Session 的连接，停止接收和发送消息
	// 关闭后，相关的 goroutine 会退出
	Close()

	// ID 返回 Session 的唯一标识符
	// 每个 Session 都有一个唯一的 64 位整数 ID
	// 可用于在多个 Session 中识别特定的连接
	ID() int64
}

// RawPacket 用于直接发送原始数据包
// 当需要发送已编码的字节数组时，可以将 *RawPacket 作为 Send 参数
// 常用于转发消息或发送自定义格式的数据
type RawPacket struct {
	// MsgData 消息的原始字节数据
	// 包含已编码的消息内容
	MsgData []byte

	// MsgID 消息的 ID
	// 用于标识消息类型，必须与已注册的消息 ID 匹配
	MsgID int
}

// Message 将 RawPacket 解码为消息对象
// 根据 MsgID 查找消息元信息，然后使用对应的 Codec 解码数据
// 如果消息未注册或解码失败，返回空结构体
func (self *RawPacket) Message() interface{} {
	// 根据消息 ID 获取消息元信息
	meta := MessageMetaByID(self.MsgID)

	// 消息没有注册，返回空结构体
	if meta == nil {
		return struct{}{}
	}

	// 创建消息类型的实例
	msg := meta.NewType()

	// 使用对应的 Codec 从字节数组解码为消息对象
	err := meta.Codec.Decode(self.MsgData, msg)
	if err != nil {
		// 解码失败，返回空结构体
		return struct{}{}
	}

	return msg
}
