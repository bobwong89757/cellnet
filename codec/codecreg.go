package codec

import (
	"fmt"

	"github.com/bobwong89757/cellnet"
)

// registedCodecs 存储所有已注册的编码器
// 编码器在 init() 函数中通过 RegisterCodec 注册
var registedCodecs []cellnet.Codec

// RegisterCodec 注册一个编码器
// c: 要注册的编码器，必须实现 cellnet.Codec 接口
// 如果编码器名称已存在，会触发 panic
// 通常在编码器包的 init() 函数中调用此函数进行注册
func RegisterCodec(c cellnet.Codec) {
	// 检查编码器名称是否已存在
	if GetCodec(c.Name()) != nil {
		panic("duplicate codec: " + c.Name())
	}

	// 添加到注册列表
	registedCodecs = append(registedCodecs, c)
}

// GetCodec 根据名称获取编码器
// name: 编码器的名称，如 "protobuf"、"json"、"binary" 等
// 返回对应的编码器，如果不存在返回 nil
func GetCodec(name string) cellnet.Codec {
	// 遍历已注册的编码器列表
	for _, c := range registedCodecs {
		if c.Name() == name {
			return c
		}
	}

	return nil
}

// getPackageByCodecName 根据编码器名称返回对应的包路径
// name: 编码器的名称
// 返回编码器所在的包路径，用于错误提示
// 这是 cellnet 自带的编码器对应的包路径
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
	case "sproto":
		return "github.com/bobwong89757/cellnet/codec/sproto"
	default:
		return "package/to/your/codec"
	}
}

// MustGetCodec 获取编码器，如果不存在则 panic
// name: 编码器的名称
// 返回对应的编码器
// 如果编码器不存在，会触发 panic 并提供导入提示信息
// 用于在编码器必须存在的情况下使用，避免 nil 检查
func MustGetCodec(name string) cellnet.Codec {
	codec := GetCodec(name)

	if codec == nil {
		// 编码器不存在，提供友好的错误提示
		panic(fmt.Sprintf("codec not found '%s'\ntry to add code below:\nimport (\n  _ \"%s\"\n)\n\n",
			name,
			getPackageByCodecName(name)))
	}

	return codec
}
