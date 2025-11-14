package util

import (
	"errors"
	"fmt"
	"github.com/bobwong89757/cellnet"
	"net"
	"strconv"
	"strings"
)

// SpliteAddress 将普通地址格式(host:port)拆分为主机和端口
// addr: 地址字符串，格式为 "host:port"
// 返回主机名、端口号和错误信息
// 如果地址格式无效，返回错误
func SpliteAddress(addr string) (host string, port int, err error) {
	var portStr string

	// 使用标准库拆分主机和端口
	host, portStr, err = net.SplitHostPort(addr)

	if err != nil {
		return "", 0, err
	}

	// 将端口字符串转换为整数
	port, err = strconv.Atoi(portStr)

	if err != nil {
		return "", 0, err
	}

	return
}

// JoinAddress 将主机和端口合并为(host:port)格式的地址
// host: 主机名或 IP 地址
// port: 端口号
// 返回格式化的地址字符串
func JoinAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// RemoteAddr 定义获取远程地址的接口
// 用于修复 WebSocket 没有实现所有 net.Conn 方法，导致无法获取客户端地址的问题
type RemoteAddr interface {
	RemoteAddr() net.Addr
}

// GetRemoteAddrss 获取 Session 的远程地址
// ses: 要获取远程地址的 Session
// 返回远程地址字符串和是否成功获取
// 如果 Session 为 nil 或无法获取地址，返回 false
func GetRemoteAddrss(ses cellnet.Session) (string, bool) {
	if ses == nil {
		return "", false
	}

	// 尝试从原始连接获取远程地址
	if c, ok := ses.Raw().(RemoteAddr); ok {
		return c.RemoteAddr().String(), true
	}

	return "", false
}

var (
	// ErrInvalidPortRange 表示端口范围无效的错误
	ErrInvalidPortRange = errors.New("invalid port range")

	// ErrInvalidAddressFormat 表示地址格式无效的错误
	ErrInvalidAddressFormat = errors.New("invalid address format")
)

// Address 支持地址范围的地址格式
// 可以表示一个地址范围，用于端口自动检测等功能
type Address struct {
	// Scheme 协议方案，如 "tcp"、"udp"、"http" 等
	Scheme string

	// Host 主机名或 IP 地址
	Host string

	// MinPort 最小端口号
	MinPort int

	// MaxPort 最大端口号
	MaxPort int

	// Path 路径部分，用于 HTTP 等协议
	Path string
}

// HostPortString 返回(host:port)格式的地址
// port: 端口号
// 返回格式化的地址字符串，不包含协议方案和路径
func (self *Address) HostPortString(port int) string {
	return fmt.Sprintf("%s:%d", self.Host, port)
}

// String 返回完整的地址字符串
// port: 端口号
// 返回 scheme://host:port/path 格式的地址
// 如果 Scheme 为空，则返回 host:port 格式
func (self *Address) String(port int) string {
	if self.Scheme == "" {
		return self.HostPortString(port)
	}

	return fmt.Sprintf("%s://%s:%d%s", self.Scheme, self.Host, port, self.Path)
}

// ParseAddress
//
//	@Description: cellnet专有的地址格式 scheme://host:minPort~maxPort/path  提供地址范围扩展
//	@param addr
//	@return addrObj
//	@return err
func ParseAddress(addr string) (addrObj *Address, err error) {
	addrObj = new(Address)

	schemePos := strings.Index(addr, "://")

	// 移除scheme部分
	if schemePos != -1 {
		addrObj.Scheme = addr[:schemePos]
		addr = addr[schemePos+3:]
	}

	// 冒号不可选
	colonPos := strings.Index(addr, ":")

	if colonPos != -1 {
		addrObj.Host = addr[:colonPos]
		addr = addr[colonPos+1:]
	} else {
		return nil, ErrInvalidAddressFormat
	}

	rangePos := strings.Index(addr, "~")

	var minStr, maxStr string
	if rangePos != -1 {
		minStr = addr[:rangePos]

		slashPos := strings.Index(addr, "/")

		if slashPos != -1 {
			maxStr = addr[rangePos+1 : slashPos]
			addrObj.Path = addr[slashPos:]
		} else {
			maxStr = addr[rangePos+1:]
		}
	} else {
		slashPos := strings.Index(addr, "/")

		if slashPos != -1 {
			addrObj.Path = addr[slashPos:]
			minStr = addr[rangePos+1 : slashPos]
		} else {
			minStr = addr[rangePos+1:]
		}
	}

	// extract min port
	addrObj.MinPort, err = strconv.Atoi(minStr)
	if err != nil {
		return nil, ErrInvalidPortRange
	}

	if maxStr != "" {
		// extract max port
		addrObj.MaxPort, err = strconv.Atoi(maxStr)
		if err != nil {
			return nil, ErrInvalidPortRange
		}
	} else {
		addrObj.MaxPort = addrObj.MinPort
	}

	return
}

