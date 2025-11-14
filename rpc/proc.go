package rpc

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/msglog"
)

// RemoteCallMsg 远程调用消息接口
// 用于标识 RPC 相关的消息（请求和响应）
type RemoteCallMsg interface {
	// GetMsgID 获取消息 ID
	GetMsgID() uint16
	// GetMsgData 获取消息数据
	GetMsgData() []byte
	// GetCallID 获取调用 ID
	GetCallID() int64
}

// ResolveInboundEvent 处理入站 RPC 事件
// inputEvent: 输入事件
// 如果是 RPC 消息，进行解码和处理：
//   - RemoteCallREQ: 服务端收到客户端的请求，转换为 RecvMsgEvent
//   - RemoteCallACK: 客户端收到服务器的回应，触发请求的回调
// 返回处理后的输出事件、是否已处理、错误
func ResolveInboundEvent(inputEvent cellnet.Event) (ouputEvent cellnet.Event, handled bool, err error) {

	// 如果已经是 RecvMsgEvent，直接返回
	if _, ok := inputEvent.(*RecvMsgEvent); ok {
		return inputEvent, false, nil
	}

	// 检查是否是 RPC 消息
	rpcMsg, ok := inputEvent.Message().(RemoteCallMsg)
	if !ok {
		return inputEvent, false, nil
	}

	// 解码用户消息
	userMsg, _, err := codec.DecodeMessage(int(rpcMsg.GetMsgID()), rpcMsg.GetMsgData())

	if err != nil {
		return inputEvent, false, err
	}

	// 记录 RPC 接收日志
	if msglog.IsMsgLogValid(int(rpcMsg.GetMsgID())) {
		peerInfo := inputEvent.Session().Peer().(cellnet.PeerProperty)

		log.GetLog().Debugf("#rpc.recv(%s)@%d len: %d %s | %s",
			peerInfo.Name(),
			inputEvent.Session().ID(),
			cellnet.MessageSize(userMsg),
			cellnet.MessageToName(userMsg),
			cellnet.MessageToString(userMsg))
	}

	// 根据消息类型处理
	switch inputEvent.Message().(type) {
	case *RemoteCallREQ: // 服务端收到客户端的请求
		// 转换为 RecvMsgEvent，包含调用 ID
		return &RecvMsgEvent{
			inputEvent.Session(),
			userMsg,
			rpcMsg.GetCallID(),
		}, true, nil

	case *RemoteCallACK: // 客户端收到服务器的回应
		// 查找对应的请求并触发回调
		request := getRequest(rpcMsg.GetCallID())
		if request != nil {
			request.RecvFeedback(userMsg)
		}

		return inputEvent, true, nil
	}

	return inputEvent, false, nil
}

// ResolveOutboundEvent 处理出站 RPC 事件
// inputEvent: 输入事件
// 如果是 RPC 消息，记录发送日志并标记为已处理
// 返回是否已处理、错误
func ResolveOutboundEvent(inputEvent cellnet.Event) (handled bool, err error) {
	// 检查是否是 RPC 消息
	rpcMsg, ok := inputEvent.Message().(RemoteCallMsg)
	if !ok {
		return false, nil
	}

	// 解码用户消息（用于日志）
	userMsg, _, err := codec.DecodeMessage(int(rpcMsg.GetMsgID()), rpcMsg.GetMsgData())

	if err != nil {
		return false, err
	}

	// 记录 RPC 发送日志
	if msglog.IsMsgLogValid(int(rpcMsg.GetMsgID())) {
		peerInfo := inputEvent.Session().Peer().(cellnet.PeerProperty)

		log.GetLog().Debugf("#rpc.send(%s)@%d len: %d %s | %s",
			peerInfo.Name(),
			inputEvent.Session().ID(),
			cellnet.MessageSize(userMsg),
			cellnet.MessageToName(userMsg),
			cellnet.MessageToString(userMsg))
	}

	// 避免后续环节处理（RPC 消息已经封装，不需要再次处理）

	return true, nil
}
