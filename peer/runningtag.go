package peer

import (
	"sync"
	"sync/atomic"
)

// CoreRunningTag 提供 Peer 运行状态管理的核心实现
// 用于跟踪 Peer 的运行状态和停止流程
// 支持优雅关闭，确保停止操作完成后再继续
type CoreRunningTag struct {
	// running 运行状态标识
	// 使用原子操作保证并发安全
	// 0 表示未运行，非 0 表示正在运行
	running int64

	// stoppingWaitor 用于等待停止操作完成的同步器
	// 当开始停止时，会增加计数；停止完成时，会减少计数
	stoppingWaitor sync.WaitGroup

	// stopping 停止状态标识
	// 使用原子操作保证并发安全
	// 0 表示未停止，非 0 表示正在停止
	stopping int64
}

// IsRunning 检查 Peer 是否正在运行
// 返回 true 表示 Peer 正在运行，false 表示未运行
func (self *CoreRunningTag) IsRunning() bool {
	return atomic.LoadInt64(&self.running) != 0
}

// SetRunning 设置运行状态标识
// v: true 表示设置为运行状态，false 表示设置为未运行状态
func (self *CoreRunningTag) SetRunning(v bool) {
	if v {
		atomic.StoreInt64(&self.running, 1)
	} else {
		atomic.StoreInt64(&self.running, 0)
	}
}

// WaitStopFinished 等待停止操作完成
// 如果正在停止，会阻塞当前 goroutine 直到停止完成
// 用于确保优雅关闭，避免在停止过程中进行其他操作
func (self *CoreRunningTag) WaitStopFinished() {
	// 如果正在停止时，等待停止完成
	self.stoppingWaitor.Wait()
}

// IsStopping 检查 Peer 是否正在停止
// 返回 true 表示 Peer 正在停止，false 表示未停止
func (self *CoreRunningTag) IsStopping() bool {
	return atomic.LoadInt64(&self.stopping) != 0
}

// StartStopping 开始停止流程
// 设置停止状态标识，并增加等待计数
// 调用此方法后，其他 goroutine 可以通过 WaitStopFinished 等待停止完成
func (self *CoreRunningTag) StartStopping() {
	// 增加等待计数
	self.stoppingWaitor.Add(1)
	// 设置停止状态
	atomic.StoreInt64(&self.stopping, 1)
}

// EndStopping 结束停止流程
// 如果正在停止，会减少等待计数并清除停止状态
// 调用此方法后，等待停止的 goroutine 会被唤醒
func (self *CoreRunningTag) EndStopping() {
	if self.IsStopping() {
		// 减少等待计数，唤醒等待的 goroutine
		self.stoppingWaitor.Done()
		// 清除停止状态
		atomic.StoreInt64(&self.stopping, 0)
	}
}
