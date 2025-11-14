package rpc

import (
	"github.com/bobwong89757/cellnet"
	"time"
)

// Call 执行异步 RPC 请求
// sesOrPeer: Session 或 Peer，用于发送请求
// reqMsg: 请求消息对象
// timeout: 超时时间，如果在此时间内未收到响应，会调用回调并传入超时错误
// userCallback: 响应回调函数，参数为响应消息或错误
//   回调函数会在 Session 对应 Peer 的事件队列中执行，保证线程安全
// 此方法不会阻塞，立即返回
func Call(sesOrPeer interface{}, reqMsg interface{}, timeout time.Duration, userCallback func(raw interface{})) {
	// 获取 Session
	ses, err := getPeerSession(sesOrPeer)

	if err != nil {
		// 获取 Session 失败，在队列中调用回调并传入错误
		cellnet.SessionQueuedCall(ses, func() {
			userCallback(err)
		})
		return
	}

	// 创建 RPC 请求，响应时在队列中调用回调
	req := createRequest(func(raw interface{}) {
		cellnet.SessionQueuedCall(ses, func() {
			userCallback(raw)
		})
	})

	// 发送 RPC 请求
	req.Send(ses, reqMsg)

	// 设置超时定时器
	time.AfterFunc(timeout, func() {
		// 取出请求，如果存在，说明请求还未收到响应，调用超时回调
		if getRequest(req.id) != nil {
			cellnet.SessionQueuedCall(ses, func() {
				userCallback(ErrTimeout)
			})
		}
	})
}
