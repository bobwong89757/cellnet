package relay

import (
	"errors"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
)

var (
	// ErrInvalidPeerSession 表示无效的 Peer 或 Session 错误
	ErrInvalidPeerSession = errors.New("Require valid cellnet.Session or cellnet.TCPConnector")
)

// Relay 转发消息到指定的 Session
// sesDetector: Session 或 TCPConnector，用于确定目标 Session
// dataList: 要转发的数据列表，支持以下类型：
//   - 消息对象（会被编码）
//   - []byte（作为原始数据）
//   - int64（作为透传数据）
//   - []int64（作为透传数据）
//   - string（作为透传数据）
// 返回转发错误，如果成功则返回 nil
// 注意：只能转发一个消息对象，多个消息对象会触发 panic
func Relay(sesDetector interface{}, dataList ...interface{}) error {
	// 获取 Session
	ses, err := getSession(sesDetector)
	if err != nil {
		log.GetLog().Errorf("relay.Relay:", err)
		return err
	}

	var ack RelayACK

	// 处理所有数据
	for _, payload := range dataList {
		switch value := payload.(type) {
		case int64:
			// 透传 int64 数据
			ack.Int64 = value
		case []int64:
			// 透传 int64 切片数据
			ack.Int64Slice = value

		case string:
			// 透传字符串数据
			ack.Str = value
		case []byte:
			// 作为原始数据 payload
			ack.Bytes = value
		default:
			// 消息对象，需要编码
			if ack.MsgID == 0 {
				var meta *cellnet.MessageMeta
				// 编码消息
				ack.Msg, meta, err = codec.EncodeMessage(payload, nil)

				if err != nil {
					return err
				}

				ack.MsgID = uint32(meta.ID)
			} else {
				// 不支持多个消息对象
				panic("Multi message relay not support")
			}
		}
	}
	// 发送转发消息
	ses.Send(&ack)

	return nil
}

// getSession 从检测器获取 Session
// sesDetector: Session 或 TCPConnector
// 返回对应的 Session 和错误信息
// 如果类型不支持，返回错误
func getSession(sesDetector interface{}) (cellnet.Session, error) {
	switch unknown := sesDetector.(type) {
	case cellnet.Session:
		// 直接是 Session
		return unknown, nil
	case cellnet.TCPConnector:
		// TCPConnector，获取其 Session
		return unknown.Session(), nil
	default:
		// 不支持的类型
		return nil, ErrInvalidPeerSession
	}
}
