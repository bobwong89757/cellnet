package http

import (
	"errors"
	"fmt"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	_ "github.com/bobwong89757/cellnet/codec/httpjson"
	"github.com/bobwong89757/cellnet/log"
	"io"
	"io/ioutil"
	"net/http"
)

// MessageRespond 消息响应处理器
// 实现 RespondProc 接口，用于返回 JSON 或其他编码格式的消息响应
type MessageRespond struct {
	// StatusCode HTTP 状态码
	// 默认为 200 (StatusOK)
	StatusCode int

	// Msg 要返回的消息对象
	Msg interface{}

	// CodecName 编码器名称
	// 默认为 "httpjson"
	CodecName string
}

// String 返回字符串表示
// 用于日志和调试
func (self *MessageRespond) String() string {
	return fmt.Sprintf("Code: %d Msg: %+v CodeName: %s", self.StatusCode, self.Msg, self.CodecName)
}

// WriteRespond 写入响应（实现 RespondProc 接口）
// ses: HTTP Session
// 将消息编码后写入 HTTP 响应
// 返回处理错误
func (self *MessageRespond) WriteRespond(ses *httpSession) error {
	peerInfo := ses.Peer().(cellnet.PeerProperty)

	// 设置默认编码器
	if self.CodecName == "" {
		self.CodecName = "httpjson"
	}
	// 设置默认状态码
	if self.StatusCode == 0 {
		self.StatusCode = http.StatusOK
	}

	// 获取编码器
	httpCodec := codec.GetCodec(self.CodecName)

	if httpCodec == nil {
		return errors.New("ResponseCodec not found:" + self.CodecName)
	}

	msg := self.Msg

	// 将消息编码为字节数组
	var data interface{}
	data, err := httpCodec.Encode(msg, nil)

	if err != nil {
		return err
	}

	// 设置响应头
	ses.resp.Header().Set("Content-Type", httpCodec.MimeType()+";charset=UTF-8")
	ses.resp.WriteHeader(self.StatusCode)

	// 读取编码后的数据
	bodyData, err := ioutil.ReadAll(data.(io.Reader))
	if err != nil {
		return err
	}

	// 记录发送日志
	log.GetLog().Debugf("#http.send(%s) '%s' %s | [%d] %s",
		peerInfo.Name(),
		ses.req.Method,
		ses.req.URL.Path,
		self.StatusCode,
		string(bodyData))

	// 写入响应体
	ses.resp.Write(bodyData)

	return nil
}
