package binary

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/goobjfmt"
)

// binaryCodec 实现二进制编码器
// 使用 goobjfmt 库进行内存流的直接序列化
// 适用于服务器内网传输，具有低 GC 和低实现复杂度的优点
type binaryCodec struct {
}

// Name 返回编码器的名称
// 返回 "binary"
func (self *binaryCodec) Name() string {
	return "binary"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/binary"
func (self *binaryCodec) MimeType() string {
	return "application/binary"
}

// Encode 将消息对象编码为二进制字节数组
// msgObj: 要编码的消息对象
// ctx: 上下文信息（此编码器不使用）
// 返回编码后的字节数组和错误信息
// 使用 goobjfmt.BinaryWrite 进行二进制序列化
func (self *binaryCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {
	return goobjfmt.BinaryWrite(msgObj)
}

// Decode 将二进制字节数组解码为消息对象
// data: 要解码的字节数组
// msgObj: 目标消息对象，解码结果会写入此对象
// 返回解码错误，如果成功则返回 nil
// 使用 goobjfmt.BinaryRead 进行二进制反序列化
func (self *binaryCodec) Decode(data interface{}, msgObj interface{}) error {
	return goobjfmt.BinaryRead(data.([]byte), msgObj)
}

// init 在包加载时自动注册二进制编码器
// 将 binaryCodec 注册到全局编码器列表中
func init() {
	codec.RegisterCodec(new(binaryCodec))
}
