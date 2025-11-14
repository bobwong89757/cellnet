package gogopb

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/gogo/protobuf/proto"
)

// gogopbCodec Gogo Protobuf 编码器实现
// 实现 cellnet.Codec 接口，用于 Protobuf 消息的编码/解码
// 使用 github.com/gogo/protobuf 库进行序列化
type gogopbCodec struct {
}

// Name 返回编码器的名称
// 返回 "gogopb"
func (self *gogopbCodec) Name() string {
	return "gogopb"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/x-protobuf"
func (self *gogopbCodec) MimeType() string {
	return "application/x-protobuf"
}

// Encode 将 Protobuf 消息编码为字节数组
// msgObj: 要编码的消息对象，必须实现 proto.Message 接口
// ctx: 上下文集合（未使用）
// 使用 gogo protobuf 进行序列化
// 返回编码后的字节数组
func (self *gogopbCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {

	return proto.Marshal(msgObj.(proto.Message))

}

// Decode 将字节数组解码为 Protobuf 消息
// data: 要解码的数据，必须是 []byte
// msgObj: 要解码到的消息对象（指针），必须实现 proto.Message 接口
// 使用 gogo protobuf 进行反序列化
func (self *gogopbCodec) Decode(data interface{}, msgObj interface{}) error {

	return proto.Unmarshal(data.([]byte), msgObj.(proto.Message))
}

// init 包初始化函数
// 自动注册 Gogo Protobuf 编码器
func init() {

	// 注册编码器
	codec.RegisterCodec(new(gogopbCodec))
}
