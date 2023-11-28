package udp

import (
	"encoding/binary"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer/udp"
)

// sendPacket
//
//	@Description: 发送消息
//	@param writer
//	@param ctx
//	@param msg
//	@return error
func sendPacket(writer udp.DataWriter, ctx cellnet.ContextSet, msg interface{}) error {

	// 将用户数据转换为字节数组和消息ID
	msgData, meta, err := codec.EncodeMessage(msg, ctx)

	if err != nil {
		log.GetLog().Errorf("send message encode error: %s", err)
		return err
	}

	pktData := make([]byte, HeaderSize+len(msgData))

	// 写入消息长度做验证
	binary.LittleEndian.PutUint16(pktData, uint16(HeaderSize+len(msgData)))

	// Type
	binary.LittleEndian.PutUint16(pktData[2:], uint16(meta.ID))

	// Value
	copy(pktData[HeaderSize:], msgData)

	writer.WriteData(pktData)

	codec.FreeCodecResource(meta.Codec, msgData, ctx)

	return nil
}
