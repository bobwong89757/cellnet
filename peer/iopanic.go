package peer

type CoreCaptureIOPanic struct {
	captureIOPanic bool
}

// EnableCaptureIOPanic
//
//	@Description: 捕获IO错误
//	@receiver self
//	@param v
func (self *CoreCaptureIOPanic) EnableCaptureIOPanic(v bool) {
	self.captureIOPanic = v
}

// CaptureIOPanic
//
//	@Description: 获取是否允许捕获IO错误
//	@receiver self
//	@return bool
func (self *CoreCaptureIOPanic) CaptureIOPanic() bool {
	return self.captureIOPanic
}
