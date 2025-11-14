package cellnet

import (
	"sync"
)

// Pipe 是一个无界队列，用于在 goroutine 之间传递数据
// 特性：
//   - 不限制大小，添加操作不会阻塞
//   - 接收操作会阻塞等待，直到有数据可用
//   - 线程安全，支持并发读写
//
// Pipe 是 EventQueue 的底层实现，用于事件队列的消息传递
type Pipe struct {
	// list 存储队列中的元素
	list []interface{}

	// listGuard 保护 list 的互斥锁
	listGuard sync.Mutex

	// listCond 条件变量，用于在队列为空时阻塞等待
	listCond *sync.Cond
}

// Add 向队列添加一个元素
// msg: 要添加的元素，可以是任意类型
// 添加操作不会阻塞，即使队列已满也会继续添加
func (self *Pipe) Add(msg interface{}) {
	self.listGuard.Lock()
	// 将元素追加到列表末尾
	self.list = append(self.list, msg)
	self.listGuard.Unlock()

	// 通知等待的接收者
	self.listCond.Signal()
}

// Count 返回队列中当前元素的数量
// 返回队列长度，线程安全
func (self *Pipe) Count() int {
	self.listGuard.Lock()
	defer self.listGuard.Unlock()
	return len(self.list)
}

// Reset 清空队列
// 移除队列中的所有元素，但保持队列结构不变
func (self *Pipe) Reset() {
	self.listGuard.Lock()
	// 将切片长度重置为 0，但保留底层数组
	self.list = self.list[0:0]
	self.listGuard.Unlock()
}

// Pick 从队列中取出所有元素
// retList: 用于接收元素的切片指针，元素会被追加到此切片
// 返回是否应该退出（当遇到 nil 元素时返回 true）
//
// 如果队列为空，此方法会阻塞等待，直到有元素被添加
// nil 元素被视为退出信号，遇到 nil 时会停止取出并返回 true
func (self *Pipe) Pick(retList *[]interface{}) (exit bool) {
	self.listGuard.Lock()

	// 如果队列为空，阻塞等待
	for len(self.list) == 0 {
		self.listCond.Wait()
	}

	// 复制队列中的所有元素到返回列表
	for _, data := range self.list {
		// nil 元素表示退出信号
		if data == nil {
			exit = true
			break
		} else {
			// 将元素追加到返回列表
			*retList = append(*retList, data)
		}
	}

	// 清空队列，保留底层数组
	self.list = self.list[0:0]
	self.listGuard.Unlock()

	return
}

// NewPipe 创建一个新的 Pipe 实例
// 返回初始化好的 Pipe，可以直接使用
func NewPipe() *Pipe {
	self := &Pipe{}
	// 创建条件变量，关联到互斥锁
	self.listCond = sync.NewCond(&self.listGuard)

	return self
}
