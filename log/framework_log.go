package log

import (
	"go.uber.org/zap"
)

var mLog *zap.SugaredLogger

func SetLog(v *zap.SugaredLogger) {
	mLog = v
}

func GetLog() *zap.SugaredLogger {
	return mLog
}
