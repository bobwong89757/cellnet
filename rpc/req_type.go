package rpc

import (
	"github.com/bobwong89757/cellnet"
	"reflect"
	"sync"
	"time"
)

// CallType 执行异步 RPC 请求，按消息类型匹配响应
// sesOrPeer: Session 或 Peer，用于发送请求
// reqMsg: 请求消息对象
// timeout: 超时时间
// userCallback: 响应回调函数，格式为 func(ackMsg *AckMsgType, err error)
//   回调函数的第一个参数类型用于匹配响应消息类型
// 此方法不会阻塞，立即返回
func CallType(sesOrPeer interface{}, reqMsg interface{}, timeout time.Duration, userCallback interface{}) {
	callType(sesOrPeer, false, reqMsg, timeout, userCallback)
}

// CallSyncType 执行同步 RPC 请求，按消息类型匹配响应
// sesOrPeer: Session 或 Peer，用于发送请求
// reqMsg: 请求消息对象
// timeout: 超时时间
// userCallback: 响应回调函数，格式为 func(ackMsg *AckMsgType, err error)
//   回调函数的第一个参数类型用于匹配响应消息类型
// 此方法会阻塞当前 goroutine，直到收到响应或超时
func CallSyncType(sesOrPeer interface{}, reqMsg interface{}, timeout time.Duration, userCallback interface{}) {
	callType(sesOrPeer, true, reqMsg, timeout, userCallback)
}

// callType 执行 RPC 请求，按消息类型匹配响应
// sesOrPeer: Session 或 Peer，用于发送请求
// sync: 是否为同步请求
// reqMsg: 请求消息对象
// timeout: 超时时间
// userCallback: 响应回调函数，格式为 func(ackMsg *AckMsgType, err error)
//   回调函数的第一个参数类型用于匹配响应消息类型
// 一般用于客户端请求，通过回调函数的参数类型自动匹配响应消息
func callType(sesOrPeer interface{}, sync bool, reqMsg interface{}, timeout time.Duration, userCallback interface{}) {
	// 获取回调函数的类型
	funcType := reflect.TypeOf(userCallback)
	if funcType.Kind() != reflect.Func {
		panic("type rpc callback require 'func'")
	}

	// 检查参数数量
	if funcType.NumIn() != 2 {
		panic("callback func param format like 'func(ack *YouMsgACK)'")
	}

	// 获取第一个参数类型（响应消息类型）
	ackType := funcType.In(0)
	if ackType.Kind() != reflect.Ptr {
		panic("callback func param format like 'func(ack *YouMsgACK)'")
	}

	// 获取非指针类型
	ackType = ackType.Elem()

	// 创建调用函数
	callFunc := func(rawACK interface{}, err error) {
		vCall := reflect.ValueOf(userCallback)

		// 如果响应为 nil，创建零值
		if rawACK == nil {
			rawACK = reflect.New(ackType).Interface()
		}

		// 处理错误参数
		var errV reflect.Value
		if err == nil {
			errV = nilError
		} else {
			errV = reflect.ValueOf(err)
		}

		// 调用回调函数
		vCall.Call([]reflect.Value{reflect.ValueOf(rawACK), errV})
	}

	// 获取 Session
	ses, err := getPeerSession(sesOrPeer)

	if err != nil {
		callFunc(nil, err)
		return
	}

	// 创建类型请求
	createTypeRequest(sync, ackType, timeout, func() {
		ses.Send(reqMsg)
	}, callFunc)
}

var (
	// nilError 表示 nil 错误的反射值
	// 用于在回调函数中传递 nil 错误
	nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

	// callByType 存储按类型匹配的 RPC 请求
	// 键为响应消息类型（reflect.Type），值为回调函数或通道
	// 使用 sync.Map 保证并发安全
	callByType sync.Map
)

// createTypeRequest 创建按类型匹配的 RPC 请求
// sync: 是否为同步请求
// ackType: 响应消息类型
// timeout: 超时时间
// onSend: 发送请求的函数
// onRecv: 接收响应的回调函数
func createTypeRequest(sync bool, ackType reflect.Type, timeout time.Duration, onSend func(), onRecv func(rawACK interface{}, err error)) {
	if sync {
		// 同步请求：使用通道等待响应
		feedBack := make(chan interface{})
		callByType.Store(ackType, feedBack)

		defer callByType.Delete(ackType)

		// 发送请求
		onSend()

		// 等待响应或超时
		select {
		case ack := <-feedBack:
			onRecv(ack, nil)
		case <-time.After(timeout):
			onRecv(nil, ErrTimeout)
		}
	} else {
		// 异步请求：使用回调函数
		callByType.Store(ackType, func(rawACK interface{}, err error) {
			onRecv(rawACK, err)
			callByType.Delete(ackType)
		})

		// 发送请求
		onSend()

		// 注意：丢弃超时的类型，避免重复请求时，将第二次请求的消息删除
	}
}

// TypeRPCHooker 按类型匹配的 RPC 钩子
// 用于在消息处理流程中匹配响应消息类型并调用对应的回调函数
type TypeRPCHooker struct {
}

// OnInboundEvent 处理入站事件
// inputEvent: 输入的接收事件
// 返回处理后的输出事件
// 如果接收到的消息类型匹配某个等待的 RPC 请求，会调用对应的回调函数
func (TypeRPCHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	// 获取消息类型（非指针类型）
	incomingMsgType := reflect.TypeOf(inputEvent.Message()).Elem()

	// 查找匹配的 RPC 请求
	if rawFeedback, ok := callByType.Load(incomingMsgType); ok {
		switch feedBack := rawFeedback.(type) {
		case func(rawACK interface{}, err error):
			// 异步请求：调用回调函数
			feedBack(inputEvent.Message(), nil)
		case chan interface{}:
			// 同步请求：通过通道发送响应
			feedBack <- inputEvent.Message()
		}

		return inputEvent
	}

	return inputEvent
}

// OnOutboundEvent 处理出站事件
// inputEvent: 输入的发送事件
// 返回处理后的输出事件
// 此实现不处理出站事件，直接返回
func (TypeRPCHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	return inputEvent
}
