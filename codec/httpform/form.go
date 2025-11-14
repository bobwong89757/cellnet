package httpform

import (
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/codec"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// httpFormCodec HTTP Form 编码器实现
// 实现 cellnet.Codec 接口，用于 HTTP 表单数据的编码/解码
// 支持将结构体编码为 URL-encoded 表单数据，以及从表单数据解码为结构体
type httpFormCodec struct {
}

// defaultMemory 默认多部分表单内存大小
// 用于解析 multipart/form-data 表单
const defaultMemory = 32 * 1024 * 1024

// Name 返回编码器的名称
// 返回 "httpform"
func (self *httpFormCodec) Name() string {
	return "httpform"
}

// MimeType 返回编码器的 MIME 类型
// 返回 "application/x-www-form-urlencoded"
func (self *httpFormCodec) MimeType() string {
	return "application/x-www-form-urlencoded"
}

// anyToString 将任意类型转换为字符串
// any: 要转换的值
// 支持 string、bool、int、int32、int64、float32、float64 类型
// 返回转换后的字符串
func anyToString(any interface{}) string {

	switch v := any.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(int64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		panic("Unknown type to convert to string")
	}
}

// structToUrlValues 将结构体转换为 URL 值
// obj: 要转换的结构体对象
// 遍历结构体的所有字段，将字段名和值添加到 URL 值中
// 返回 URL 值对象
func structToUrlValues(obj interface{}) url.Values {
	objValue := reflect.Indirect(reflect.ValueOf(obj))

	objType := objValue.Type()

	var formValues = url.Values{}
	// 遍历所有字段
	for i := 0; i < objValue.NumField(); i++ {

		fieldType := objType.Field(i)

		fieldValue := objValue.Field(i)

		// 将字段名和值添加到表单值中
		formValues.Add(fieldType.Name, anyToString(fieldValue.Interface()))
	}

	return formValues
}

// Encode 将结构体编码为 URL-encoded 表单数据
// msgObj: 要编码的消息对象
// ctx: 上下文集合（未使用）
// 将结构体转换为 URL 值，然后编码为字符串
// 返回 io.Reader 接口，用于 HTTP 请求体
func (self *httpFormCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {

	return strings.NewReader(structToUrlValues(msgObj).Encode()), err
}

// Decode 将 URL-encoded 表单数据解码为结构体
// data: 数据源，必须是 *http.Request
// msgObj: 要解码到的消息对象（指针）
// 解析 HTTP 请求的表单数据，并映射到结构体字段
// 支持普通表单和 multipart 表单
func (self *httpFormCodec) Decode(data interface{}, msgObj interface{}) error {

	req := data.(*http.Request)

	// 解析普通表单
	if err := req.ParseForm(); err != nil {
		return err
	}
	// 解析 multipart 表单
	req.ParseMultipartForm(defaultMemory)
	// 将表单数据映射到结构体
	if err := mapForm(msgObj, req.Form); err != nil {
		return err
	}

	return nil
}

// init 包初始化函数
// 自动注册 HTTP Form 编码器
func init() {

	codec.RegisterCodec(new(httpFormCodec))
}
