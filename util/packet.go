package util

import (
	"encoding/binary"
	"errors"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"io"
)

var (
	// ErrMaxPacket 表示数据包超过最大大小的错误
	ErrMaxPacket = errors.New("packet over size")

	// ErrMinPacket 表示数据包大小不足的错误
	ErrMinPacket = errors.New("packet short size")

	// ErrShortMsgID 表示消息 ID 字段不足的错误
	ErrShortMsgID = errors.New("short msgid")
)

const (
	// bodySize 包体大小字段的字节数
	// 使用 uint16，占 2 字节
	bodySize = 2

	// msgIDSize 消息 ID 字段的字节数
	// 使用 uint16，占 2 字节
	msgIDSize = 2
)

// RecvLTVPacket 接收 Length-Type-Value 格式的封包
// reader: 数据读取器
// maxPacketSize: 最大数据包大小，如果为 0 则不限制
// 返回解码后的消息对象和错误信息
//
// 数据包格式：
//   - Length (2 bytes): 包体大小（Type + Value）
//   - Type (2 bytes): 消息 ID
//   - Value (N bytes): 消息数据
func RecvLTVPacket(reader io.Reader, maxPacketSize int) (msg interface{}, err error) {
	// Size 为 uint16，占 2 字节
	var sizeBuffer = make([]byte, bodySize)

	// 持续读取 Size 直到读到为止
	_, err = io.ReadFull(reader, sizeBuffer)

	// 发生错误时返回
	if err != nil {
		return
	}

	// 检查缓冲区大小
	if len(sizeBuffer) < bodySize {
		return nil, ErrMinPacket
	}

	// 用小端格式读取 Size
	size := binary.LittleEndian.Uint16(sizeBuffer)

	// 检查数据包大小是否超过限制
	if maxPacketSize > 0 && int(size) >= maxPacketSize {
		return nil, ErrMaxPacket
	}

	// 分配包体大小
	body := make([]byte, size)

	// 读取包体数据
	_, err = io.ReadFull(reader, body)

	// 发生错误时返回
	if err != nil {
		return
	}

	// 检查包体大小是否足够包含消息 ID
	if len(body) < msgIDSize {
		return nil, ErrShortMsgID
	}

	// 读取消息 ID
	msgid := binary.LittleEndian.Uint16(body)

	// 提取消息数据
	msgData := body[msgIDSize:]

	// 将字节数组和消息 ID 解码为消息对象
	msg, _, err = codec.DecodeMessage(int(msgid), msgData)
	if err != nil {
		// TODO 接收错误时，返回消息
		return nil, err
	}

	return
}

// SendLTVPacket 发送 Length-Type-Value 格式的封包
// writer: 数据写入器
// ctx: 上下文信息，用于传递编码相关的配置或资源
// data: 要发送的数据，可以是消息对象或 *cellnet.RawPacket
// 返回写入错误，如果成功则返回 nil
//
// 数据包格式：
//   - Length (2 bytes): 包体大小（Type + Value）
//   - Type (2 bytes): 消息 ID
//   - Value (N bytes): 消息数据
//
// 如果 data 是 *cellnet.RawPacket，则直接使用其数据；否则会先编码消息
func SendLTVPacket(writer io.Writer, ctx cellnet.ContextSet, data interface{}) error {
	var (
		msgData []byte
		msgID   int
		meta    *cellnet.MessageMeta
	)

	switch m := data.(type) {
	case *cellnet.RawPacket:
		// 发送裸包，直接使用原始数据
		msgData = m.MsgData
		msgID = m.MsgID
	default:
		// 发送普通编码包，需要先编码消息
		var err error

		// 将用户数据转换为字节数组和消息 ID
		msgData, meta, err = codec.EncodeMessage(data, ctx)

		if err != nil {
			return err
		}

		msgID = meta.ID
	}

	// 分配数据包缓冲区
	pkt := make([]byte, bodySize+msgIDSize+len(msgData))

	// 写入 Length（包体大小 = Type + Value）
	binary.LittleEndian.PutUint16(pkt, uint16(msgIDSize+len(msgData)))

	// 写入 Type（消息 ID）
	binary.LittleEndian.PutUint16(pkt[bodySize:], uint16(msgID))

	// 写入 Value（消息数据）
	copy(pkt[bodySize+msgIDSize:], msgData)

	// 将数据写入 Socket
	err := WriteFull(writer, pkt)

	// Codec 中使用内存池时的释放位置
	if meta != nil {
		codec.FreeCodecResource(meta.Codec, msgData, ctx)
	}

	return err
}
