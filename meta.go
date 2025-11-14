package cellnet

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// context 用于存储消息元信息的上下文数据
// 支持为消息绑定自定义的键值对信息
type context struct {
	// name 上下文数据的键名
	name string

	// data 上下文数据的值，可以是任意类型
	data interface{}
}

// MessageMeta 存储消息的元信息
// 包括消息类型、编码器、消息ID等，用于消息的注册、查找和编解码
// 每个注册的消息都有一个对应的 MessageMeta 实例
type MessageMeta struct {
	// Codec 消息使用的编码器
	// 用于将消息对象序列化为字节数组，或将字节数组反序列化为消息对象
	Codec Codec

	// Type 消息的类型，使用 reflect.Type 表示
	// 注册时可以使用指针类型，内部会自动转换为非指针类型
	Type reflect.Type

	// ID 消息的唯一标识符，用于二进制协议
	// 在二进制协议中，通过 ID 来识别消息类型
	// ID 必须唯一，不能为 0

	ID int

	// ctxListGuard 保护 ctxList 的读写锁
	// 用于并发安全地访问上下文列表
	ctxListGuard sync.RWMutex

	// ctxList 存储消息的上下文数据列表
	// 支持为消息绑定多个键值对信息
	ctxList []*context
}

// TypeName 返回消息类型的名称（不包含包名）
// 例如，如果消息类型是 "proto.MyMessage"，则返回 "MyMessage"
// 如果 MessageMeta 为 nil，返回空字符串
func (self *MessageMeta) TypeName() string {
	if self == nil {
		return ""
	}

	return self.Type.Name()
}

// FullName 返回消息类型的完整名称（包含包名）
// 格式为 "包名.类型名"，例如 "proto.MyMessage"
// 用于唯一标识消息类型
// 如果 MessageMeta 为 nil，返回空字符串
func (self *MessageMeta) FullName() string {
	if self == nil {
		return ""
	}

	var sb strings.Builder
	// 获取包名（路径的最后一部分）
	sb.WriteString(path.Base(self.Type.PkgPath()))
	sb.WriteString(".")
	// 获取类型名
	sb.WriteString(self.Type.Name())

	return sb.String()
}

// NewType 创建消息类型的实例
// 使用反射创建一个新的消息对象
// 返回指向消息类型的指针，可以直接用于消息处理
// 如果 Type 为 nil，返回 nil
func (self *MessageMeta) NewType() interface{} {
	if self.Type == nil {
		return nil
	}

	// 使用反射创建新实例，返回指针类型
	return reflect.New(self.Type).Interface()
}

// SetContext 为消息元信息绑定上下文数据
// name: 上下文数据的键名
// data: 上下文数据的值，可以是任意类型
// 返回自身以便链式调用
// 如果键名已存在，则更新其值；否则添加新的上下文数据
func (self *MessageMeta) SetContext(name string, data interface{}) *MessageMeta {
	self.ctxListGuard.Lock()
	defer self.ctxListGuard.Unlock()

	// 查找是否已存在相同键名的上下文
	for _, ctx := range self.ctxList {
		if ctx.name == name {
			// 更新已存在的上下文数据
			ctx.data = data
			return self
		}
	}

	// 添加新的上下文数据
	self.ctxList = append(self.ctxList, &context{
		name: name,
		data: data,
	})

	return self
}

// GetContext 获取消息元信息的上下文数据
// key: 上下文数据的键名
// 返回上下文数据的值和是否存在
// 如果键名不存在，返回 nil, false
func (self *MessageMeta) GetContext(key string) (interface{}, bool) {
	self.ctxListGuard.RLock()
	defer self.ctxListGuard.RUnlock()

	// 遍历上下文列表查找匹配的键名
	for _, ctx := range self.ctxList {
		if ctx.name == key {
			return ctx.data, true
		}
	}

	return nil, false
}

