package peer

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"sync"
	"sync/atomic"
)

// SessionManager
// @Description: 完整功能的会话管理
type SessionManager interface {
	cellnet.SessionAccessor

	Add(cellnet.Session)
	Remove(cellnet.Session)
	Count() int

	// 设置ID开始的号
	SetIDBase(base int64)
}

type CoreSessionManager struct {
	sesById sync.Map // 使用Id关联会话

	sesIDGen int64 // 记录已经生成的会话ID流水号

	count int64 // 记录当前在使用的会话数量
}

// SetIDBase
//
//	@Description: 设置初始递增会话id
//	@receiver self
//	@param base
func (self *CoreSessionManager) SetIDBase(base int64) {

	atomic.StoreInt64(&self.sesIDGen, base)
}

// Count
//
//	@Description: 获取会话数目
//	@receiver self
//	@return int
func (self *CoreSessionManager) Count() int {
	return int(atomic.LoadInt64(&self.count))
}

// Add
//
//	@Description: 添加会话
//	@receiver self
//	@param ses
func (self *CoreSessionManager) Add(ses cellnet.Session) {

	id := atomic.AddInt64(&self.sesIDGen, 1)

	atomic.AddInt64(&self.count, 1)

	ses.(interface {
		SetID(int64)
	}).SetID(id)
	log.GetLog().Warnf("添加%d到sessMgr", id)
	self.sesById.Store(id, ses)
}

// Remove
//
//	@Description: 移除会话
//	@receiver self
//	@param ses
func (self *CoreSessionManager) Remove(ses cellnet.Session) {

	self.sesById.Delete(ses.ID())

	atomic.AddInt64(&self.count, -1)
}

// GetSession
//
//	@Description: 通过会话id获取一个连接
//	@receiver self
//	@param id
//	@return cellnet.Session
func (self *CoreSessionManager) GetSession(id int64) cellnet.Session {
	if v, ok := self.sesById.Load(id); ok {
		return v.(cellnet.Session)
	}

	return nil
}

// VisitSession
//
//	@Description: 遍历会话
//	@receiver self
//	@param callback
func (self *CoreSessionManager) VisitSession(callback func(cellnet.Session) bool) {

	self.sesById.Range(func(key, value interface{}) bool {

		return callback(value.(cellnet.Session))

	})
}

// CloseAllSession
//
//	@Description: 关闭所有会话
//	@receiver self
func (self *CoreSessionManager) CloseAllSession() {

	self.VisitSession(func(ses cellnet.Session) bool {

		ses.Close()

		return true
	})
}

// SessionCount
//
//	@Description: 活跃的会话数量
//	@receiver self
//	@return int
func (self *CoreSessionManager) SessionCount() int {

	v := atomic.LoadInt64(&self.count)
	return int(v)
}
