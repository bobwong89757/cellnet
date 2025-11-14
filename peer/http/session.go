package http

import (
	"errors"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/peer"
	"html/template"
	"net/http"
)

// RequestMatcher 请求匹配器接口
// 用于匹配 HTTP 请求的方法和路径
type RequestMatcher interface {
	// Match 匹配请求
	// method: HTTP 方法
	// url: 请求路径
	// 返回是否匹配
	Match(method, url string) bool
}

// RespondProc 响应处理器接口
// 用于处理 HTTP 响应
type RespondProc interface {
	// WriteRespond 写入响应
	// ses: HTTP Session
	// 返回处理错误
	WriteRespond(*httpSession) error
}

var (
	// ErrUnknownOperation 未知操作错误
	// 当发送不支持的操作类型时返回
	ErrUnknownOperation = errors.New("Unknown http operation")
)

// httpSession HTTP 会话实现
// 表示一个 HTTP 请求-响应会话
// HTTP 是无状态协议，每个请求都是独立的会话
type httpSession struct {
	peer.CoreContextSet  // 上下文数据存储
	*peer.CoreProcBundle // 消息处理组件（编码器、钩子、回调等）

	// req HTTP 请求对象
	req *http.Request

	// resp HTTP 响应写入器
	resp http.ResponseWriter

	// peerInterface 所属的 Peer
	// 用于访问 Peer 的配置和功能
	peerInterface cellnet.Peer

	// t 模板对象
	// 用于渲染 HTML 模板
	t *template.Template

	// respond 是否已响应
	// 用于标记是否已经发送了响应
	respond bool

	// err 错误信息
	// 用于保存处理过程中的错误
	err error
}

// Match 匹配请求（实现 RequestMatcher 接口）
// method: HTTP 方法
// url: 请求路径
// 返回是否匹配当前请求
func (self *httpSession) Match(method, url string) bool {

	return self.req.Method == method && self.req.URL.Path == url
}

// Request 获取 HTTP 请求对象
// 返回原始的 HTTP 请求对象
func (self *httpSession) Request() *http.Request {
	return self.req
}

// Response 获取 HTTP 响应写入器
// 返回 HTTP 响应写入器，用于写入响应
func (self *httpSession) Response() http.ResponseWriter {
	return self.resp
}

// Raw 获取原始连接
// HTTP Session 没有原始连接，返回 nil
func (self *httpSession) Raw() interface{} {
	return nil
}

// ID 获取会话 ID
// HTTP Session 的 ID 始终为 0（HTTP 是无状态协议）
func (self *httpSession) ID() int64 {
	return 0
}

// Close 关闭会话
// HTTP Session 的关闭是空操作（HTTP 是无状态协议）
func (self *httpSession) Close() {
}

// Peer 获取所属的 Peer
// 返回会话所属的 Peer 对象
func (self *httpSession) Peer() cellnet.Peer {
	return self.peerInterface
}

// Send 发送响应
// raw: 响应处理器（实现 RespondProc 接口）
// 通过响应处理器写入响应，并标记已响应
func (self *httpSession) Send(raw interface{}) {

	if proc, ok := raw.(RespondProc); ok {
		// 使用响应处理器写入响应
		self.err = proc.WriteRespond(self)
		self.respond = true
	} else {
		// 不支持的操作类型
		self.err = ErrUnknownOperation
	}

}

// newHttpSession 创建新的 HTTP Session
// acc: HTTP 接受器
// req: HTTP 请求对象
// response: HTTP 响应写入器
// 返回新创建的 HTTP Session
func newHttpSession(acc *httpAcceptor, req *http.Request, response http.ResponseWriter) *httpSession {

	return &httpSession{
		req:            req,
		resp:           response,
		peerInterface:  acc,
		t:              acc.Compile(), // 编译模板
		CoreProcBundle: acc.GetBundle(),
	}
}
