package peer

import (
	"sync"
	"sync/atomic"
)

// CoreRunningTag
// @Description: 通信通讯端共享的数据
type CoreRunningTag struct {

	// 运行状态
	running int64

	stoppingWaitor sync.WaitGroup
	stopping       int64
}

// IsRunning
//
//	@Description: 是否运行
//	@receiver self
//	@return bool
func (self *CoreRunningTag) IsRunning() bool {

	return atomic.LoadInt64(&self.running) != 0
}

// SetRunning
//
//	@Description: 设置运行标识
//	@receiver self
//	@param v
func (self *CoreRunningTag) SetRunning(v bool) {

	if v {
		atomic.StoreInt64(&self.running, 1)
	} else {
		atomic.StoreInt64(&self.running, 0)
	}

}

// WaitStopFinished
//
//	@Description: 等待停止完成
//	@receiver self
func (self *CoreRunningTag) WaitStopFinished() {
	// 如果正在停止时, 等待停止完成
	self.stoppingWaitor.Wait()
}

// IsStopping
//
//	@Description: 是否停止
//	@receiver self
//	@return bool
func (self *CoreRunningTag) IsStopping() bool {
	return atomic.LoadInt64(&self.stopping) != 0
}

// StartStopping
//
//	@Description: 停止
//	@receiver self
func (self *CoreRunningTag) StartStopping() {
	self.stoppingWaitor.Add(1)
	atomic.StoreInt64(&self.stopping, 1)
}

// EndStopping
//
//	@Description: 结束停止
//	@receiver self
func (self *CoreRunningTag) EndStopping() {

	if self.IsStopping() {
		self.stoppingWaitor.Done()
		atomic.StoreInt64(&self.stopping, 0)
	}

}
