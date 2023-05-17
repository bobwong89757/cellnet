package relay

type BroadcasterFunc func(event *RecvMsgEvent)

var bcFunc BroadcasterFunc

// SetBroadcaster
//
//	@Description: 设置广播函数, 回调时，按对应Peer/Session所在的队列中调用
//	@param callback
func SetBroadcaster(callback BroadcasterFunc) {

	bcFunc = callback
}
