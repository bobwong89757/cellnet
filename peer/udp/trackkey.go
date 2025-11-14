package udp

import (
	"encoding/binary"
	"net"
)

// connTrackKey 连接跟踪键
// 用于在 map 中标识 UDP 连接
// 参考: https://github.com/docker/go-connections/blob/master/proxy/udp_proxy.go
// 将 IP 地址拆分为两个字段，以便作为 map 的键使用
// 支持 IPv4 和 IPv6
type connTrackKey struct {
	// IPHigh IP 地址的高 64 位
	// IPv4 时始终为 0，IPv6 时为 IP 的前 8 字节
	IPHigh uint64

	// IPLow IP 地址的低 64 位
	// IPv4 时为整个 IP 地址（4 字节），IPv6 时为 IP 的后 8 字节
	IPLow uint64

	// Port 端口号
	Port int
}

// newConnTrackKey 根据 UDP 地址创建连接跟踪键
// addr: UDP 地址
// 返回新创建的连接跟踪键
// 支持 IPv4 和 IPv6 地址
func newConnTrackKey(addr *net.UDPAddr) *connTrackKey {
	// IPv4 地址处理
	if len(addr.IP) == net.IPv4len {
		return &connTrackKey{
			IPHigh: 0,
			// IPv4 地址只有 4 字节，放在 IPLow 中
			IPLow: uint64(binary.BigEndian.Uint32(addr.IP)),
			Port:  addr.Port,
		}
	}
	// IPv6 地址处理
	// IPv6 地址有 16 字节，分为两部分
	return &connTrackKey{
		// 前 8 字节放在 IPHigh
		IPHigh: binary.BigEndian.Uint64(addr.IP[:8]),
		// 后 8 字节放在 IPLow
		IPLow: binary.BigEndian.Uint64(addr.IP[8:]),
		Port:  addr.Port,
	}
}
