package http

import (
	"fmt"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	"io"
	"net/http"
	"reflect"
)

// httpConnector HTTP 连接器实现
// 用于创建 HTTP 客户端，发送 HTTP 请求
// HTTP 是无状态协议，连接器不需要维护连接状态
type httpConnector struct {
	peer.CorePeerProperty // 核心 Peer 属性（名称、地址、队列等）
	peer.CoreProcBundle   // 消息处理组件（编码器、钩子、回调等）
	peer.CoreContextSet   // 上下文数据存储
}

// Start 开始连接器
// HTTP 连接器不需要启动，直接返回
// 返回自身以支持链式调用
func (self *httpConnector) Start() cellnet.Peer {

	return self
}

// Stop 停止连接器
// HTTP 连接器不需要停止操作（无状态）
func (self *httpConnector) Stop() {

}

// getCodec 获取编码器
// codecName: 编码器名称，如果为空则使用默认的 "httpjson"
// 返回对应的编码器
func getCodec(codecName string) cellnet.Codec {

	if codecName == "" {
		codecName = "httpjson"
	}

	return codec.MustGetCodec(codecName)
}

// getTypeName 获取消息类型名称
// msg: 消息对象
// 返回消息类型的名称（不包含包名）
func getTypeName(msg interface{}) string {
	if msg == nil {
		return ""
	}

	return reflect.TypeOf(msg).Elem().Name()
}

// Request 发送 HTTP 请求
// method: HTTP 方法（GET、POST、PUT、DELETE 等）
// path: 请求路径
// param: 请求参数，包含请求消息、响应消息和编码器名称
// 同步发送请求并等待响应，将响应解码到 param.ACKMsg 中
// 返回请求错误，如果成功则返回 nil
func (self *httpConnector) Request(method, path string, param *cellnet.HTTPRequest) error {

	// 将消息编码为字节数组
	reqCodec := getCodec(param.REQCodecName)
	data, err := reqCodec.Encode(param.REQMsg, nil)

	// 记录发送日志
	log.GetLog().Debugf("#http.send(%s) '%s' %s | Message(%s) %s",
		self.Name(),
		method,
		path,
		getTypeName(param.REQMsg),
		cellnet.MessageToString(param.REQMsg))

	// 构建完整的 URL
	url := fmt.Sprintf("http://%s%s", self.Address(), path)

	// 创建 HTTP 请求
	req, err := http.NewRequest(method, url, data.(io.Reader))

	if err != nil {
		return err
	}

	// 设置 Content-Type 头
	mimeType := reqCodec.(interface {
		MimeType() string
	}).MimeType()

	req.Header.Set("Content-Type", mimeType)

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// 确保响应体被关闭
	defer resp.Body.Close()

	// 解码响应消息
	err = getCodec(param.ACKCodecName).Decode(resp.Body, param.ACKMsg)

	// 记录接收日志
	log.GetLog().Debugf("#http.recv(%s) '%s' %s | [%d] Message(%s) %s",
		self.Name(),
		resp.Request.Method,
		path,
		resp.StatusCode,
		getTypeName(param.ACKMsg),
		cellnet.MessageToString(param.ACKMsg))

	return err
}

// TypeName 返回连接器的类型名称
// 用于标识和日志记录
func (self *httpConnector) TypeName() string {
	return "http.Connector"
}

// init 包初始化函数
// 自动注册 HTTP 连接器的创建函数
// 当调用 cellnet.NewPeer("http.Connector", ...) 时会使用此函数创建实例
func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &httpConnector{}

		return p
	})
}
