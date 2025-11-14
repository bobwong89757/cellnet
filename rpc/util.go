package rpc

import (
	"errors"
	"github.com/bobwong89757/cellnet"
)

var (
	// ErrInvalidPeerSession 表示无效的 Peer 或 Session 错误
	// 要求提供 cellnet.RPCSessionGetter 或 cellnet.Session
	ErrInvalidPeerSession = errors.New("rpc: Invalid peer type, require cellnet.RPCSessionGetter or cellnet.Session")

	// ErrEmptySession 表示 Session 为空的错误
	ErrEmptySession = errors.New("rpc: Empty session")
)

// RPCSessionGetter 定义获取 RPC Session 的接口
// 某些 Peer 类型（如 TCPConnector）可以通过此接口获取用于 RPC 的 Session
type RPCSessionGetter interface {
	// RPCSession 获取用于 RPC 的 Session
	RPCSession() cellnet.Session
}

// getPeerSession 从 Peer 获取 RPC 使用的 Session
// ud: Session、RPCSessionGetter 或 TCPConnector
// 返回对应的 Session 和错误信息
// 如果类型不支持或 Session 为空，返回错误
func getPeerSession(ud interface{}) (ses cellnet.Session, err error) {
	if ud == nil {
		return nil, ErrInvalidPeerSession
	}

	// 根据类型获取 Session
	switch i := ud.(type) {
	case RPCSessionGetter:
		// 实现了 RPCSessionGetter 接口
		ses = i.RPCSession()
	case cellnet.Session:
		// 直接是 Session
		ses = i
	case cellnet.TCPConnector:
		// TCPConnector，获取其 Session
		ses = i.Session()
	default:
		// 不支持的类型
		err = ErrInvalidPeerSession
		return
	}

	// 检查 Session 是否为空
	if ses == nil {
		return nil, ErrEmptySession
	}

	return
}
