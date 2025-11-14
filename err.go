package cellnet

import "fmt"

// Error 是 cellnet 框架的自定义错误类型
// 支持携带上下文信息，便于错误追踪和调试
type Error struct {
	// s 错误消息字符串
	s string

	// context 错误的上下文信息，可以是任意类型
	// 用于提供额外的错误相关信息
	context interface{}
}

// Error 实现 error 接口，返回错误字符串
// 如果有关联的上下文信息，会在错误消息中包含
func (self *Error) Error() string {
	// 没有上下文信息，直接返回错误消息
	if self.context == nil {
		return self.s
	}

	// 包含上下文信息的错误消息
	return fmt.Sprintf("%s, '%v'", self.s, self.context)
}

// NewError 创建一个新的错误
// s: 错误消息字符串
// 返回一个 Error 实例
func NewError(s string) error {
	return &Error{s: s}
}

// NewErrorContext 创建一个带上下文信息的错误
// s: 错误消息字符串
// context: 错误的上下文信息，可以是任意类型
// 返回一个包含上下文信息的 Error 实例
func NewErrorContext(s string, context interface{}) error {
	return &Error{s: s, context: context}
}
