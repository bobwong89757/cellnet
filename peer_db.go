package cellnet

import "time"

// RedisPoolOperator 定义 Redis 连接池操作接口
// 用于在连接池中执行 Redis 操作
type RedisPoolOperator interface {
	// Operate 在连接池中执行操作
	// callback: 操作回调函数，参数为原始 Redis 客户端
	// 返回操作结果
	// 连接池会自动管理连接的获取和释放
	Operate(callback func(rawClient interface{}) interface{}) interface{}
}

// RedisConnector 定义 Redis 连接器接口
// 用于创建 Redis 客户端，连接到 Redis 服务器
type RedisConnector interface {
	GenericPeer

	// SetPassword 设置 Redis 密码
	// v: Redis 服务器密码
	SetPassword(v string)

	// SetConnectionCount 设置连接池中的连接数
	// v: 连接池大小
	SetConnectionCount(v int)

	// SetDBIndex 设置 Redis 数据库索引
	// v: 数据库索引，默认为 0
	SetDBIndex(v int)
}

// MySQLOperator 定义 MySQL 连接池操作接口
// 用于在连接池中执行 MySQL 操作
type MySQLOperator interface {
	// Operate 在连接池中执行操作
	// callback: 操作回调函数，参数为原始 MySQL 客户端
	// 返回操作结果
	// 连接池会自动管理连接的获取和释放
	Operate(callback func(rawClient interface{}) interface{}) interface{}
}

// MySQLConnector 定义 MySQL 连接器接口
// 用于创建 MySQL 客户端，连接到 MySQL 服务器
type MySQLConnector interface {
	GenericPeer

	// SetPassword 设置 MySQL 密码
	// v: MySQL 服务器密码
	SetPassword(v string)

	// SetConnectionCount 设置连接池中的连接数
	// v: 连接池大小
	SetConnectionCount(v int)

	// SetReconnectDuration 设置自动重连间隔
	// v: 重连时间间隔，0 为默认值，表示关闭自动重连
	SetReconnectDuration(v time.Duration)

	// ReconnectDuration 获取自动重连间隔
	// 返回当前设置的重连时间间隔
	ReconnectDuration() time.Duration
}
