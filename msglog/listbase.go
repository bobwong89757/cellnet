package msglog

import (
	"errors"
	"github.com/bobwong89757/cellnet"
	"sync"
)

var (
	// whiteListByMsgID 存储白名单消息 ID 的映射表
	// 键为消息 ID（int），值为对应的 MessageMeta
	// 使用 sync.Map 保证并发安全
	whiteListByMsgID sync.Map

	// blackListByMsgID 存储黑名单消息 ID 的映射表
	// 键为消息 ID（int），值为对应的 MessageMeta
	// 使用 sync.Map 保证并发安全
	blackListByMsgID sync.Map

	// currMsgLogMode 当前的消息日志处理模式
	// 使用原子操作或锁保证并发安全
	currMsgLogMode MsgLogMode = MsgLogMode_BlackList

	// currMsgLogModeGuard 保护 currMsgLogMode 的读写锁
	currMsgLogModeGuard sync.RWMutex
)

// MsgLogRule 定义消息日志规则
// 用于控制特定消息的日志显示
type MsgLogRule int

const (
	// MsgLogRule_None 显示所有的消息日志
	// 不应用任何规则，消息正常显示
	MsgLogRule_None MsgLogRule = iota

	// MsgLogRule_BlackList 黑名单内的不显示
	// 添加到黑名单的消息不会显示日志
	MsgLogRule_BlackList

	// MsgLogRule_WhiteList 只显示白名单的日志
	// 只有添加到白名单的消息才会显示日志
	MsgLogRule_WhiteList
)

// MsgLogMode 定义消息日志模式
// 用于控制全局的消息日志显示策略
type MsgLogMode int

const (
	// MsgLogMode_ShowAll 显示所有的消息日志
	// 所有消息都会显示日志，不受白名单和黑名单影响
	MsgLogMode_ShowAll MsgLogMode = iota

	// MsgLogMode_Mute 禁用所有的消息日志
	// 所有消息都不会显示日志
	MsgLogMode_Mute

	// MsgLogMode_BlackList 黑名单内的不显示
	// 黑名单中的消息不显示日志，其他消息正常显示
	MsgLogMode_BlackList

	// MsgLogMode_WhiteList 只显示白名单的日志
	// 只有白名单中的消息显示日志，其他消息不显示
	MsgLogMode_WhiteList
)

// SetCurrMsgLogMode 设置当前的消息日志处理模式
// mode: 要设置的日志模式
// 用于控制全局的消息日志显示策略
func SetCurrMsgLogMode(mode MsgLogMode) {
	currMsgLogModeGuard.Lock()
	currMsgLogMode = mode
	currMsgLogModeGuard.Unlock()
}

// GetCurrMsgLogMode 获取当前的消息日志处理模式
// 返回当前的日志模式
func GetCurrMsgLogMode() MsgLogMode {
	currMsgLogModeGuard.RLock()
	defer currMsgLogModeGuard.RUnlock()
	return currMsgLogMode
}

// SetMsgLogRule 指定某个消息的处理规则
// name: 消息的完整名称，格式为 "packageName.MsgName"
// rule: 要设置的日志规则
// 返回错误信息，如果成功则返回 nil
// 如果消息未注册，返回错误
func SetMsgLogRule(name string, rule MsgLogRule) error {
	// 根据消息名称获取消息元信息
	meta := cellnet.MessageMetaByFullName(name)
	if meta == nil {
		return errors.New("msg not found")
	}

	// 根据规则添加到对应的列表
	switch rule {
	case MsgLogRule_BlackList:
		// 添加到黑名单
		blackListByMsgID.Store(int(meta.ID), meta)
	case MsgLogRule_WhiteList:
		// 添加到白名单
		whiteListByMsgID.Store(int(meta.ID), meta)
	case MsgLogRule_None:
		// 从黑名单和白名单中移除
		blackListByMsgID.Delete(int(meta.ID))
		whiteListByMsgID.Delete(int(meta.ID))
	}

	return nil
}

// IsMsgLogValid 检查能否显示消息日志
// msgid: 消息的唯一标识符
// 返回 true 表示可以显示日志，false 表示不显示
// 根据当前的日志模式和消息是否在白名单/黑名单中判断
func IsMsgLogValid(msgid int) bool {
	switch GetCurrMsgLogMode() {
	case MsgLogMode_BlackList:
		// 黑名单模式：黑名单里的不显示
		if _, ok := blackListByMsgID.Load(msgid); ok {
			return false
		} else {
			return true
		}
	case MsgLogMode_WhiteList:
		// 白名单模式：只有在白名单里才显示
		if _, ok := whiteListByMsgID.Load(msgid); ok {
			return true
		} else {
			return false
		}
	case MsgLogMode_Mute:
		// 静默模式：所有消息都不显示
		return false
	}

	// MsgLogMode_ShowAll：显示所有消息
	return true
}

// VisitMsgLogRule 遍历消息规则
// mode: 要遍历的日志模式（黑名单或白名单）
// callback: 遍历回调函数，参数为消息元信息
// 如果回调返回 false，则停止遍历
func VisitMsgLogRule(mode MsgLogMode, callback func(*cellnet.MessageMeta) bool) {
	switch mode {
	case MsgLogMode_BlackList:
		// 遍历黑名单
		blackListByMsgID.Range(func(key, value interface{}) bool {
			meta := value.(*cellnet.MessageMeta)
			return callback(meta)
		})
	case MsgLogMode_WhiteList:
		// 遍历白名单
		whiteListByMsgID.Range(func(key, value interface{}) bool {
			meta := value.(*cellnet.MessageMeta)
			return callback(meta)
		})
	}
}
