package peer

import (
	"reflect"
	"sync"
)

// ctx 用于存储上下文数据的键值对
type ctx struct {
	// key 上下文数据的键，可以是任意类型
	key interface{}

	// value 上下文数据的值，可以是任意类型
	value interface{}
}

// CoreContextSet 提供上下文数据存储和访问的核心实现
// 用于绑定用户自定义数据，支持任意类型的键值对
// 线程安全，支持并发访问
type CoreContextSet struct {
	// ctxes 存储上下文数据的列表
	ctxes []ctx

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
func (self *CoreContextSet) GetContext(key interface{}) (interface{}, bool) {
	self.ctxesGuard.RLock()
	defer self.ctxesGuard.RUnlock()

	// 遍历上下文列表查找匹配的键
	for _, t := range self.ctxes {
		if t.key == key {
			return t.value, true
		}
	}

	return nil, false
}

// SetContext 设置上下文数据
// key: 上下文数据的键，可以是任意类型
// v: 上下文数据的值，可以是任意类型
// 如果键已存在，则更新其值；否则添加新的键值对
func (self *CoreContextSet) SetContext(key, v interface{}) {
	self.ctxesGuard.Lock()
	defer self.ctxesGuard.Unlock()

	// 查找是否已存在相同键的上下文
	for i, t := range self.ctxes {
		if t.key == key {
			// 更新已存在的上下文数据
			self.ctxes[i] = ctx{key, v}
			return
		}
	}

	// 添加新的上下文数据
	self.ctxes = append(self.ctxes, ctx{key, v})
}
