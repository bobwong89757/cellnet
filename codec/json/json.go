package json

import (
	"encoding/json"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
)

// jsonCodec 实现 JSON 编码器
// 使用 Go 标准库的 encoding/json 进行 JSON 序列化和反序列化
// 适合与第三方服务通信，具有良好的可读性和兼容性
type jsonCodec struct {
}

// Name 返回编码器的名称
// 返回 "json"
func (self *jsonCodec) Name() string {
	return "json"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/json"
func (self *jsonCodec) MimeType() string {
	return "application/json"
}

// Encode 将消息对象编码为 JSON 字节数组
// msgObj: 要编码的消息对象，通常是结构体指针
// ctx: 上下文信息（此编码器不使用）
// 返回编码后的 JSON 字节数组和错误信息
// 使用标准库的 json.Marshal 进行编码
func (self *jsonCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {
	return json.Marshal(msgObj)
}

// Decode 将 JSON 字节数组解码为消息对象
// data: 要解码的 JSON 字节数组
// msgObj: 目标消息对象，通常是指针类型，解码结果会写入此对象
// 返回解码错误，如果成功则返回 nil
// 使用标准库的 json.Unmarshal 进行解码
func (self *jsonCodec) Decode(data interface{}, msgObj interface{}) error {
	return json.Unmarshal(data.([]byte), msgObj)
}

// init 在包加载时自动注册 JSON 编码器
// 将 jsonCodec 注册到全局编码器列表中
func init() {
	// 注册编码器
	codec.RegisterCodec(new(jsonCodec))
}
