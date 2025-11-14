package httpjson

import (
	"bytes"
	"encoding/json"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"io"
	"io/ioutil"
	"net/http"
)

// httpjsonCodec HTTP JSON 编码器实现
// 实现 cellnet.Codec 接口，用于 HTTP 请求和响应的 JSON 编码/解码
type httpjsonCodec struct {
}

// Name 返回编码器的名称
// 返回 "httpjson"
func (self *httpjsonCodec) Name() string {
	return "httpjson"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/json"
func (self *httpjsonCodec) MimeType() string {
	return "application/json"
}

// Encode 将结构体编码为 JSON 的字节数组
// msgObj: 要编码的消息对象
// ctx: 上下文集合（未使用）
// 返回 io.Reader 接口，用于 HTTP 请求体
func (self *httpjsonCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {

	// 将对象编码为 JSON 字节数组
	bdata, err := json.Marshal(msgObj)
	if err != nil {
		return nil, err
	}

	// 返回 Reader 接口，用于 HTTP 请求体
	return bytes.NewReader(bdata), nil
}

// Decode 将 JSON 的字节数组解码为结构体
// data: 数据源，可以是 *http.Request 或 io.Reader
// msgObj: 要解码到的消息对象（指针）
// 从 HTTP 请求体或 Reader 中读取 JSON 数据并解码
func (self *httpjsonCodec) Decode(data interface{}, msgObj interface{}) error {

	var reader io.Reader
	// 根据数据类型获取 Reader
	switch v := data.(type) {
	case *http.Request:
		// 从 HTTP 请求体中读取
		reader = v.Body
	case io.Reader:
		// 直接使用 Reader
		reader = v
	}

	// 读取所有数据
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	// 解码 JSON 数据
	return json.Unmarshal(body, msgObj)
}

// init 包初始化函数
// 自动注册 HTTP JSON 编码器
func init() {

	// 注册编码器
	codec.RegisterCodec(new(httpjsonCodec))
}
