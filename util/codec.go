package util

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
)

// StringHash
//
//	@Description: 字符串转为16位整形哈希
//	@param s
//	@return hash
func StringHash(s string) (hash uint16) {

	for _, c := range s {

		ch := uint16(c)

		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}

	return
}

// BytesMD5
//
//	@Description: 字节计算MD5
//	@param data
//	@return string
func BytesMD5(data []byte) string {
	m := md5.New()
	m.Write(data)
	return hex.EncodeToString(m.Sum(nil))
}

// StringMD5
//
//	@Description: 字符串计算MD5
//	@param str
//	@return string
func StringMD5(str string) string {
	return BytesMD5([]byte(str))
}

// CompressBytes
//
//	@Description: 压缩字节
//	@param data
//	@return []byte
//	@return error
func CompressBytes(data []byte) ([]byte, error) {

	var buf bytes.Buffer

	writer := zlib.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()

	return buf.Bytes(), nil
}

// DecompressBytes
//
//	@Description: 解压字节
//	@param data
//	@return []byte
//	@return error
func DecompressBytes(data []byte) ([]byte, error) {

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	return ioutil.ReadAll(reader)
}
