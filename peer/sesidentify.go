package peer

// CoreSessionIdentify
// @Description: 会话id
type CoreSessionIdentify struct {
	id int64
}

// ID
//
//	@Description: 获取会话id
//	@receiver self
//	@return int64
func (self *CoreSessionIdentify) ID() int64 {
	return self.id
}

// SetID
//
//	@Description: 设置会话id
//	@receiver self
//	@param id
func (self *CoreSessionIdentify) SetID(id int64) {
	self.id = id
}
