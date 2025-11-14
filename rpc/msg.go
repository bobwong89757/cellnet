package rpc

// GetMsgID 获取消息 ID（实现 RemoteCallMsg 接口）
// 返回消息 ID
func (self *RemoteCallREQ) GetMsgID() uint16 { return uint16(self.MsgID) }

// GetMsgData 获取消息数据（实现 RemoteCallMsg 接口）
// 返回消息数据字节数组
func (self *RemoteCallREQ) GetMsgData() []byte { return self.Data }

// GetCallID 获取调用 ID（实现 RemoteCallMsg 接口）
// 返回调用 ID，用于关联请求和响应
func (self *RemoteCallREQ) GetCallID() int64 { return self.CallID }

// GetMsgID 获取消息 ID（实现 RemoteCallMsg 接口）
// 返回消息 ID
func (self *RemoteCallACK) GetMsgID() uint16 { return uint16(self.MsgID) }

// GetMsgData 获取消息数据（实现 RemoteCallMsg 接口）
// 返回消息数据字节数组
func (self *RemoteCallACK) GetMsgData() []byte { return self.Data }

// GetCallID 获取调用 ID（实现 RemoteCallMsg 接口）
// 返回调用 ID，用于关联请求和响应
func (self *RemoteCallACK) GetCallID() int64 { return self.CallID }
