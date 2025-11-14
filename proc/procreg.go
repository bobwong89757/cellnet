package proc

import (
	"fmt"
	"github.com/bobwong89757/cellnet"
	"sort"
)

// ProcessorBinder 是处理器绑定函数的类型
// 用于配置 ProcessorBundle，设置消息传输器、事件钩子和用户回调
// bundle: 处理器资源包，用于配置消息处理流程
// userCallback: 用户定义的事件处理回调函数
// args: 可选的额外参数
type ProcessorBinder func(bundle ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{})

var (
	// procByName 存储所有已注册的处理器
	// 键为处理器名称（如 "tcp.ltv"、"udp.ltv"），值为对应的绑定函数
	procByName = map[string]ProcessorBinder{}
)

// RegisterProcessor 注册事件处理器
// procName: 处理器的名称，如 "tcp.ltv"、"udp.ltv"、"http" 等
// f: 处理器绑定函数，用于配置消息处理流程
// 如果处理器名称已存在，会触发 panic
// 通常在处理器包的 init() 函数中调用此函数进行注册
func RegisterProcessor(procName string, f ProcessorBinder) {
	// 检查处理器名称是否已存在
	if _, ok := procByName[procName]; ok {
		panic("duplicate peer type: " + procName)
	}

	// 注册处理器
	procByName[procName] = f
}

// ProcessorList 获取所有已注册的处理器名称列表
// 返回处理器名称的切片，按字母顺序排序
// 用于调试和查看可用的处理器
func ProcessorList() (ret []string) {
	// 收集所有处理器名称
	for name := range procByName {
		ret = append(ret, name)
	}

	// 按字母顺序排序
	sort.Strings(ret)
	return
}

// getPackageByCodecName 根据处理器名称返回对应的包路径
// name: 处理器的名称
// 返回处理器所在的包路径，用于错误提示
// 这是 cellnet 自带的处理器对应的包路径
func getPackageByCodecName(name string) string {
	switch name {
	case "gorillaws.ltv":
		return "github.com/bobwong89757/cellnet/proc/gorillaws"
	case "http":
		return "github.com/bobwong89757/cellnet/proc/http"
	case "tcp.ltv":
		return "github.com/bobwong89757/cellnet/proc/tcp"
	case "udp.ltv":
		return "github.com/bobwong89757/cellnet/proc/udp"
	default:
		return "package/to/your/proc"
	}
}

// BindProcessorHandler 将处理器绑定到 Peer
// peer: 要绑定处理器的 Peer
// procName: 处理器的名称，来源于 RegisterProcessor 注册的处理器，形如 "tcp.ltv"、"udp.ltv"
// userCallback: 用户定义的事件处理回调函数
// args: 可选的额外参数，传递给处理器绑定函数
// 如果处理器不存在，会触发 panic 并提供导入提示信息
func BindProcessorHandler(peer cellnet.Peer, procName string, userCallback cellnet.EventCallback, args ...interface{}) {
	// 查找处理器
	if proc, ok := procByName[procName]; ok {
		// 将 Peer 转换为 ProcessorBundle
		bundle := peer.(ProcessorBundle)

		// 调用处理器绑定函数，配置消息处理流程
		proc(bundle, userCallback, args...)

	} else {
		// 处理器不存在，提供友好的错误提示
		panic(fmt.Sprintf("processor not found '%s'\ntry to add code below:\nimport (\n  _ \"%s\"\n)\n\n",
			procName,
			getPackageByCodecName(procName)))
	}
}
