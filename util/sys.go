package util

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// StackToString 将调用栈转换为字符串
// count: 要打印的栈层数，一般 3~5 层可以覆盖逻辑及封装代码范围
// 返回格式化的调用栈字符串，格式为 "file:line -> file:line -> ..."
// 用于调试和错误追踪
func StackToString(count int) string {
	// 从第 2 层开始（跳过 runtime.Caller 和 StackToString 自身）
	const startStack = 2

	var sb strings.Builder

	// 遍历指定层数的调用栈
	for i := startStack; i < startStack+count; i++ {
		// 获取调用栈信息
		_, file, line, ok := runtime.Caller(i)

		var str string

		if ok {
			// 格式化文件名和行号
			str = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		} else {
			// 无法获取调用栈信息
			str = "??"
		}

		// 跳过 "??" 标记
		if str != "??" {
			// 如果不是第一层，添加分隔符
			if i > startStack {
				sb.WriteString(" -> ")
			}

			sb.WriteString(str)
		} else {
			// 遇到 "??" 则停止
			break
		}
	}

	return sb.String()
}
