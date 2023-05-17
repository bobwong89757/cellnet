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

// httpConnector
// @Description: http连接器
type httpConnector struct {
	peer.CorePeerProperty
	peer.CoreProcBundle
	peer.CoreContextSet
}

func (self *httpConnector) Start() cellnet.Peer {

	return self
}

func (self *httpConnector) Stop() {

}

func getCodec(codecName string) cellnet.Codec {

	if codecName == "" {
		codecName = "httpjson"
	}

	return codec.MustGetCodec(codecName)
}

func getTypeName(msg interface{}) string {
	if msg == nil {
		return ""
	}

	return reflect.TypeOf(msg).Elem().Name()
}

func (self *httpConnector) Request(method, path string, param *cellnet.HTTPRequest) error {

	// 将消息编码为字节数组
	reqCodec := getCodec(param.REQCodecName)
	data, err := reqCodec.Encode(param.REQMsg, nil)

	log.GetLog().Debug("#http.send(%s) '%s' %s | Message(%s) %s",
		self.Name(),
		method,
		path,
		getTypeName(param.REQMsg),
		cellnet.MessageToString(param.REQMsg))

	url := fmt.Sprintf("http://%s%s", self.Address(), path)

	req, err := http.NewRequest(method, url, data.(io.Reader))

	if err != nil {
		return err
	}

	mimeType := reqCodec.(interface {
		MimeType() string
	}).MimeType()

	req.Header.Set("Content-Type", mimeType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = getCodec(param.ACKCodecName).Decode(resp.Body, param.ACKMsg)

	log.GetLog().Debug("#http.recv(%s) '%s' %s | [%d] Message(%s) %s",
		self.Name(),
		resp.Request.Method,
		path,
		resp.StatusCode,
		getTypeName(param.ACKMsg),
		cellnet.MessageToString(param.ACKMsg))

	return err
}

func (self *httpConnector) TypeName() string {
	return "http.Connector"
}

func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &httpConnector{}

		return p
	})
}
