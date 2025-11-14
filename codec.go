package cellnet

// Codec 定义消息编解码器接口
// 用于将消息对象序列化为字节数组，或将字节数组反序列化为消息对象
// cellnet 支持多种编码格式：Protobuf、JSON、Binary、ProtoPlus、Sproto 等
//
// 实现 Codec 接口可以扩展支持新的编码格式
type Codec interface {
	// Encode 将消息对象编码为字节数组
	// msgObj: 要编码的消息对象，通常是指针类型
	// ctx: 上下文信息，可用于传递编码相关的配置或资源
	// 返回编码后的数据和错误信息
	Encode(msgObj interface{}, ctx ContextSet) (data interface{}, err error)

	// Decode 将字节数组解码为消息对象
	// data: 要解码的字节数组
	// msgObj: 目标消息对象，通常是指针类型，解码结果会写入此对象
	// 返回解码错误，如果成功则返回 nil
	Decode(data interface{}, msgObj interface{}) error

	// Name 返回编码器的名称
	// 用于标识编码器类型，如 "protobuf"、"json"、"binary" 等
	Name() string

	// MimeType 返回编码器的 MIME 类型
	// 主要用于 HTTP 协议兼容，如 "application/json"、"application/protobuf" 等
	MimeType() string
}
