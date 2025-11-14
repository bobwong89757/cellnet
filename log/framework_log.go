package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var mLog *zap.SugaredLogger

// fmtCore 是一个使用fmt打印的zap core
type fmtCore struct{}

func (c *fmtCore) Enabled(level zapcore.Level) bool {
	return true
}

func (c *fmtCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *fmtCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, c)
}

func (c *fmtCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	level := ent.Level.String()
	msg := ent.Message

	// 对于SugaredLogger，格式化字符串已经包含在ent.Message中
	// 使用fmt打印，格式: [LEVEL] message
	fmt.Printf("[%s] %s\n", level, msg)
	return nil
}

func (c *fmtCore) Sync() error {
	return nil
}

var defaultLog *zap.SugaredLogger

func init() {
	// 创建一个使用fmt的默认logger
	core := &fmtCore{}
	logger := zap.New(core, zap.AddCallerSkip(1))
	defaultLog = logger.Sugar()
}

func SetLog(v *zap.SugaredLogger) {
	mLog = v
}

func GetLog() *zap.SugaredLogger {
	if mLog != nil {
		return mLog
	}
	return defaultLog
}

// Debugf 打印调试日志，如果log未初始化则使用fmt打印
func Debugf(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Debugf(template, args...)
	} else {
		fmt.Printf("[DEBUG] "+template+"\n", args...)
	}
}

// Infof 打印信息日志，如果log未初始化则使用fmt打印
func Infof(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Infof(template, args...)
	} else {
		fmt.Printf("[INFO] "+template+"\n", args...)
	}
}

// Warnf 打印警告日志，如果log未初始化则使用fmt打印
func Warnf(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Warnf(template, args...)
	} else {
		fmt.Printf("[WARN] "+template+"\n", args...)
	}
}

// Errorf 打印错误日志，如果log未初始化则使用fmt打印
func Errorf(template string, args ...interface{}) {
	if mLog != nil {
		mLog.Errorf(template, args...)
	} else {
		fmt.Printf("[ERROR] "+template+"\n", args...)
	}
}

// Debugln 打印调试日志（换行），如果log未初始化则使用fmt打印
func Debugln(args ...interface{}) {
	if mLog != nil {
		mLog.Debugln(args...)
	} else {
		fmt.Println(append([]interface{}{"[DEBUG]"}, args...)...)
	}
}
