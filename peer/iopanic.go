package peer

// CoreCaptureIOPanic 提供 IO 层异常捕获控制的核心实现
// 用于控制是否捕获 IO 操作中的 panic，避免程序崩溃
// 所有 Peer 实现都可以嵌入此结构体来获得异常捕获功能
type CoreCaptureIOPanic struct {
	// captureIOPanic 是否启用 IO 层异常捕获
	// true 表示启用异常捕获，false 表示禁用
	captureIOPanic bool
}

// EnableCaptureIOPanic 启用或禁用 IO 层异常捕获
// v: true 表示启用异常捕获，false 表示禁用
// 启用后，IO 操作中的 panic 会被捕获，避免程序崩溃
// 在生产环境的对外端口应该启用此设置
func (self *CoreCaptureIOPanic) EnableCaptureIOPanic(v bool) {
	self.captureIOPanic = v
}

// CaptureIOPanic 获取当前 IO 层异常捕获设置
// 返回 true 表示已启用异常捕获，false 表示未启用
func (self *CoreCaptureIOPanic) CaptureIOPanic() bool {
	return self.captureIOPanic
}
