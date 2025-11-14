package proc

import (
	"github.com/bobwong89757/cellnet"
	"reflect"
	"sync"
)

// SyncReceiver 同步接收消息器，可选组件
// 可作为流程测试辅助工具，用于同步等待消息到达
// 所有接收到的消息都会通过通道传递，可以同步等待特定消息
type SyncReceiver struct {
	// evChan 事件通道
	// 用于传递接收到的消息事件
	evChan chan cellnet.Event

	// callback 事件回调函数
	// 当消息到达时，会将事件发送到 evChan
	callback func(ev cellnet.Event)
}

// EventCallback 将处理回调返回给 BindProcessorHandler 用于注册
// 返回事件回调函数，可以用于绑定到处理器
func (self *SyncReceiver) EventCallback() cellnet.EventCallback {
	return self.callback
}

// Recv 持续阻塞，直到某个消息到达后，使用回调返回消息
// callback: 消息到达时的回调函数
// 返回自身以便链式调用
// 此方法会阻塞当前 goroutine，直到有消息到达
func (self *SyncReceiver) Recv(callback cellnet.EventCallback) *SyncReceiver {
	// 从通道接收事件并调用回调
	callback(<-self.evChan)
	return self
}

// WaitMessage 持续阻塞，直到某个指定类型的消息到达后，返回消息
// msgName: 消息的完整名称，格式为 "包名.类型名"
// 返回接收到的消息对象
// 如果消息名称未注册，会触发 panic
// 此方法会阻塞当前 goroutine，直到匹配的消息到达
func (self *SyncReceiver) WaitMessage(msgName string) (msg interface{}) {
	var wg sync.WaitGroup

	// 根据消息名称获取消息元信息
	meta := cellnet.MessageMetaByFullName(msgName)
	if meta == nil {
		panic("unknown message name:" + msgName)
	}

	wg.Add(1)

	// 接收消息，直到匹配到指定类型
	self.Recv(func(ev cellnet.Event) {
		// 检查消息类型是否匹配
		inMeta := cellnet.MessageMetaByType(reflect.TypeOf(ev.Message()))
		if inMeta == meta {
			// 类型匹配，保存消息并通知等待
			msg = ev.Message()
			wg.Done()
		}
	})

	// 等待匹配的消息到达
	wg.Wait()
	return
}

// NewSyncReceiver 新建同步消息接收器
// p: 要绑定接收器的 Peer
// 返回初始化好的 SyncReceiver
// 需要将返回的 EventCallback 绑定到 Peer 的处理器才能接收消息
func NewSyncReceiver(p cellnet.Peer) *SyncReceiver {
	self := &SyncReceiver{
		evChan: make(chan cellnet.Event),
	}

	// 创建回调函数，将事件发送到通道
	self.callback = func(ev cellnet.Event) {
		self.evChan <- ev
	}

	return self
}
