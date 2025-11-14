package util

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
)

// StringHash 将字符串转换为 16 位整数哈希值
// s: 要哈希的字符串
// 返回 16 位无符号整数哈希值
// 使用简单的哈希算法，适合用于消息 ID 生成等场景
func StringHash(s string) (hash uint16) {
	// 遍历字符串的每个字符
	for _, c := range s {
		ch := uint16(c)

		// 哈希算法：hash = hash + (hash << 5) + ch + (ch << 7)
		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}

	return
}

// BytesMD5 计算字节数组的 MD5 哈希值
// data: 要计算哈希的字节数组
// 返回 MD5 哈希值的十六进制字符串表示
func BytesMD5(data []byte) string {
	m := md5.New()
	m.Write(data)
	return hex.EncodeToString(m.Sum(nil))
}

// StringMD5 计算字符串的 MD5 哈希值
// str: 要计算哈希的字符串
// 返回 MD5 哈希值的十六进制字符串表示
func StringMD5(str string) string {
	return BytesMD5([]byte(str))
}

// CompressBytes 使用 zlib 压缩字节数组
// data: 要压缩的字节数组
// 返回压缩后的字节数组和错误信息
// 如果压缩失败，返回错误
func CompressBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	// 创建 zlib 压缩写入器
	writer := zlib.NewWriter(&buf)

	// 写入数据
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	// 关闭写入器，完成压缩
	writer.Close()

	return buf.Bytes(), nil
}

// DecompressBytes 使用 zlib 解压字节数组
// data: 要解压的字节数组
// 返回解压后的字节数组和错误信息
// 如果解压失败，返回错误
func DecompressBytes(data []byte) ([]byte, error) {
	// 创建 zlib 解压读取器
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	// 读取所有解压后的数据
	return ioutil.ReadAll(reader)
}
