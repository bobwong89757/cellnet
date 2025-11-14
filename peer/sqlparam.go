package peer

// CoreSQLParameter SQL 参数核心实现
// 用于配置 SQL 数据库连接参数
type CoreSQLParameter struct {
	// PoolConnCount 连接池中的连接数
	// 控制连接池的大小
	PoolConnCount int
}

// Init 初始化 SQL 参数
// 设置默认连接池大小为 1
func (self *CoreSQLParameter) Init() {
	self.PoolConnCount = 1
}

// SetPassword 设置密码
// v: 数据库密码
// 当前实现为空操作（具体实现由子类提供）
func (self *CoreSQLParameter) SetPassword(v string) {
}

// SetConnectionCount 设置连接池中的连接数
// v: 连接池大小
func (self *CoreSQLParameter) SetConnectionCount(v int) {
	self.PoolConnCount = v
}
