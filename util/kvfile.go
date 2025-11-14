package util

import (
	"errors"
	"strings"
)

// ReadKVFile 读取以等号分割的配置文件
// filename: 要读取的文件名
// callback: 每行的回调函数，参数为键和值
//   如果回调返回 false，则停止读取
// 返回读取错误，如果成功则返回 nil
//
// 文件格式：
//   - 以 "#" 开头的行被视为注释，会被忽略
//   - 每行格式为 "key=value" 或 "value"（无键）
//   - 等号前后可以有空格，会被自动去除
func ReadKVFile(filename string, callback func(k, v string) bool) (ret error) {
	readErr := ReadFileLines(filename, func(line string) bool {
		// 去除首尾空格
		line = strings.TrimSpace(line)

		// 注释行，跳过
		if strings.HasPrefix(line, "#") {
			return true
		}

		// 等号切分键值对
		pairs := strings.Split(line, "=")

		switch len(pairs) {
		case 1:
			// 只有值，没有键
			value := strings.TrimSpace(pairs[0])

			if value == "" {
				return true
			}

			return callback("", value)
		case 2:
			// 键值对
			key := strings.TrimSpace(pairs[0])
			value := strings.TrimSpace(pairs[1])

			if key == "" {
				return true
			}

			return callback(key, value)
		default:
			// 多个等号，格式错误
			ret = errors.New("Require '=' splite key and value")
			return false
		}
	})

	if readErr != nil {
		return readErr
	}

	return
}

// KVPair 表示一个键值对
type KVPair struct {
	// Key 键
	Key string

	// Value 值
	Value string
}

// ReadKVFileValues 读取以等号分割的配置文件并返回所有键值对
// filename: 要读取的文件名
// 返回所有键值对的切片和错误信息
// 如果读取失败，返回错误
func ReadKVFileValues(filename string) (ret []KVPair, err error) {
	err = ReadKVFile(filename, func(k, v string) bool {
		// 将键值对添加到结果列表
		ret = append(ret, KVPair{k, v})
		return true
	})

	return
}
