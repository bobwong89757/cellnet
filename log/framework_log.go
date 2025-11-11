package log

import (
	"github.com/rs/zerolog"
)

var mLog *zerolog.Logger

func SetLog(v *zerolog.Logger) {
	mLog = v
}

func GetLog() *zerolog.Logger {
	return mLog
}
