package util

// Queue 简单的队列实现
// 使用切片实现，支持入队、出队、查看等操作
// 注意：此队列不是线程安全的，如果需要在并发环境中使用，需要外部加锁
type Queue struct {
	// list 存储队列元素的切片
	list []interface{}
}

// Enqueue 将元素加入队列
// data: 要加入队列的元素，可以是任意类型
// 元素会被添加到队列末尾
func (self *Queue) Enqueue(data interface{}) {
	self.list = append(self.list, data)
}

// Count 返回队列中元素的数量
// 返回队列长度
func (self *Queue) Count() int {
	return len(self.list)
}

// Peek 查看队列头部的元素，但不移除
// 返回队列头部的元素
// 注意：如果队列为空，可能会 panic
func (self *Queue) Peek() interface{} {
	return self.list[0]
}

// Dequeue 从队列头部移除并返回元素
// 返回队列头部的元素，如果队列为空返回 nil
func (self *Queue) Dequeue() (ret interface{}) {
	// 如果队列为空，返回 nil
	if len(self.list) == 0 {
		return nil
	}

	// 获取队列头部元素
	ret = self.list[0]

	// 移除队列头部元素
	self.list = self.list[1:]

	return
}

// Clear 清空队列
// 移除队列中的所有元素，但保留底层数组
func (self *Queue) Clear() {
	self.list = self.list[0:0]
}

// NewQueue 创建一个新的队列
// size: 队列的初始容量
// 返回初始化好的 Queue
func NewQueue(size int) *Queue {
	return &Queue{
		list: make([]interface{}, 0, size),
	}
}