// GetContextAsString 以字符串格式获取上下文数据
// key: 上下文数据的键名
// defaultValue: 如果键名不存在或类型不匹配时返回的默认值
// 返回字符串类型的上下文数据，如果获取失败返回默认值
func (self *MessageMeta) GetContextAsString(key, defaultValue string) string {
	if v, ok := self.GetContext(key); ok {
		// 尝试类型断言为字符串
		if str, ok := v.(string); ok {
			return str
		}
	}

	return defaultValue
}

// GetContextAsInt 以整数格式获取上下文数据
// name: 上下文数据的键名
// defaultValue: 如果键名不存在或类型不匹配时返回的默认值
// 返回整数类型的上下文数据，如果获取失败返回默认值
func (self *MessageMeta) GetContextAsInt(name string, defaultValue int) int {
	if v, ok := self.GetContext(name); ok {
		// 尝试类型断言为整数
		if intV, ok := v.(int); ok {
			return intV
		}
	}

	return defaultValue
}

var (
	// metaByFullName 通过消息完整名称（包名.类型名）查找消息元信息
	// 键为消息的完整名称，值为对应的 MessageMeta
	metaByFullName = map[string]*MessageMeta{}

	// metaByID 通过消息ID查找消息元信息
	// 键为消息ID，值为对应的 MessageMeta
	// 用于二进制协议中根据ID识别消息类型
	metaByID = map[int]*MessageMeta{}

	// metaByType 通过消息类型查找消息元信息
	// 键为 reflect.Type，值为对应的 MessageMeta
	// 用于根据消息对象类型查找元信息
	metaByType = map[reflect.Type]*MessageMeta{}
)

// 消息查找规则说明：
//
// HTTP 消息：
//   - 通过 Method URL -> Meta 查找
//   - 通过 Type -> Meta 查找
//
// 非 HTTP 消息：
//   - 通过 ID -> Meta 查找（二进制协议）
//   - 通过 Type -> Meta 查找

// RegisterMessageMeta 注册消息元信息
// 将消息的元信息注册到全局映射表中，支持通过名称、ID、类型查找
// meta: 要注册的消息元信息
// 返回注册后的 MessageMeta（可能与输入相同，但类型已统一）
//
// 注意：
//   - 消息ID必须唯一且不为0
//   - 消息类型和完整名称也必须唯一
//   - 如果存在重复注册，会触发 panic
func RegisterMessageMeta(meta *MessageMeta) *MessageMeta {
	// 注册时，统一转换为非指针类型
	// 这样无论注册时传入的是指针类型还是非指针类型，都能正确匹配
	if meta.Type.Kind() == reflect.Ptr {
		meta.Type = meta.Type.Elem()
	}

	// 检查类型是否已注册
	if _, ok := metaByType[meta.Type]; ok {
		panic(fmt.Sprintf("Duplicate message meta register by type: %d name: %s", meta.ID, meta.Type.Name()))
	} else {
		metaByType[meta.Type] = meta
	}

	// 检查完整名称是否已注册
	if _, ok := metaByFullName[meta.FullName()]; ok {
		panic(fmt.Sprintf("Duplicate message meta register by fullname: %s", meta.FullName()))
	} else {
		metaByFullName[meta.FullName()] = meta
	}

	// 检查消息ID是否有效
	if meta.ID == 0 {
		panic("message meta require 'ID' field: " + meta.TypeName())
	}

	// 检查消息ID是否已注册
	if prev, ok := metaByID[meta.ID]; ok {
		panic(fmt.Sprintf("Duplicate message meta register by id: %d type: %s, pre type: %s", meta.ID, meta.TypeName(), prev.TypeName()))
	} else {
		metaByID[meta.ID] = meta
	}

	return meta
}

// MessageMetaByFullName 根据消息完整名称查找消息元信息
// name: 消息的完整名称，格式为 "包名.类型名"，例如 "proto.MyMessage"
// 返回对应的 MessageMeta，如果不存在返回 nil
func MessageMetaByFullName(name string) *MessageMeta {
	if v, ok := metaByFullName[name]; ok {
		return v
	}

	return nil
}

