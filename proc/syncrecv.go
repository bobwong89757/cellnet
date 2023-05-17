package proc

import (
	"github.com/bobwong89757/cellnet"
	"reflect"
	"sync"
)

// SyncReceiver
// @Description: 同步接收消息器, 可选件，可作为流程测试辅助工具
type SyncReceiver struct {
	evChan chan cellnet.Event

	callback func(ev cellnet.Event)
}

// EventCallback
//
//	@Description: 将处理回调返回给BindProcessorHandler用于注册
//	@receiver self
//	@return cellnet.EventCallback
func (self *SyncReceiver) EventCallback() cellnet.EventCallback {

	return self.callback
}

// Recv
//
//	@Description: 持续阻塞，直到某个消息到达后，使用回调返回消息
//	@receiver self
//	@param callback
//	@return *SyncReceiver
func (self *SyncReceiver) Recv(callback cellnet.EventCallback) *SyncReceiver {
	callback(<-self.evChan)
	return self
}

// WaitMessage
//
//	@Description: 持续阻塞，直到某个消息到达后，返回消息
//	@receiver self
//	@param msgName
//	@return msg
func (self *SyncReceiver) WaitMessage(msgName string) (msg interface{}) {

	var wg sync.WaitGroup

	meta := cellnet.MessageMetaByFullName(msgName)
	if meta == nil {
		panic("unknown message name:" + msgName)
	}

	wg.Add(1)

	self.Recv(func(ev cellnet.Event) {

		inMeta := cellnet.MessageMetaByType(reflect.TypeOf(ev.Message()))
		if inMeta == meta {
			msg = ev.Message()
			wg.Done()
		}

	})

	wg.Wait()
	return
}

// NewSyncReceiver
//
//	@Description: 新建同步消息接收器
//	@param p
//	@return *SyncReceiver
func NewSyncReceiver(p cellnet.Peer) *SyncReceiver {

	self := &SyncReceiver{
		evChan: make(chan cellnet.Event),
	}

	self.callback = func(ev cellnet.Event) {

		self.evChan <- ev
	}

	return self
}
