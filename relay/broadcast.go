package relay

// BroadcasterFunc 定义广播函数的类型
// 当接收到 relay 消息时，会调用此函数进行广播
// event: 接收到的 relay 消息事件
type BroadcasterFunc func(event *RecvMsgEvent)

// bcFunc 全局的广播函数
// 通过 SetBroadcaster 设置
var bcFunc BroadcasterFunc

// SetBroadcaster 设置广播函数
// callback: 广播回调函数
// 回调时，会在对应 Peer/Session 所在的队列中调用，保证线程安全
// 用于处理接收到的 relay 消息，可以实现消息广播、转发等功能
func SetBroadcaster(callback BroadcasterFunc) {
	bcFunc = callback
}
