package http

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
)

// StatusRespond 状态码响应处理器
// 实现 RespondProc 接口，用于只返回 HTTP 状态码（无响应体）
type StatusRespond struct {
	// StatusCode HTTP 状态码
	StatusCode int
}

// WriteRespond 写入响应（实现 RespondProc 接口）
// ses: HTTP Session
// 只写入 HTTP 状态码，不写入响应体
// 返回处理错误
func (self *StatusRespond) WriteRespond(ses *httpSession) error {

	peerInfo := ses.Peer().(cellnet.PeerProperty)

	// 记录日志
	log.GetLog().Debugf("#http.recv(%s) '%s' %s | [%d] Status",
		peerInfo.Name(),
		ses.req.Method,
		ses.req.URL.Path,
		self.StatusCode)

	// 只写入状态码
	ses.resp.WriteHeader(int(self.StatusCode))
	return nil
}
