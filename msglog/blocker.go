package msglog

import (
	"github.com/bobwong89757/cellnet"
)

// IsBlockedMessageByID
//
//	@Description: 当前的某个消息ID是否被屏蔽
//	@param msgid
//	@return bool
func IsBlockedMessageByID(msgid int) bool {

	_, ok := blackListByMsgID.Load(msgid)

	return ok
}

// BlockMessageLog
//
//	@Description: 按指定规则(或消息名)屏蔽消息日志, 需要使用完整消息名 例如 proto.MsgName
//	@param nameRule
//	@return err
//	@return matchCount
func BlockMessageLog(nameRule string) (err error, matchCount int) {

	err = cellnet.MessageMetaVisit(nameRule, func(meta *cellnet.MessageMeta) bool {

		blackListByMsgID.Store(int(meta.ID), meta)
		matchCount++

		return true
	})

	return
}

// RemoveBlockedMessage
//
//	@Description: 移除被屏蔽的消息
//	@param nameRule
//	@return err
//	@return matchCount
func RemoveBlockedMessage(nameRule string) (err error, matchCount int) {

	err = cellnet.MessageMetaVisit(nameRule, func(meta *cellnet.MessageMeta) bool {

		blackListByMsgID.Delete(int(meta.ID))
		matchCount++

		return true
	})

	return
}

// VisitBlockedMessage
//
//	@Description: 遍历被屏蔽的消息
//	@param callback
func VisitBlockedMessage(callback func(*cellnet.MessageMeta) bool) {

	blackListByMsgID.Range(func(key, value interface{}) bool {
		meta := value.(*cellnet.MessageMeta)

		return callback(meta)
	})

}