// MessageMetaVisit 遍历匹配指定规则的消息元信息
// nameRule: 消息名称的匹配规则，支持正则表达式
// callback: 遍历回调函数，参数为匹配的 MessageMeta
// 如果回调返回 false，则停止遍历
// 返回错误信息，如果成功则返回 nil
func MessageMetaVisit(nameRule string, callback func(meta *MessageMeta) bool) error {
	// 编译正则表达式
	exp, err := regexp.Compile(nameRule)
	if err != nil {
		return err
	}

	// 遍历所有消息元信息
	for name, meta := range metaByFullName {
		// 检查名称是否匹配规则
		if exp.MatchString(name) {
			// 调用回调函数
			if !callback(meta) {
				// 回调返回 false，停止遍历
				return nil
			}
		}
	}

	return nil
}

// MessageMetaByType 根据消息类型查找消息元信息
// t: 消息的 reflect.Type
// 返回对应的 MessageMeta，如果不存在返回 nil
// 如果传入的是指针类型，会自动转换为非指针类型进行查找
func MessageMetaByType(t reflect.Type) *MessageMeta {
	if t == nil {
		return nil
	}

	// 统一转换为非指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if v, ok := metaByType[t]; ok {
		return v
	}

	return nil
}

// MessageMetaByMsg 根据消息对象获取消息元信息
// msg: 消息对象，可以是任意类型
// 返回对应的 MessageMeta，如果消息为 nil 或未注册返回 nil
func MessageMetaByMsg(msg interface{}) *MessageMeta {
	if msg == nil {
		return nil
	}

	// 通过消息对象的类型查找元信息
	return MessageMetaByType(reflect.TypeOf(msg))
}

// MessageMetaByID 根据消息ID查找消息元信息
// id: 消息的唯一标识符
// 返回对应的 MessageMeta，如果不存在返回 nil
// 主要用于二进制协议中根据ID识别消息类型
func MessageMetaByID(id int) *MessageMeta {
	if v, ok := metaByID[id]; ok {
		return v
	}

	return nil
}

// MessageToName 获取消息的类型名称（不包含包名）
// msg: 消息对象
// 返回消息的类型名称，如果消息为 nil 或未注册返回空字符串
// 例如，如果消息类型是 "proto.MyMessage"，则返回 "MyMessage"
func MessageToName(msg interface{}) string {
	if msg == nil {
		return ""
	}

	meta := MessageMetaByMsg(msg)
	if meta == nil {
		return ""
	}

	return meta.TypeName()
}

// MessageToID 获取消息的ID
// msg: 消息对象
// 返回消息的唯一标识符，如果消息为 nil 或未注册返回 0
func MessageToID(msg interface{}) int {
	if msg == nil {
		return 0
	}

	meta := MessageMetaByMsg(msg)
	if meta == nil {
		return 0
	}

	return int(meta.ID)
}

// MessageSize 计算消息编码后的字节大小
// msg: 消息对象
// 返回消息编码后的字节数，如果消息为 nil、未注册或编码失败返回 0
// 此方法会实际编码消息来计算大小，可能有一定的性能开销
func MessageSize(msg interface{}) int {
	if msg == nil {
		return 0
	}

	// 获取消息元信息
	meta := MessageMetaByType(reflect.TypeOf(msg))
	if meta == nil {
		return 0
	}

	// 将消息编码为字节数组
	raw, err := meta.Codec.Encode(msg, nil)

	if err != nil {
		return 0
	}

	return len(raw.([]byte))
}

// MessageToString 将消息转换为字符串表示
// msg: 消息对象
// 返回消息的字符串表示，如果消息为 nil 返回空字符串
// 如果消息实现了 String() 方法，则使用该方法；否则使用 fmt.Sprintf 格式化
func MessageToString(msg interface{}) string {
	if msg == nil {
		return ""
	}

	// 如果消息实现了 String() 方法，优先使用
	if stringer, ok := msg.(interface {
		String() string
	}); ok {
		return stringer.String()
	}

	// 否则使用默认格式化
	return fmt.Sprintf("%+v", msg)
}
