package peer

// CoreRedisParameter
// @Description: redis参数
type CoreRedisParameter struct {
	Password      string
	DBIndex       int
	PoolConnCount int
}

func (self *CoreRedisParameter) Init() {
	self.PoolConnCount = 1
}

// SetPassword
//
//	@Description: 设置redis密码
//	@receiver self
//	@param v
func (self *CoreRedisParameter) SetPassword(v string) {
	self.Password = v
}

// SetDBIndex
//
//	@Description: 设置redis数据库index
//	@receiver self
//	@param v
func (self *CoreRedisParameter) SetDBIndex(v int) {
	self.DBIndex = v
}

// SetConnectionCount
//
//	@Description: 设置redis连接池允许的数目
//	@receiver self
//	@param v
func (self *CoreRedisParameter) SetConnectionCount(v int) {
	self.PoolConnCount = v
}
