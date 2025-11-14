package http

import (
	"errors"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"github.com/bobwong89757/cellnet/util"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"
)

// httpAcceptor HTTP 接受器实现
// 用于创建 HTTP 服务器，处理 HTTP 请求
// 支持消息处理、静态文件服务和模板渲染
type httpAcceptor struct {
	peer.CorePeerProperty // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreProcBundle   // 消息处理组件（编码器、钩子、回调等）
	peer.CoreContextSet   // 上下文数据存储

	// sv HTTP 服务器
	sv *http.Server

	// httpDir 静态文件服务的虚拟路径
	// httpRoot 静态文件服务的根目录
	httpDir  string
	httpRoot string

	// templateDir 模板文件目录
	templateDir string

	// delimsLeft 模板左分隔符
	// delimsRight 模板右分隔符
	delimsLeft  string
	delimsRight string

	// templateExts 模板文件扩展名列表
	templateExts []string

	// templateFuncs 模板函数映射列表
	templateFuncs []template.FuncMap

	// listener TCP 监听器
	listener net.Listener
}

var (
	// errNotFound 404 错误
	errNotFound = errors.New("404 Not found")
)

// tcpKeepAliveListener TCP Keep-Alive 监听器
// 包装 TCPListener，自动为接受的连接启用 Keep-Alive
type tcpKeepAliveListener struct {
	*net.TCPListener
}

// Accept 接受连接并启用 Keep-Alive
// 返回启用了 Keep-Alive 的 TCP 连接
func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	// 启用 Keep-Alive
	tc.SetKeepAlive(true)
	// 设置 Keep-Alive 周期为 3 分钟
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// Port 获取当前侦听的端口号
// 如果 listener 未初始化，返回 0
// 返回当前 TCP 监听器绑定的端口号
func (self *httpAcceptor) Port() int {
	if self.listener == nil {
		return 0
	}

	return self.listener.Addr().(*net.TCPAddr).Port
}

// IsReady 检查接受器是否已准备好
// 返回 true 表示接受器正在运行并可以接受请求
// 返回 false 表示接受器未运行
func (self *httpAcceptor) IsReady() bool {
	return self.Port() != 0
}

// WANAddress 获取 WAN（广域网）地址
// 如果地址中的主机为空，使用本地 IP 地址
// 返回完整的 "host:port" 格式地址
func (self *httpAcceptor) WANAddress() string {

	pos := strings.Index(self.Address(), ":")
	if pos == -1 {
		return self.Address()
	}

	host := self.Address()[:pos]

	// 如果主机为空，使用本地 IP
	if host == "" {
		host = util.GetLocalIP()
	}

	return util.JoinAddress(host, self.Port())
}

// Start 异步开始侦听 HTTP 请求
// 如果地址中包含端口 0，会自动分配可用端口
// 启动成功后会在后台 goroutine 中处理 HTTP 请求
// 返回自身以支持链式调用
func (self *httpAcceptor) Start() cellnet.Peer {

	// 创建 HTTP 服务器，使用自身作为处理器
	self.sv = &http.Server{Addr: self.Address(), Handler: self}

	// 尝试监听指定地址，如果端口为 0 则自动分配
	ln, err := util.DetectPort(self.Address(), func(a *util.Address, port int) (interface{}, error) {
		return net.Listen("tcp", a.HostPortString(port))
	})

	if err != nil {
		// 监听失败，记录错误
		log.GetLog().Errorf("#http.listen failed(%s) %v", self.Name(), err.Error())

		return self
	}

	self.listener = ln.(net.Listener)

	log.GetLog().Infof("#http.listen(%s) http://%s", self.Name(), self.WANAddress())

	// 在后台 goroutine 中启动 HTTP 服务器
	go func() {

		// 使用 Keep-Alive 监听器启动服务器
		err = self.sv.Serve(tcpKeepAliveListener{self.listener.(*net.TCPListener)})
		if err != nil && err != http.ErrServerClosed {
			log.GetLog().Errorf("#http.listen failed(%s) %v", self.Name(), err.Error())
		}

	}()

	return self
}

// ServeHTTP 处理 HTTP 请求（实现 http.Handler 接口）
// res: HTTP 响应写入器
// req: HTTP 请求对象
// 处理流程：1. 创建 Session 2. 处理消息 3. 处理静态文件 4. 返回响应
func (self *httpAcceptor) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	// 创建 HTTP Session
	ses := newHttpSession(self, req, res)

	var msg interface{}
	var err error
	var fileHandled bool

	// 处理消息及页面下发
	// 通过事件系统分发请求，上层应用可以处理消息或返回响应
	self.ProcEvent(&cellnet.RecvMsgEvent{Ses: ses, Msg: msg})

	// 检查是否有错误
	if ses.err != nil {
		err = ses.err
		goto OnError
	}

	// 如果已经响应，直接返回
	if ses.respond {
		return
	}

	// 处理静态文件
	_, err, fileHandled = self.ServeFileWithDir(res, req)

	if err != nil {
		// 或者是普通消息没有 Handled
		log.GetLog().Warnf("#http.recv(%s) '%s' %s | [%d] Not found",
			self.Name(),
			req.Method,
			req.URL.Path,
			http.StatusNotFound)

		// 返回 404 错误
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte(err.Error()))

		return
	}

	// 如果文件已处理，返回
	if fileHandled {
		log.GetLog().Debugf("#http.recv(%s) '%s' %s | [%d] File",
			self.Name(),
			req.Method,
			req.URL.Path,
			http.StatusOK)
		return
	}

	// 未处理的请求
	log.GetLog().Warnf("#http.recv(%s) '%s' %s | Unhandled",
		self.Name(),
		req.Method,
		req.URL.Path)

	return
OnError:
	// 处理错误
	log.GetLog().Errorf("#http.recv(%s) '%s' %s | [%d] %s",
		self.Name(),
		req.Method,
		req.URL.Path,
		http.StatusInternalServerError,
		err.Error())

	// 返回 500 错误
	http.Error(ses.resp, err.Error(), http.StatusInternalServerError)
}

// Stop 停止接受器
// 优雅关闭 HTTP 服务器，等待所有请求处理完成
func (self *httpAcceptor) Stop() {

	if self.sv == nil {
		return
	}

	// 优雅关闭服务器
	if err := self.sv.Shutdown(nil); err != nil {
		log.GetLog().Errorf("#http.stop failed(%s) %v", self.Name(), err.Error())
	}
}

// TypeName 返回接受器的类型名称
// 用于标识和日志记录
func (self *httpAcceptor) TypeName() string {
	return "http.Acceptor"
}

// init 包初始化函数
// 自动注册 HTTP 接受器的创建函数
// 当调用 cellnet.NewPeer("http.Acceptor", ...) 时会使用此函数创建实例
func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &httpAcceptor{}

		return p
	})
}
