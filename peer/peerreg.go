package peer

import (
	"fmt"
	"github.com/bobwong89757/cellnet"
	"sort"
)

type PeerCreateFunc func() cellnet.Peer

var peerByName = map[string]PeerCreateFunc{}

// RegisterPeerCreator
//
//	@Description: 注册Peer创建器
//	@param f
func RegisterPeerCreator(f PeerCreateFunc) {

	// 临时实例化一个，获取类型
	dummyPeer := f()

	if _, ok := peerByName[dummyPeer.TypeName()]; ok {
		panic("duplicate peer type: " + dummyPeer.TypeName())
	}

	peerByName[dummyPeer.TypeName()] = f
}

// PeerCreatorList
//
//	@Description: 获取Peer创建器列表
//	@return ret
func PeerCreatorList() (ret []string) {

	for name := range peerByName {
		ret = append(ret, name)
	}

	sort.Strings(ret)
	return
}

// getPackageByPeerName
//
//	@Description: cellnet自带的peer对应包
//	@param name
//	@return string
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

// NewPeer
//
//	@Description: 创建一个Peer
//	@param peerType
//	@return cellnet.Peer
func NewPeer(peerType string) cellnet.Peer {
	peerCreator := peerByName[peerType]
	if peerCreator == nil {
		panic(fmt.Sprintf("peer type not found '%s'\ntry to add code below:\nimport (\n  _ \"%s\"\n)\n\n",
			peerType,
			getPackageByPeerName(peerType)))
	}

	return peerCreator()
}

// NewGenericPeer
//
//	@Description: 创建Peer并设置基本属性
//	@param peerType
//	@param name
//	@param addr
//	@param q
//	@return cellnet.GenericPeer
func NewGenericPeer(peerType, name, addr string, q cellnet.EventQueue) cellnet.GenericPeer {

	p := NewPeer(peerType)
	gp := p.(cellnet.GenericPeer)
	gp.SetName(name)
	gp.SetAddress(addr)
	gp.SetQueue(q)
	return gp
}
