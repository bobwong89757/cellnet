package codec

import (
	"fmt"
	"github.com/bobwong89757/cellnet"
)

var registedCodecs []cellnet.Codec

// RegisterCodec
//
//	@Description: 注册编码器
//	@param c
func RegisterCodec(c cellnet.Codec) {

	if GetCodec(c.Name()) != nil {
		panic("duplicate codec: " + c.Name())
	}

	registedCodecs = append(registedCodecs, c)
}

// GetCodec
//
//	@Description: 获取编码器
//	@param name
//	@return cellnet.Codec
func GetCodec(name string) cellnet.Codec {

	for _, c := range registedCodecs {
		if c.Name() == name {
			return c
		}
	}

	return nil
}

// getPackageByCodecName
//
//	@Description: cellnet自带的编码对应包
//	@param name
//	@return string
func getPackageByCodecName(name string) string {
	switch name {
	case "binary":
		return "github.com/bobwong89757/cellnet/codec/binary"
	case "gogopb":
		return "github.com/bobwong89757/cellnet/codec/gogopb"
	case "httpjson":
		return "github.com/bobwong89757/cellnet/codec/httpjson"
	case "json":
		return "github.com/bobwong89757/cellnet/codec/json"
	case "protoplus":
		return "github.com/bobwong89757/cellnet/codec/protoplus"
	default:
		return "package/to/your/codec"
	}
}

// MustGetCodec
//
//	@Description: 指定编码器不存在时，报错
//	@param name
//	@return cellnet.Codec
func MustGetCodec(name string) cellnet.Codec {
	codec := GetCodec(name)

	if codec == nil {
		panic(fmt.Sprintf("codec not found '%s'\ntry to add code below:\nimport (\n  _ \"%s\"\n)\n\n",
			name,
			getPackageByCodecName(name)))
	}

	return codec
}
