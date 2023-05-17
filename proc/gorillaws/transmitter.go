package gorillaws

import (
	"encoding/binary"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/util"
	"github.com/gorilla/websocket"
)

const (
	MsgIDSize = 2 // uint16
)

// WSMessageTransmitter
// @Description: WS消息传输器
type WSMessageTransmitter struct {
}

// OnRecvMessage
//
//	@Description: 接收消息
//	@receiver WSMessageTransmitter
//	@param ses
//	@return msg
//	@return err
func (WSMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {

	conn, ok := ses.Raw().(*websocket.Conn)

	// 转换错误，或者连接已经关闭时退出
	if !ok || conn == nil {
		return nil, nil
	}

	var messageType int
	var raw []byte
	messageType, raw, err = conn.ReadMessage()

	if err != nil {
		return
	}

	if len(raw) < MsgIDSize {
		return nil, util.ErrMinPacket
	}

	switch messageType {
	case websocket.BinaryMessage:
		msgID := binary.LittleEndian.Uint16(raw)
		msgData := raw[MsgIDSize:]

		msg, _, err = codec.DecodeMessage(int(msgID), msgData)
	}

	return
}

// OnSendMessage
//
//	@Description: 发送消息
//	@receiver WSMessageTransmitter
//	@param ses
//	@param msg
//	@return error
func (WSMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {

	conn, ok := ses.Raw().(*websocket.Conn)

	// 转换错误，或者连接已经关闭时退出
	if !ok || conn == nil {
		return nil
	}

	var (
		msgData []byte
		msgID   int
	)

	switch m := msg.(type) {
	case *cellnet.RawPacket: // 发裸包
		msgData = m.MsgData
		msgID = m.MsgID
	default: // 发普通编码包
		var err error
		var meta *cellnet.MessageMeta

		// 将用户数据转换为字节数组和消息ID
		msgData, meta, err = codec.EncodeMessage(msg, nil)

		if err != nil {
			return err
		}

		msgID = meta.ID
	}

	pkt := make([]byte, MsgIDSize+len(msgData))
	binary.LittleEndian.PutUint16(pkt, uint16(msgID))
	copy(pkt[MsgIDSize:], msgData)

	conn.WriteMessage(websocket.BinaryMessage, pkt)

	return nil
}
