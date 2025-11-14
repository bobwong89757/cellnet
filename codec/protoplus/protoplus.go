package protoplus

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/protoplus/proto"
)

// protoplus ProtoPlus 编码器实现
// 实现 cellnet.Codec 接口，用于 ProtoPlus 消息的编码/解码
// 使用 github.com/bobwong89757/protoplus 库进行序列化
type protoplus struct {
}

// Name 返回编码器的名称
// 返回 "protoplus"
func (self *protoplus) Name() string {
	return "protoplus"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/binary"
func (self *protoplus) MimeType() string {
	return "application/binary"
}

// Encode 将 ProtoPlus 消息编码为字节数组
// msgObj: 要编码的消息对象
// ctx: 上下文集合（未使用）
// 使用 protoplus 进行序列化
// 返回编码后的字节数组
func (self *protoplus) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {

	return proto.Marshal(msgObj)

}

// Decode 将字节数组解码为 ProtoPlus 消息
// data: 要解码的数据，必须是 []byte
// msgObj: 要解码到的消息对象（指针）
// 使用 protoplus 进行反序列化
func (self *protoplus) Decode(data interface{}, msgObj interface{}) error {

	return proto.Unmarshal(data.([]byte), msgObj)
}

// init 包初始化函数
// 自动注册 ProtoPlus 编码器
func init() {

	codec.RegisterCodec(new(protoplus))
}
