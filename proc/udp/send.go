package udp

import (
	"encoding/binary"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer/udp"
)

// sendPacket 发送 UDP 数据包
// writer: 数据写入器
// ctx: 上下文集合，用于内存池等资源管理
// msg: 要发送的消息
// UDP 数据包格式：[包体大小(2字节)][消息ID(2字节)][消息数据]
// 返回发送错误
func sendPacket(writer udp.DataWriter, ctx cellnet.ContextSet, msg interface{}) error {

	// 将用户数据转换为字节数组和消息 ID
	msgData, meta, err := codec.EncodeMessage(msg, ctx)

	if err != nil {
		log.GetLog().Errorf("send message encode error: %s", err)
		return err
	}

	// 创建数据包缓冲区（包头 + 消息数据）
	pktData := make([]byte, HeaderSize+len(msgData))

	// 写入消息长度做验证（包体大小 = 包头 + 消息数据）
	binary.LittleEndian.PutUint16(pktData, uint16(HeaderSize+len(msgData)))

	// 写入消息 ID（Type）
	binary.LittleEndian.PutUint16(pktData[2:], uint16(meta.ID))

	// 写入消息数据（Value）
	copy(pktData[HeaderSize:], msgData)

	// 发送数据包
	writer.WriteData(pktData)

	// 释放编码器资源（内存池等）
	codec.FreeCodecResource(meta.Codec, msgData, ctx)

	return nil
}
