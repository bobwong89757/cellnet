package peer

// CoreRedisParameter Redis 参数核心实现
// 用于配置 Redis 连接参数
type CoreRedisParameter struct {
	// Password Redis 服务器密码
	Password string

	// DBIndex Redis 数据库索引
	// 默认为 0
	DBIndex int

	// PoolConnCount 连接池中的连接数
	// 控制连接池的大小
	PoolConnCount int
}

// Init 初始化 Redis 参数
// 设置默认连接池大小为 1
func (self *CoreRedisParameter) Init() {
	self.PoolConnCount = 1
}

// SetPassword 设置 Redis 密码
// v: Redis 服务器密码
func (self *CoreRedisParameter) SetPassword(v string) {
	self.Password = v
}

// SetDBIndex 设置 Redis 数据库索引
// v: 数据库索引，默认为 0
func (self *CoreRedisParameter) SetDBIndex(v int) {
	self.DBIndex = v
}

// SetConnectionCount 设置 Redis 连接池允许的连接数
// v: 连接池大小
func (self *CoreRedisParameter) SetConnectionCount(v int) {
	self.PoolConnCount = v
}
