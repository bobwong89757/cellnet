package cellnet

import (
	"html/template"
	"net/http"
)

// HTTPAcceptor 定义 HTTP 接受器接口
// 用于创建 HTTP 服务器，处理 HTTP 请求
type HTTPAcceptor interface {
	GenericPeer

	// SetFileServe 设置 HTTP 文件服务
	// dir: 虚拟地址路径，如 "/static"
	// root: 文件系统根目录，如 "/var/www"
	// 用于提供静态文件服务
	SetFileServe(dir string, root string)

	// SetTemplateDir 设置模板文件目录
	// dir: 模板文件所在的目录路径
	// 用于加载 HTML 模板文件
	SetTemplateDir(dir string)

	// SetTemplateDelims 设置 HTTP 模板的分隔符
	// delimsLeft: 左分隔符，默认 "{{"
	// delimsRight: 右分隔符，默认 "}}"
	// 用于解决默认 {{ }} 与其他模板引擎冲突的问题
	SetTemplateDelims(delimsLeft, delimsRight string)

	// SetTemplateExtensions 设置模板文件的扩展名
	// exts: 扩展名列表，默认包含 ".tpl" 和 ".html"
	// 只有匹配扩展名的文件才会被识别为模板文件
	SetTemplateExtensions(exts []string)

	// SetTemplateFunc 设置模板函数
	// f: 模板函数映射列表
	// 用于在模板中调用自定义函数
	SetTemplateFunc(f []template.FuncMap)
}

// HTTPRequest 定义 HTTP 请求参数
// 用于 HTTP 连接器发送请求
type HTTPRequest struct {
	// REQMsg 请求消息对象，通常是指针类型
	REQMsg interface{}

	// ACKMsg 响应消息对象，通常是指针类型
	// 响应会解码到此对象中
	ACKMsg interface{}

	// REQCodecName 请求消息的编码器名称
	// 可为空，默认为 "json" 格式
	REQCodecName string

	// ACKCodecName 响应消息的编码器名称
	// 可为空，默认为 "json" 格式
	ACKCodecName string
}

// HTTPConnector 定义 HTTP 连接器接口
// 用于创建 HTTP 客户端，发送 HTTP 请求
type HTTPConnector interface {
	GenericPeer

	// Request 发送 HTTP 请求
	// method: HTTP 方法，如 "GET"、"POST"、"PUT"、"DELETE" 等
	// path: 请求路径，如 "/api/user"
	// param: 请求参数，包含请求消息、响应消息和编码器名称
	// 返回请求错误，如果成功则返回 nil
	Request(method, path string, param *HTTPRequest) error
}

// HTTPSession 定义 HTTP Session 接口
// 用于访问 HTTP 请求的详细信息
type HTTPSession interface {
	// Request 获取原始的 HTTP 请求对象
	// 返回 *http.Request，可以访问请求头、URL、方法等信息
	Request() *http.Request
}
