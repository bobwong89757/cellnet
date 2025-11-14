package peer

// CoreSessionIdentify 提供会话 ID 的核心实现
// 用于为 Session 提供唯一标识符
// 所有 Session 实现都可以嵌入此结构体来获得 ID 功能
type CoreSessionIdentify struct {
	// id 会话的唯一标识符
	// 每个 Session 都有一个唯一的 64 位整数 ID
	id int64
}

// ID 获取会话 ID
// 返回会话的唯一标识符
func (self *CoreSessionIdentify) ID() int64 {
	return self.id
}

// SetID 设置会话 ID
// id: 会话的唯一标识符
// 通常由 SessionManager 在添加会话时调用
func (self *CoreSessionIdentify) SetID(id int64) {
	self.id = id
}
