package codec

import (
	"github.com/bobwong89757/cellnet"
)

// EncodeMessage 编码消息对象为字节数组
// msg: 要编码的消息对象，通常是指针类型
// ctx: 上下文信息，用于传递编码相关的配置或资源
//      在使用带内存池的 codec 时，可以传入 session 或 peer 的 ContextSet 来保存内存池上下文
//      默认可以传 nil
// 返回编码后的字节数组、消息元信息和错误信息
// 如果消息未注册，返回错误
func EncodeMessage(msg interface{}, ctx cellnet.ContextSet) (data []byte, meta *cellnet.MessageMeta, err error) {
	// 根据消息对象获取消息元信息
	meta = cellnet.MessageMetaByMsg(msg)
	if meta == nil {
		return nil, nil, cellnet.NewErrorContext("msg not exists", msg)
	}

	// 使用消息对应的 Codec 将消息编码为字节数组
	var raw interface{}
	raw, err = meta.Codec.Encode(msg, ctx)

	if err != nil {
		return
	}

	// 类型断言为字节数组
	data = raw.([]byte)

	return
}

// DecodeMessage 根据消息ID解码字节数组为消息对象
// msgid: 消息的唯一标识符，用于查找消息类型
// data: 要解码的字节数组
// 返回解码后的消息对象、消息元信息和错误信息
// 如果消息ID未注册，返回错误
func DecodeMessage(msgid int, data []byte) (interface{}, *cellnet.MessageMeta, error) {
	// 根据消息ID获取消息元信息
	meta := cellnet.MessageMetaByID(msgid)

	// 检查消息是否已注册
	if meta == nil {
		return nil, nil, cellnet.NewErrorContext("msg not exists", msgid)
	}

	// 创建消息类型的实例
	msg := meta.NewType()

	// 使用消息对应的 Codec 从字节数组解码为消息对象
	err := meta.Codec.Decode(data, msg)

	if err != nil {
		return nil, meta, err
	}

	return msg, meta, nil
}

// DecodeMessageByType 根据消息类型解码字节数组
// data: 要解码的字节数组
// msg: 目标消息对象，通常是指针类型，解码结果会写入此对象
// 返回消息元信息和错误信息
// 如果消息类型未注册，返回错误
// 与 DecodeMessage 不同，此方法需要预先知道消息类型
func DecodeMessageByType(data []byte, msg interface{}) (*cellnet.MessageMeta, error) {
	// 根据消息对象获取消息元信息
	meta := cellnet.MessageMetaByMsg(msg)
	// 检查消息是否已注册
	if meta == nil {
		return nil, cellnet.NewErrorContext("msg not exists", nil)
	}

	// 使用消息对应的 Codec 解码字节数组到消息对象
	err := meta.Codec.Decode(data, msg)
	if err != nil {
		return meta, err
	}

	return meta, nil
}

// CodecRecycler 定义编码器资源回收接口
// 用于回收 Codec.Encode 内分配的资源，例如内存池对象
// 实现了此接口的 Codec 可以在编码后回收临时资源，提高性能
type CodecRecycler interface {
	// Free 释放编码过程中分配的资源
	// data: 编码后的数据
	// ctx: 上下文信息，用于传递资源回收相关的信息
	Free(data interface{}, ctx cellnet.ContextSet)
}

// FreeCodecResource 释放 Codec 编码时分配的资源
// codec: 编码器，如果为 nil 则不处理
// data: 编码后的数据
// ctx: 上下文信息
// 如果 Codec 实现了 CodecRecycler 接口，会调用其 Free 方法回收资源
// 用于在使用内存池等优化技术时，及时释放临时资源
func FreeCodecResource(codec cellnet.Codec, data interface{}, ctx cellnet.ContextSet) {
	if codec == nil {
		return
	}

	// 检查 Codec 是否实现了资源回收接口
	if recycler, ok := codec.(CodecRecycler); ok {
		recycler.Free(data, ctx)
	}
}
