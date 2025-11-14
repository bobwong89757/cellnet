package udp

import (
	"encoding/binary"
	"github.com/bobwong89757/cellnet/codec"
)

const (
	// MTU 最大传输单元
	// UDP 数据包的最大大小（以太网 MTU 1500 - IP 头 20 - UDP 头 8 = 1472）
	MTU = 1472

	// packetLen 包体大小字段长度（字节）
	packetLen = 2

	// MsgIDLen 消息 ID 字段长度（字节）
	MsgIDLen = 2

	// HeaderSize UDP 包头总大小
	// 包含：包体大小（2字节）+ 消息ID（2字节）
	HeaderSize = MsgIDLen + MsgIDLen
)

// RecvPacket 接收 UDP 数据包并解码为消息
// pktData: 接收到的 UDP 数据包
// UDP 数据包格式：[包体大小(2字节)][消息ID(2字节)][消息数据]
// 返回解码后的消息和错误
func RecvPacket(pktData []byte) (msg interface{}, err error) {

	// 小于包头，使用 nc 指令测试时，可能为 1
	if len(pktData) < packetLen {
		return nil, nil
	}

	// 用小端格式读取包体大小
	datasize := binary.LittleEndian.Uint16(pktData)

	// 验证数据包大小：包体大小必须等于实际数据长度，且不能超过 MTU
	// 出错时，等待下次数据
	if int(datasize) != len(pktData) || datasize > MTU {
		return nil, nil
	}

	// 读取消息 ID（从第 2 字节开始）
	msgid := binary.LittleEndian.Uint16(pktData[packetLen:])

	// 提取消息数据（跳过包头）
	msgData := pktData[HeaderSize:]

	// 将字节数组和消息 ID 解码为消息
	msg, _, err = codec.DecodeMessage(int(msgid), msgData)
	if err != nil {
		// TODO 接收错误时，返回消息
		return nil, err
	}

	return
}
