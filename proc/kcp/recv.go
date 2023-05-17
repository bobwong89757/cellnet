package kcp

import (
	"encoding/binary"
	"github.com/bobwong89757/cellnet/codec"
)

const (
	MTU       = 1472 // 最大传输单元
	BodySize  = 2    // 包体大小字段
	MsgIDSize = 2    // 消息ID字段
)

// RecvPacket
//
//	@Description: 解包, 2字节消息长度+2字节消息id+消息内容
//	@param pktData
//	@return msg
//	@return err
func RecvPacket(pktData []byte) (msg interface{}, err error) {

	// 用小端格式读取Size
	datasize := binary.LittleEndian.Uint16(pktData)

	//小于包头，使用nc指令测试时，为1
	if datasize < BodySize {
		return nil, nil
	}

	// 出错，等待下次数据
	if datasize > MTU {
		return nil, nil
	}

	// 读取消息ID
	msgid := binary.LittleEndian.Uint16(pktData[BodySize:])

	msgData := pktData[BodySize+MsgIDSize:]

	// 将字节数组和消息ID用户解出消息
	msg, _, err = codec.DecodeMessage(int(msgid), msgData)
	if err != nil {
		// TODO 接收错误时，返回消息
		return nil, err
	}

	return
}
