package msglog

import (
	"errors"
	"github.com/bobwong89757/cellnet"
	"sync"
)

var (
	whiteListByMsgID    sync.Map
	blackListByMsgID    sync.Map
	currMsgLogMode      = MsgLogMode_BlackList
	currMsgLogModeGuard sync.RWMutex
)

type MsgLogRule int

const (
	// 显示所有的消息日志
	MsgLogRule_None MsgLogRule = iota

	// 黑名单内的不显示
	MsgLogRule_BlackList

	// 只显示白名单的日志
	MsgLogRule_WhiteList
)

type MsgLogMode int

const (
	// 显示所有的消息日志
	MsgLogMode_ShowAll MsgLogMode = iota

	// 禁用所有的消息日志
	MsgLogMode_Mute

	// 黑名单内的不显示
	MsgLogMode_BlackList

	// 只显示白名单的日志
	MsgLogMode_WhiteList
)

// SetCurrMsgLogMode
//
//	@Description: 设置当前的消息日志处理模式
//	@param mode
func SetCurrMsgLogMode(mode MsgLogMode) {
	currMsgLogModeGuard.Lock()
	currMsgLogMode = mode
	currMsgLogModeGuard.Unlock()
}

// GetCurrMsgLogMode
//
//	@Description: 获取当前的消息日志处理模式
//	@return MsgLogMode
func GetCurrMsgLogMode() MsgLogMode {
	currMsgLogModeGuard.RLock()
	defer currMsgLogModeGuard.RUnlock()
	return currMsgLogMode
}

// SetMsgLogRule
//
//	@Description: 指定某个消息的处理规则, 消息格式: packageName.MsgName
//	@param name
//	@param rule
//	@return error
func SetMsgLogRule(name string, rule MsgLogRule) error {

	meta := cellnet.MessageMetaByFullName(name)
	if meta == nil {
		return errors.New("msg not found")
	}

	switch rule {
	case MsgLogRule_BlackList:
		blackListByMsgID.Store(int(meta.ID), meta)
	case MsgLogRule_WhiteList:
		whiteListByMsgID.Store(int(meta.ID), meta)
	case MsgLogRule_None:
		blackListByMsgID.Delete(int(meta.ID))
		whiteListByMsgID.Delete(int(meta.ID))
	}

	return nil
}

// IsMsgLogValid
//
//	@Description: 能否显示消息日志
//	@param msgid
//	@return bool
func IsMsgLogValid(msgid int) bool {
	switch GetCurrMsgLogMode() {
	case MsgLogMode_BlackList: // 黑名单里不显示
		if _, ok := blackListByMsgID.Load(msgid); ok {
			return false
		} else {
			return true
		}
	case MsgLogMode_WhiteList: // 只有在白名单里才显示
		if _, ok := whiteListByMsgID.Load(msgid); ok {
			return true
		} else {
			return false
		}
	case MsgLogMode_Mute:
		return false
	}

	// MsgLogMode_ShowAll
	return true
}

// VisitMsgLogRule
//
//	@Description: 遍历消息规则
//	@param mode
//	@param callback
func VisitMsgLogRule(mode MsgLogMode, callback func(*cellnet.MessageMeta) bool) {

	switch mode {
	case MsgLogMode_BlackList:
		blackListByMsgID.Range(func(key, value interface{}) bool {
			meta := value.(*cellnet.MessageMeta)

			return callback(meta)
		})
	case MsgLogMode_WhiteList:
		whiteListByMsgID.Range(func(key, value interface{}) bool {
			meta := value.(*cellnet.MessageMeta)

			return callback(meta)
		})
	}

}
