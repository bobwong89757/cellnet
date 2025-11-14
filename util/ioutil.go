package util

import (
	"bufio"
	"io"
	"net"
	"os"
)

// WriteFull 完整发送所有数据
// writer: 数据写入器
// buf: 要写入的字节数组
// 返回写入错误，如果成功则返回 nil
// 此函数会确保所有数据都被写入，即使需要多次 Write 调用
func WriteFull(writer io.Writer, buf []byte) error {
	total := len(buf)

	// 循环写入，直到所有数据都写入完成
	for pos := 0; pos < total; {
		// 写入剩余数据
		n, err := writer.Write(buf[pos:])

		if err != nil {
			return err
		}

		// 更新写入位置
		pos += n
	}

	return nil
}

// ReadFileLines 读取文本文件的所有行
// filename: 要读取的文件名
// callback: 每行的回调函数，参数为行内容
//   如果回调返回 false，则停止读取
// 返回读取错误，如果成功则返回 nil
func ReadFileLines(filename string, callback func(line string) bool) error {
	// 打开文件
	f, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	// 创建扫描器
	reader := bufio.NewScanner(f)

	// 按行分割
	reader.Split(bufio.ScanLines)
	// 逐行读取
	for reader.Scan() {
		// 调用回调函数
		if !callback(reader.Text()) {
			// 回调返回 false，停止读取
			break
		}
	}

	return nil
}

// FileExists 检查文件是否存在
// name: 文件名或路径
// 返回 true 表示文件存在，false 表示不存在
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// FileSize 获取文件大小
// name: 文件名或路径
// 返回文件大小（字节数），如果文件不存在或无法获取，返回 0
func FileSize(name string) int64 {
	if info, err := os.Stat(name); err == nil {
		return info.Size()
	}

	return 0
}

// IsEOFOrNetReadError 判断是否为 EOF 或网络读取错误
// err: 要检查的错误
// 返回 true 表示是 EOF 或网络读取错误，false 表示其他错误
// 用于判断网络连接是否正常关闭或发生读取错误
func IsEOFOrNetReadError(err error) bool {
	// 检查是否为 EOF
	if err == io.EOF {
		return true
	}
	// 检查是否为网络操作错误且操作类型为 "read"
	ne, ok := err.(*net.OpError)
	return ok && ne.Op == "read"
}
