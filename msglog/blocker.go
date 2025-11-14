package msglog

import (
	"github.com/bobwong89757/cellnet"
)

// IsBlockedMessageByID 检查某个消息 ID 是否被屏蔽
// msgid: 消息的唯一标识符
// 返回 true 表示消息已被屏蔽，false 表示未被屏蔽
func IsBlockedMessageByID(msgid int) bool {
	_, ok := blackListByMsgID.Load(msgid)
	return ok
}

// BlockMessageLog 按指定规则屏蔽消息日志
// nameRule: 消息名称的匹配规则，支持正则表达式，需要使用完整消息名，例如 "proto.MsgName"
// 返回错误信息和匹配的消息数量
// 匹配的消息会被添加到黑名单，不再显示日志
func BlockMessageLog(nameRule string) (err error, matchCount int) {
	// 遍历匹配的消息
	err = cellnet.MessageMetaVisit(nameRule, func(meta *cellnet.MessageMeta) bool {
		// 添加到黑名单
		blackListByMsgID.Store(int(meta.ID), meta)
		matchCount++

		return true
	})

	return
}

// RemoveBlockedMessage 移除被屏蔽的消息
// nameRule: 消息名称的匹配规则，支持正则表达式
// 返回错误信息和匹配的消息数量
// 匹配的消息会从黑名单中移除，恢复显示日志
func RemoveBlockedMessage(nameRule string) (err error, matchCount int) {
	// 遍历匹配的消息
	err = cellnet.MessageMetaVisit(nameRule, func(meta *cellnet.MessageMeta) bool {
		// 从黑名单中移除
		blackListByMsgID.Delete(int(meta.ID))
		matchCount++

		return true
	})

	return
}

// VisitBlockedMessage 遍历被屏蔽的消息
// callback: 遍历回调函数，参数为消息元信息
// 如果回调返回 false，则停止遍历
func VisitBlockedMessage(callback func(*cellnet.MessageMeta) bool) {
	blackListByMsgID.Range(func(key, value interface{}) bool {
		meta := value.(*cellnet.MessageMeta)
		return callback(meta)
	})
}
