package cellnet

import "time"

// UDPConnector 定义 UDP 连接器接口
// 用于创建 UDP 客户端，连接到服务器
// UDP 是无连接协议，连接器维护一个默认的 Session
type UDPConnector interface {
	GenericPeer

	// Session 获取默认会话
	// 返回当前连接的 Session
	// UDP 连接器通常只有一个 Session
	Session() Session
}

// UDPAcceptor 定义 UDP 接受器接口
// 用于创建 UDP 服务器，接受客户端数据包
// UDP 是无连接协议，需要基于地址管理 Session
type UDPAcceptor interface {
	// SetSessionTTL 设置 Session 的生存时间（TTL）
	// dur: Session 的生存时间
	// 底层使用 TTL 做 Session 生命周期管理
	// 超时时间越短，内存占用越低，但可能会过早清理活跃连接
	// 如果在此时间内没有收到来自该地址的数据包，Session 会被清理
	SetSessionTTL(dur time.Duration)
}
