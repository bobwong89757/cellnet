package peer

import (
	"fmt"
	"sort"

	"github.com/bobwong89757/cellnet"
)

// PeerCreateFunc 是 Peer 创建函数的类型
// 用于创建特定类型的 Peer 实例
type PeerCreateFunc func() cellnet.Peer

// peerByName 存储所有已注册的 Peer 创建器
// 键为 Peer 类型名称（如 "tcp.Acceptor"、"udp.Connector"），值为对应的创建函数
var peerByName = map[string]PeerCreateFunc{}

// RegisterPeerCreator 注册 Peer 创建器
// f: Peer 创建函数，用于创建特定类型的 Peer 实例
// 如果 Peer 类型已存在，会触发 panic
// 通常在 Peer 包的 init() 函数中调用此函数进行注册
func RegisterPeerCreator(f PeerCreateFunc) {
	// 临时实例化一个 Peer，获取其类型名称
	dummyPeer := f()

	// 检查 Peer 类型是否已注册
	if _, ok := peerByName[dummyPeer.TypeName()]; ok {
		panic("duplicate peer type: " + dummyPeer.TypeName())
	}

	// 注册 Peer 创建器
	peerByName[dummyPeer.TypeName()] = f
}

// PeerCreatorList 获取所有已注册的 Peer 类型名称列表
// 返回 Peer 类型名称的切片，按字母顺序排序
// 用于调试和查看可用的 Peer 类型
func PeerCreatorList() (ret []string) {
	// 收集所有 Peer 类型名称
	for name := range peerByName {
		ret = append(ret, name)
	}

	// 按字母顺序排序
	sort.Strings(ret)
	return
}

// getPackageByPeerName 根据 Peer 类型名称返回对应的包路径
// name: Peer 的类型名称
// 返回 Peer 所在的包路径，用于错误提示
// 这是 cellnet 自带的 Peer 对应的包路径
func getPackageByPeerName(name string) string {
	switch name {
	case "tcp.Connector", "tcp.Acceptor", "tcp.SyncConnector":
		return "github.com/bobwong89757/cellnet/peer/tcp"
	case "udp.Connector", "udp.Acceptor":
		return "github.com/bobwong89757/cellnet/peer/udp"
	case "gorillaws.Acceptor", "gorillaws.Connector", "gorillaws.SyncConnector":
		return "github.com/bobwong89757/cellnet/peer/gorillaws"
	case "http.Connector", "http.Acceptor":
		return "github.com/bobwong89757/cellnet/peer/http"
	case "redix.Connector":
		return "github.com/bobwong89757/cellnet/peer/redix"
	case "mysql.Connector":
		return "github.com/bobwong89757/cellnet/peer/mysql"
	case "kcp.Connector", "kcp.Acceptor", "kcp.SyncConnector":
		return "github.com/bobwong89757/cellnet/peer/kcp"
	default:
		return "package/to/your/peer"
	}
}

// NewPeer 创建一个指定类型的 Peer
// peerType: Peer 的类型名称，如 "tcp.Acceptor"、"udp.Connector" 等
// 返回创建的 Peer 实例
// 如果 Peer 类型不存在，会触发 panic 并提供导入提示信息
func NewPeer(peerType string) cellnet.Peer {
	// 查找 Peer 创建器
	peerCreator := peerByName[peerType]
	if peerCreator == nil {
		// Peer 类型不存在，提供友好的错误提示
		panic(fmt.Sprintf("peer type not found '%s'\ntry to add code below:\nimport (\n  _ \"%s\"\n)\n\n",
			peerType,
			getPackageByPeerName(peerType)))
	}

	// 调用创建器创建 Peer
	return peerCreator()
}

// NewGenericPeer 创建 Peer 并设置基本属性
// peerType: Peer 的类型名称，如 "tcp.Acceptor"、"udp.Connector" 等
// name: Peer 的名称，用于标识和日志记录
// addr: Peer 的地址，对于 Acceptor 是监听地址，对于 Connector 是连接地址
// q: Peer 关联的事件队列，用于异步处理网络事件
// 返回配置好的 GenericPeer
// 这是一个便捷函数，创建 Peer 并一次性设置常用属性
func NewGenericPeer(peerType, name, addr string, q cellnet.EventQueue) cellnet.GenericPeer {
	// 创建 Peer
	p := NewPeer(peerType)
	// 转换为 GenericPeer
	gp := p.(cellnet.GenericPeer)
	// 设置名称
	gp.SetName(name)
	// 设置地址
	gp.SetAddress(addr)
	// 设置事件队列
	gp.SetQueue(q)
	return gp
}