// DetectPort
//
//	@Description: 在给定的端口范围内找到一个能用的端口 addr格式参考ParseAddress函数
//	@param addr
//	@param fn
//	@return interface{}
//	@return error
func DetectPort(addr string, fn func(a *Address, port int) (interface{}, error)) (interface{}, error) {

	addrObj, err := ParseAddress(addr)
	if err != nil {
		return nil, err
	}

	for port := addrObj.MinPort; port <= addrObj.MaxPort; port++ {

		// 使用回调侦听
		ln, err := fn(addrObj, port)
		if err == nil {
			return ln, nil
		}

		// hit max port
		if port == addrObj.MaxPort {
			return nil, err
		}
	}

	return nil, fmt.Errorf("unable to bind to %s", addr)
}

// GetLocalIP
//
//	@Description: 获取本地IP地址，有多重IP时，默认取第一个
//	@return string
func GetLocalIP() string {

	// TODO 全面支持IPV6地址
	list, err := GetPrivateIPv4()
	if err != nil {
		return ""
	}

	if len(list) == 0 {
		return ""
	}

	return list[0].String()
}

// GetPrivateIPv4
//
//	@Description: 获得本机的IPV4的地址
//	@return []*net.IPAddr
//	@return error
func GetPrivateIPv4() ([]*net.IPAddr, error) {
	addresses, err := activeInterfaceAddresses()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
	}

	var addrs []*net.IPAddr
	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if ip.To4() == nil {
			continue
		}
		if !isPrivate(ip) {
			continue
		}
		addrs = append(addrs, &net.IPAddr{IP: ip})
	}
	return addrs, nil
}

// GetPublicIPv6
//
//	@Description: 获得本机的IPV6地址
//	@return []*net.IPAddr
//	@return error
func GetPublicIPv6() ([]*net.IPAddr, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
	}

	var addrs []*net.IPAddr
	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if ip.To4() != nil {
			continue
		}
		if isPrivate(ip) {
			continue
		}
		addrs = append(addrs, &net.IPAddr{IP: ip})
	}
	return addrs, nil
}

// privateBlocks contains non-forwardable address blocks which are used
// for private networks. RFC 6890 provides an overview of special
// address blocks.
var privateBlocks = []*net.IPNet{
	parseCIDR("10.0.0.0/8"),     // RFC 1918 IPv4 private network address
	parseCIDR("100.64.0.0/10"),  // RFC 6598 IPv4 shared address space
	parseCIDR("127.0.0.0/8"),    // RFC 1122 IPv4 loopback address
	parseCIDR("169.254.0.0/16"), // RFC 3927 IPv4 link local address
	parseCIDR("172.16.0.0/12"),  // RFC 1918 IPv4 private network address
	parseCIDR("192.0.0.0/24"),   // RFC 6890 IPv4 IANA address
	parseCIDR("192.0.2.0/24"),   // RFC 5737 IPv4 documentation address
	parseCIDR("192.168.0.0/16"), // RFC 1918 IPv4 private network address
	parseCIDR("::1/128"),        // RFC 1884 IPv6 loopback address
	parseCIDR("fe80::/10"),      // RFC 4291 IPv6 link local addresses
	parseCIDR("fc00::/7"),       // RFC 4193 IPv6 unique local addresses
	parseCIDR("fec0::/10"),      // RFC 1884 IPv6 site-local addresses
	parseCIDR("2001:db8::/32"),  // RFC 3849 IPv6 documentation address
}

func parseCIDR(s string) *net.IPNet {
	_, block, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("Bad CIDR %s: %s", s, err))
	}
	return block
}

func isPrivate(ip net.IP) bool {
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

// Returns addresses from interfaces that is up
func activeInterfaceAddresses() ([]net.Addr, error) {
	var upAddrs []net.Addr
	var loAddrs []net.Addr

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Require interface to be up
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
		}

		if iface.Flags&net.FlagLoopback != 0 {
			loAddrs = append(loAddrs, addresses...)
			continue
		}

		upAddrs = append(upAddrs, addresses...)
	}

	if len(upAddrs) == 0 {
		return loAddrs, nil
	}

	return upAddrs, nil
}
