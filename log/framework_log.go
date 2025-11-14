package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// mLog 全局的日志记录器
// 可以通过 SetLog 设置自定义的日志记录器
var mLog *zap.SugaredLogger

// fmtCore 是一个使用 fmt 打印的 zap core
// 当没有设置自定义日志记录器时，使用此 core 进行日志输出
type fmtCore struct{}

// Enabled 检查日志级别是否启用
// 所有级别都启用
func (c *fmtCore) Enabled(level zapcore.Level) bool {
	return true
}

// With 添加字段到日志记录器
// 此实现不添加字段，直接返回自身
func (c *fmtCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

// Check 检查日志条目并添加到检查条目
func (c *fmtCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, c)
}

// Write 写入日志条目
// 使用 fmt.Printf 打印日志，格式为 [LEVEL] message
func (c *fmtCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	level := ent.Level.String()
	msg := ent.Message

	// 对于 SugaredLogger，格式化字符串已经包含在 ent.Message 中
	// 使用 fmt 打印，格式: [LEVEL] message
	fmt.Printf("[%s] %s\n", level, msg)
	return nil
}

// Sync 同步日志输出
// 此实现不执行任何操作
func (c *fmtCore) Sync() error {
	return nil
}

// defaultLog 默认的日志记录器
// 使用 fmtCore 进行日志输出
var defaultLog *zap.SugaredLogger

// init 初始化默认日志记录器
// 创建一个使用 fmt 的默认 logger
func init() {
	// 创建一个使用 fmt 的默认 logger
	core := &fmtCore{}
	logger := zap.New(core, zap.AddCallerSkip(1))
	defaultLog = logger.Sugar()
}

// SetLog 设置全局日志记录器
// v: 要设置的日志记录器，如果为 nil 则使用默认记录器
func SetLog(v *zap.SugaredLogger) {
	mLog = v
}

// GetLog 获取全局日志记录器
// 如果已设置自定义记录器，返回自定义记录器；否则返回默认记录器
func GetLog() *zap.SugaredLogger {
	if mLog != nil {
		return mLog
	}
	return defaultLog
}

// Debugf 打印调试日志
// template: 日志模板字符串
// args: 日志参数
// 如果 log 未初始化则使用 fmt 打印
func Debugf(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Debugf(template, args...)
	} else {
		fmt.Printf("[DEBUG] "+template+"\n", args...)
	}
}

// Infof 打印信息日志
// template: 日志模板字符串
// args: 日志参数
// 如果 log 未初始化则使用 fmt 打印
func Infof(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Infof(template, args...)
	} else {
		fmt.Printf("[INFO] "+template+"\n", args...)
	}
}

// Warnf 打印警告日志
// template: 日志模板字符串
// args: 日志参数
// 如果 log 未初始化则使用 fmt 打印
func Warnf(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Warnf(template, args...)
	} else {
		fmt.Printf("[WARN] "+template+"\n", args...)
	}
}

// Errorf 打印错误日志
// template: 日志模板字符串
// args: 日志参数
// 如果 log 未初始化则使用 fmt 打印
func Errorf(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Errorf(template, args...)
	} else {
		fmt.Printf("[ERROR] "+template+"\n", args...)
	}
}

// Debugln 打印调试日志（换行）
// args: 日志参数
// 如果 log 未初始化则使用 fmt 打印
func Debugln(args ...interface{}) {
	if mLog != nil {
		mLog.Debugln(args...)
	} else {
		fmt.Println(append([]interface{}{"[DEBUG]"}, args...)...)
	}
}
