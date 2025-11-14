package rpc

import (
	"errors"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
	"sync"
	"sync/atomic"
)

var (
	// rpcIDSeq RPC 请求 ID 序列号
	// 使用原子操作保证并发安全，每次创建请求时递增
	rpcIDSeq int64

	// requestByCallID 存储所有待响应的 RPC 请求
	// 键为请求 ID（int64），值为 request 实例
	// 使用 sync.Map 保证并发安全
	requestByCallID sync.Map
)

// request 表示一个 RPC 请求
// 用于跟踪 RPC 请求的状态和回调
type request struct {
	// id 请求的唯一标识符
	id int64

	// onRecv 接收到响应时的回调函数
	// 参数为响应消息
	onRecv func(interface{})
}

// ErrTimeout 表示 RPC 请求超时的错误
var ErrTimeout = errors.New("RPC time out")

// RecvFeedback 接收 RPC 响应反馈
// msg: 响应消息
// 调用注册的回调函数处理响应
// 注意：异步和同步执行复杂，队列处理在具体的逻辑中手动处理
func (self *request) RecvFeedback(msg interface{}) {
	// 异步和同步执行复杂，队列处理在具体的逻辑中手动处理
	self.onRecv(msg)
}

// Send 发送 RPC 请求
// ses: 要发送请求的 Session
// msg: 请求消息对象
// 将请求消息编码后发送到远程端
func (self *request) Send(ses cellnet.Session, msg interface{}) {
	// 编码请求消息
	data, meta, err := codec.EncodeMessage(msg, nil)

	if err != nil {
		log.GetLog().Errorf("rpc request message encode error: %s", err)
		return
	}

	// 发送 RPC 请求消息
	ses.Send(&RemoteCallREQ{
		MsgID:  uint32(meta.ID),
		Data:   data,
		CallID: self.id,
	})

	// 注意：如果需要释放 Codec 资源，可以在这里调用
	// codec.FreeCodecResource(meta.Codec, data, ctx)
}

// createRequest 创建一个新的 RPC 请求
// onRecv: 接收到响应时的回调函数
// 返回创建的 request 实例
// 请求会被分配一个唯一的 ID 并存储到全局映射表中
func createRequest(onRecv func(interface{})) *request {
	self := &request{
		onRecv: onRecv,
	}

	// 生成唯一的请求 ID
	self.id = atomic.AddInt64(&rpcIDSeq, 1)

	// 存储到全局映射表
	requestByCallID.Store(self.id, self)

	return self
}

// getRequest 根据请求 ID 获取并移除 RPC 请求
// callid: 请求的唯一标识符
// 返回对应的 request 实例，如果不存在返回 nil
// 获取后会将请求从映射表中删除
func getRequest(callid int64) *request {
	if v, ok := requestByCallID.Load(callid); ok {
		// 从映射表中删除
		requestByCallID.Delete(callid)
		return v.(*request)
	}

	return nil
}
