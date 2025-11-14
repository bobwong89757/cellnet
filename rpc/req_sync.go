package rpc

import (
	"time"
)

// CallSync 执行同步 RPC 请求
// ud: Peer 或 Session，用于发送请求
// reqMsg: 请求消息对象
// timeout: 超时时间，如果在此时间内未收到响应，返回超时错误
// 返回响应消息和错误信息
// 此方法会阻塞当前 goroutine，直到收到响应或超时
func CallSync(ud interface{}, reqMsg interface{}, timeout time.Duration) (interface{}, error) {
	// 获取 Session
	ses, err := getPeerSession(ud)

	if err != nil {
		return nil, err
	}

	// 创建响应通道
	ret := make(chan interface{})
	// 创建 RPC 请求，响应时通过通道返回
	req := createRequest(func(feedbackMsg interface{}) {
		ret <- feedbackMsg
	})

	// 发送 RPC 请求
	req.Send(ses, reqMsg)

	// 等待 RPC 回复或超时
	select {
	case v := <-ret:
		// 收到响应，返回响应消息
		return v, nil
	case <-time.After(timeout):
		// 超时，清理请求并返回超时错误
		getRequest(req.id)
		return nil, ErrTimeout
	}
}
