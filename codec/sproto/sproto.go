package sprotocodec

import (
	"fmt"
	"path"
	"reflect"

	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/util"
	"github.com/bobwong89757/gosproto"
)

// sprotoCodec 实现 Sproto 编码器
// Sproto 是一个轻量级二进制协议，常用于游戏开发
// 使用 gosproto 库进行 Sproto 协议的编码和解码
type sprotoCodec struct {
}

// Name 返回编码器的名称
// 返回 "sproto"
func (self *sprotoCodec) Name() string {
	return "sproto"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/sproto"
func (self *sprotoCodec) MimeType() string {
	return "application/sproto"
}

// Encode 将消息对象编码为 Sproto 字节数组
// msgObj: 要编码的消息对象
// ctx: 上下文信息（此编码器不使用）
// 返回编码后的 Sproto 字节数组和错误信息
// 编码过程：先使用 sproto.Encode 编码，然后使用 sproto.Pack 打包
func (self *sprotoCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {
	// 第一步：编码消息对象
	result, err := sproto.Encode(msgObj)
	if err != nil {
		return nil, err
	}

	// 第二步：打包编码结果
	return sproto.Pack(result), nil
}

// Decode 将 Sproto 字节数组解码为消息对象
// data: 要解码的 Sproto 字节数组
// msgObj: 目标消息对象，解码结果会写入此对象
// 返回解码错误，如果成功则返回 nil
// 解码过程：先使用 sproto.Unpack 解包，然后使用 sproto.Decode 解码
// 注意：sproto 要求必须有头，但空包也是可以的
func (self *sprotoCodec) Decode(data interface{}, msgObj interface{}) error {
	tmp := data.([]byte)
	// sproto 要求必须有头，但空包也是可以的
	if len(tmp) == 0 {
		return nil
	}

	// 第一步：解包数据
	raw, err := sproto.Unpack(tmp)
	if err != nil {
		return err
	}

	// 第二步：解码数据到消息对象
	_, err2 := sproto.Decode(raw, msgObj)

	return err2
}

// AutoRegisterMessageMeta 自动注册多个消息类型的元信息
// msgTypes: 要注册的消息类型列表
// 为每个消息类型创建 MessageMeta 并注册到全局消息注册表
// 消息 ID 通过消息名称的字符串哈希生成，确保唯一性
// 所有消息都使用 sproto 编码器
func AutoRegisterMessageMeta(msgTypes []reflect.Type) {
	for _, tp := range msgTypes {
		// 生成消息的完整名称（包名.类型名）
		msgName := fmt.Sprintf("%s.%s", path.Base(tp.PkgPath()), tp.Name())

		// 注册消息元信息
		// 使用字符串哈希生成消息 ID，确保唯一性
		cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
			Codec: codec.MustGetCodec("sproto"),
			Type:  tp,
			ID:    int(util.StringHash(msgName)),
		})
	}
}

// init 在包加载时自动注册 Sproto 编码器
// 将 sprotoCodec 注册到全局编码器列表中
func init() {
	codec.RegisterCodec(new(sprotoCodec))
}