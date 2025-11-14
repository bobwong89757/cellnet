package peer

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"sync"
	"sync/atomic"
)

// SessionManager 定义完整功能的会话管理接口
// 扩展了 SessionAccessor 接口，增加了添加、移除会话和设置 ID 基数的功能
// 用于管理 Peer 下的所有 Session
type SessionManager interface {
	cellnet.SessionAccessor

	// Add 添加一个会话到管理器
	// 会话会被分配一个唯一的 ID
	Add(cellnet.Session)

	// Remove 从管理器中移除一个会话
	Remove(cellnet.Session)

	// Count 返回当前会话数量
	Count() int

	// SetIDBase 设置会话 ID 的起始值
	// base: 会话 ID 的起始值，后续会话 ID 会从此值开始递增
	SetIDBase(base int64)
}

// CoreSessionManager 提供会话管理的核心实现
// 使用 sync.Map 存储会话，支持并发访问
// 自动为每个会话分配唯一的 ID
type CoreSessionManager struct {
	// sesById 使用会话 ID 关联会话的映射表
	// 键为会话 ID（int64），值为 Session
	sesById sync.Map

	// sesIDGen 记录已经生成的会话 ID 流水号
	// 使用原子操作保证并发安全
	sesIDGen int64

	// count 记录当前在使用的会话数量
	// 使用原子操作保证并发安全
	count int64
}

// SetIDBase 设置会话 ID 的起始值
// base: 会话 ID 的起始值
// 后续添加的会话 ID 会从此值开始递增
func (self *CoreSessionManager) SetIDBase(base int64) {
	atomic.StoreInt64(&self.sesIDGen, base)
}

// Count 获取当前会话数量
// 返回活跃的会话数量
func (self *CoreSessionManager) Count() int {
	return int(atomic.LoadInt64(&self.count))
}

// Add 添加一个会话到管理器
// ses: 要添加的会话
// 会话会被分配一个唯一的 ID，并存储到管理器中
func (self *CoreSessionManager) Add(ses cellnet.Session) {
	// 生成新的会话 ID（原子递增）
	id := atomic.AddInt64(&self.sesIDGen, 1)

	// 增加会话计数
	atomic.AddInt64(&self.count, 1)

	// 设置会话 ID
	ses.(interface {
		SetID(int64)
	}).SetID(id)
	log.GetLog().Warnf("添加%d到sessMgr", id)
	// 存储会话到映射表
	self.sesById.Store(id, ses)
}

// Remove 从管理器中移除一个会话
// ses: 要移除的会话
// 会话会从映射表中删除，并减少会话计数
func (self *CoreSessionManager) Remove(ses cellnet.Session) {
	// 从映射表中删除会话
	self.sesById.Delete(ses.ID())

	// 减少会话计数
	atomic.AddInt64(&self.count, -1)
}

// GetSession 通过会话 ID 获取一个会话
// id: 会话的唯一标识符
// 返回对应的 Session，如果不存在返回 nil
func (self *CoreSessionManager) GetSession(id int64) cellnet.Session {
	if v, ok := self.sesById.Load(id); ok {
		return v.(cellnet.Session)
	}

	return nil
}

// VisitSession 遍历所有会话
// callback: 遍历回调函数，参数为 Session
// 如果回调返回 false，则停止遍历
func (self *CoreSessionManager) VisitSession(callback func(cellnet.Session) bool) {
	// 使用 sync.Map 的 Range 方法遍历所有会话
	self.sesById.Range(func(key, value interface{}) bool {
		// 调用回调函数
		return callback(value.(cellnet.Session))
	})
}

// CloseAllSession 关闭所有会话
// 遍历所有会话并调用其 Close 方法
func (self *CoreSessionManager) CloseAllSession() {
	self.VisitSession(func(ses cellnet.Session) bool {
		// 关闭会话
		ses.Close()
		// 继续遍历
		return true
	})
}

// SessionCount 返回活跃的会话数量
// 返回当前在使用的会话数量
// 与 Count 方法功能相同，提供一致的接口
func (self *CoreSessionManager) SessionCount() int {
	v := atomic.LoadInt64(&self.count)
	return int(v)
}
