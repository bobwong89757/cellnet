package peer

import (
	"reflect"
	"sync"
)

// CoreContextSet 提供上下文数据存储和访问的核心实现
// 用于绑定用户自定义数据，支持任意类型的键值对
// 线程安全，支持并发访问
// 使用 map 实现，提供 O(1) 的查找、设置和删除性能
type CoreContextSet struct {
	// ctxes 存储上下文数据的 map
	// 使用 map 而不是 slice，提供更好的性能（O(1) vs O(n)）
	ctxes map[interface{}]interface{}

	// ctxesGuard 保护 ctxes 的读写锁
	// 用于并发安全地访问上下文数据
	ctxesGuard sync.RWMutex
}

// FetchContext 根据值的类型自动获取上下文并设置到值指针
// key: 上下文数据的键
// valuePtr: 指向目标值的指针，类型会自动匹配
// 返回是否成功获取并设置
// 支持常见类型的自动转换：string、int、int32、int64、uint、uint32、uint64、bool、float32、float64、[]byte
// 对于其他类型，使用反射进行设置
func (self *CoreContextSet) FetchContext(key, valuePtr interface{}) bool {
	// 获取上下文数据
	pv, ok := self.GetContext(key)
	if !ok {
		return false
	}

	// 根据值指针的类型进行类型断言和设置
	switch rawValue := valuePtr.(type) {
	case *string:
		*rawValue = pv.(string)
	case *int:
		*rawValue = pv.(int)
	case *int32:
		*rawValue = pv.(int32)
	case *int64:
		*rawValue = pv.(int64)
	case *uint:
		*rawValue = pv.(uint)
	case *uint32:
		*rawValue = pv.(uint32)
	case *uint64:
		*rawValue = pv.(uint64)
	case *bool:
		*rawValue = pv.(bool)
	case *float32:
		*rawValue = pv.(float32)
	case *float64:
		*rawValue = pv.(float64)
	case *[]byte:
		*rawValue = pv.([]byte)
	default:
		// 对于其他类型，使用反射进行设置
		v := reflect.Indirect(reflect.ValueOf(valuePtr))

		// 避免 call of reflect.Value.Set on zero Value
		if pv == nil {
			// 如果值为 nil，设置为零值
			v.Set(reflect.Zero(v.Type()))
		} else {
			// 设置值
			v.Set(reflect.ValueOf(pv))
		}
	}

	return true
}

// GetContext 获取上下文数据
// key: 上下文数据的键
// 返回上下文数据的值和是否存在
// 如果键不存在，返回 nil, false
// 时间复杂度: O(1)
func (self *CoreContextSet) GetContext(key interface{}) (interface{}, bool) {
	self.ctxesGuard.RLock()
	defer self.ctxesGuard.RUnlock()

	// 如果 map 未初始化，返回不存在
	if self.ctxes == nil {
		return nil, false
	}

	// 直接从 map 中查找
	value, ok := self.ctxes[key]
	return value, ok
}

// SetContext 设置上下文数据
// key: 上下文数据的键，可以是任意类型
// v: 上下文数据的值，可以是任意类型
// 如果键已存在，则更新其值；否则添加新的键值对
// 时间复杂度: O(1)
func (self *CoreContextSet) SetContext(key, v interface{}) {
	self.ctxesGuard.Lock()
	defer self.ctxesGuard.Unlock()

	// 如果 map 未初始化，先初始化
	if self.ctxes == nil {
		self.ctxes = make(map[interface{}]interface{})
	}

	// 直接设置到 map 中（如果 key 已存在会自动更新）
	self.ctxes[key] = v
}

// RemoveContext 删除指定 key 的上下文数据
// key: 要删除的上下文数据的键
// 返回是否成功删除，如果 key 不存在返回 false
// 线程安全，支持并发调用
// 时间复杂度: O(1)
func (self *CoreContextSet) RemoveContext(key interface{}) bool {
	self.ctxesGuard.Lock()
	defer self.ctxesGuard.Unlock()

	// 如果 map 未初始化，返回 false
	if self.ctxes == nil {
		return false
	}

	// 检查 key 是否存在
	if _, exists := self.ctxes[key]; exists {
		// 从 map 中删除
		delete(self.ctxes, key)
		return true
	}

	return false
}

// ClearContext 清理所有上下文数据
// 释放所有存储的上下文数据
// 建议在业务层处理 SessionClosed 事件时调用，在读取完需要的 context 数据后清理
// 这样可以避免竞态条件，确保业务层能够安全地读取 context
// 线程安全，支持并发调用
// 时间复杂度: O(1)
func (self *CoreContextSet) ClearContext() {
	self.ctxesGuard.Lock()
	defer self.ctxesGuard.Unlock()

	// 清空 map（设置为 nil，让 GC 回收）
	self.ctxes = nil
}

// GetAllContext 获取所有上下文数据的副本
// 返回一个包含所有 key-value 对的 map
// 线程安全，返回的是数据的快照，不会影响原始数据
// 用于调试和打印 context 信息
// 时间复杂度: O(n)，n 为 context 数量
func (self *CoreContextSet) GetAllContext() map[interface{}]interface{} {
	self.ctxesGuard.RLock()
	defer self.ctxesGuard.RUnlock()

	// 如果 map 未初始化，返回空的 map
	if self.ctxes == nil {
		return make(map[interface{}]interface{})
	}

	// 创建 map 存储所有上下文数据的副本
	result := make(map[interface{}]interface{}, len(self.ctxes))

	// 复制所有上下文数据
	for key, value := range self.ctxes {
		result[key] = value
	}

	return result
}
