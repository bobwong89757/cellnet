package util

import (
	"bufio"
	"io"
	"net"
	"os"
)

// WriteFull
//
//	@Description: 完整发送所有封包
//	@param writer
//	@param buf
//	@return error
func WriteFull(writer io.Writer, buf []byte) error {

	total := len(buf)

	for pos := 0; pos < total; {

		n, err := writer.Write(buf[pos:])

		if err != nil {
			return err
		}

		pos += n
	}

	return nil

}

// ReadFileLines
//
//	@Description: 读取文本文件的所有行
//	@param filename
//	@param callback
//	@return error
func ReadFileLines(filename string, callback func(line string) bool) error {

	f, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	reader := bufio.NewScanner(f)

	reader.Split(bufio.ScanLines)
	for reader.Scan() {

		if !callback(reader.Text()) {
			break
		}
	}

	return nil
}

// FileExists
//
//	@Description: 检查文件是否存在
//	@param name
//	@return bool
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// FileSize
//
//	@Description: 获取文件大小
//	@param name
//	@return int64
func FileSize(name string) int64 {
	if info, err := os.Stat(name); err == nil {
		return info.Size()
	}

	return 0
}

// IsEOFOrNetReadError
//
//	@Description: 判断网络错误
//	@param err
//	@return bool
func IsEOFOrNetReadError(err error) bool {
	if err == io.EOF {
		return true
	}
	ne, ok := err.(*net.OpError)
	return ok && ne.Op == "read"
}
