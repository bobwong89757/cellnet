# Cellnet 架构文档

## 概述

Cellnet 是一个组件化、高扩展性、高性能的开源服务器网络库，采用 Go 语言开发。它通过高度抽象的接口设计，支持多种传输协议和编码格式，提供了灵活的消息处理机制，适用于游戏服务器、设备间通信、RPC 服务等多种场景。

## 核心设计理念

1. **组件化设计**：通过接口抽象，实现 Peer、Processor、Codec 等组件的可插拔
2. **事件驱动**：基于事件队列的异步消息处理模型
3. **高度可扩展**：支持自定义 Peer、Processor、Codec，满足不同业务需求
4. **协议无关**：统一的接口设计，支持多种传输协议和编码格式混合使用

## 架构层次

Cellnet 采用分层架构设计，从下到上分为以下几个层次：

```text
┌─────────────────────────────────────────┐
│         应用层 (Application)             │
│     用户业务逻辑、消息处理回调            │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         处理器层 (Processor)             │
│   消息收发、事件钩子、消息分发            │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         编码层 (Codec)                   │
│   Protobuf/JSON/二进制/ProtoPlus         │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         传输层 (Peer)                    │
│    TCP/UDP/HTTP/WebSocket/KCP           │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         网络层 (Network)                 │
│         操作系统网络接口                 │
└─────────────────────────────────────────┘
```

## 核心组件

### 1. Peer（端）

Peer 是网络连接的抽象，代表一个网络端点，可以是服务器（Acceptor）或客户端（Connector）。

#### 接口定义

```go
type Peer interface {
    Start() Peer      // 启动端
    Stop()            // 停止端
    TypeName() string // 获取类型名称
}
```

#### 支持的协议类型

- **TCP**: `tcp.Acceptor` / `tcp.Connector`
- **UDP**: `udp.Acceptor` / `udp.Connector`
- **HTTP**: `http.Acceptor` / `http.Connector`
- **WebSocket**: `gorillaws.Acceptor` / `gorillaws.Connector`
- **KCP**: `kcp.Acceptor` / `kcp.Connector`

#### 扩展接口

- `PeerProperty`: 基础属性（名称、地址、队列）
- `SessionAccessor`: 会话访问（获取、遍历、关闭会话）
- `ContextSet`: 自定义属性存储
- `PeerCaptureIOPanic`: IO 层异常捕获

### 2. Session（会话）

Session 代表一个长连接会话，是消息收发的载体。

#### Session 接口定义

```go
type Session interface {
    Raw() interface{}      // 获取原始连接
    Peer() Peer            // 获取所属 Peer
    Send(msg interface{})  // 发送消息
    Close()                // 关闭连接
    ID() int64             // 会话 ID
}
```

#### 特性

- 每个 Session 有唯一的 ID
- 支持异步发送，内部维护发送队列
- 自动处理连接断开和错误恢复

### 3. Processor（处理器）

Processor 负责消息的收发处理流程，包括编码解码、事件钩子、消息分发等。

#### 核心接口

```go
// 消息传输器
type MessageTransmitter interface {
    OnRecvMessage(ses Session) (msg interface{}, err error)
    OnSendMessage(ses Session, msg interface{}) error
}

// 事件钩子
type EventHooker interface {
    OnInboundEvent(input Event) (output Event)   // 入站事件处理
    OnOutboundEvent(input Event) (output Event)  // 出站事件处理
}

// 事件回调
type EventCallback func(ev Event)
```

#### 处理流程

```text
接收消息流程：
网络数据 → Transmitter.OnRecvMessage → 解码 → Hooker.OnInboundEvent → EventCallback

发送消息流程：
EventCallback → Hooker.OnOutboundEvent → 编码 → Transmitter.OnSendMessage → 网络数据
```

#### 内置处理器

- `tcp.ltv`: TCP 长连接，LTV（Length-Type-Value）封包格式
- `udp.ltv`: UDP 数据包，LTV 封包格式
- `kcp.ltv`: KCP 可靠 UDP，LTV 封包格式
- `http`: HTTP 请求/响应处理
- `gorillaws`: WebSocket 消息处理

### 4. Codec（编解码器）

Codec 负责消息的序列化和反序列化。

#### Codec 接口定义

```go
type Codec interface {
    Encode(msgObj interface{}, ctx ContextSet) (data interface{}, err error)
    Decode(data interface{}, msgObj interface{}) error
    Name() string      // 编码器名称
    MimeType() string  // MIME 类型（HTTP 兼容）
}
```

#### 支持的编码格式

- **Protobuf**: Google Protocol Buffers
- **JSON**: JSON 格式，适合与第三方服务通信
- **Binary**: 二进制协议（goobjfmt），内存流直接序列化，低 GC
- **ProtoPlus**: 优化的 Protobuf 编码格式
- **Sproto**: Sproto 协议格式，轻量级二进制协议，常用于游戏开发
- **HTTP Form**: HTTP 表单格式

#### 混合编码支持

Cellnet 支持在同一应用中混合使用多种编码格式，通过消息注册时指定不同的 Codec 实现。

### 5. EventQueue（事件队列）

EventQueue 是事件驱动的核心，负责异步消息处理。

#### EventQueue 接口定义

```go
type EventQueue interface {
    StartLoop() EventQueue           // 启动事件循环
    StopLoop() EventQueue            // 停止事件循环
    Wait()                           // 等待退出
    Post(callback func())            // 投递事件
    EnableCapturePanic(v bool)       // 启用异常捕获
    Count() int                      // 获取事件数量
}
```

#### 实现说明

**eventQueue 是 EventQueue 接口的默认实现**，基于 `Pipe` 实现，提供线程安全的事件队列功能。

**核心组件：**

1. **Pipe（底层数据管道）**
   - 无界队列，用于存储和传递事件
   - 使用 `sync.Mutex` 保护队列的并发访问
   - 使用 `sync.Cond` 条件变量实现阻塞等待和唤醒机制
   - `Add()` 操作非阻塞，立即返回
   - `Pick()` 操作在队列为空时阻塞等待，有新事件时被唤醒

2. **事件循环（独立 goroutine）**
   - 通过 `StartLoop()` 在独立 goroutine 中启动
   - 持续调用 `Pipe.Pick()` 从队列取出事件
   - 使用 `protectedCall()` 执行回调，支持 panic 捕获

3. **线程安全机制**
   - 所有队列操作都通过 `sync.Mutex` 保护
   - 使用条件变量实现高效的阻塞/唤醒机制
   - 支持多个 goroutine 并发投递事件，事件循环串行处理

**实现结构：**

```go
type eventQueue struct {
    *Pipe                    // 底层数据管道（嵌入）
    endSignal sync.WaitGroup // 等待事件循环退出的同步信号
    capturePanic bool        // 是否启用异常捕获
    onPanic CapturePanicNotifyFunc // panic 通知函数
}
```

#### 处理模型

通过 EventQueue 可以实现多种处理模型：

1. **单线程异步模型**：适用于 MMORPG 等复杂交互场景，免加锁处理共享数据
2. **多线程同步模型**：适用于机器人逻辑，每个机器人独立 goroutine
3. **多线程并发模型**：适用于网关、消息转发、HTTP 服务器

### 6. MessageMeta（消息元信息）

MessageMeta 存储消息的元数据，包括消息类型、ID、编码器等。

#### 核心功能

- 消息注册：通过 `RegisterMessageMeta` 注册消息
- 消息查找：支持通过 ID、类型、全名查找消息元信息
- 上下文绑定：支持为消息绑定自定义上下文数据

#### 消息注册

```go
type MessageMeta struct {
    Codec Codec           // 消息编码器
    Type  reflect.Type    // 消息类型
    ID    int             // 消息 ID（二进制协议使用）
}
```

## 消息处理流程

### 接收流程（按协议区分）

不同协议的接收方式不同，主要分为两类：

#### 1. 面向连接的协议（TCP、KCP、WebSocket）

这些协议每个连接有独立的 socket，接收循环在 **Session 层面**：

```text
┌─────────────┐
│  网络数据   │
└──────┬──────┘
       │
       ↓ [1] 在独立的 goroutine 中持续循环
┌─────────────────────────────────────────┐
│  Session.recvLoop()                     │  接收循环（每个 Session 独立）
│  (tcpSession.recvLoop / KcpSession.recvLoop / wsSession.recvLoop) │
│                                         │
│  调用: self.ReadMessage(self)          │
└──────┬──────────────────────────────────┘
       │
       ↓ [2] 调用 CoreProcBundle.ReadMessage()
┌─────────────────────────────────────────┐
│  CoreProcBundle.ReadMessage(ses)        │  消息读取入口
│                                         │
│  调用: self.transmit.OnRecvMessage(ses) │
└──────┬──────────────────────────────────┘
       │
       ↓ [3] 调用具体的 Transmitter 实现
┌─────────────────────────────────────────┐
│  Transmitter.OnRecvMessage(ses)         │  消息传输器
│  (TCPMessageTransmitter / KCPMessageTransmitter / WSMessageTransmitter) │
│                                         │
│  【步骤 3.1】从网络读取原始数据          │
│  - TCP: 调用 util.RecvLTVPacket()      │
│    * 读取 Length(2字节) → 读取 Type(2字节) → 读取 Value(n字节) │
│  - KCP: 从 KCP Session 读取数据包       │
│  - WebSocket: 调用 conn.ReadMessage()  │
│    * 读取完整的 WebSocket 二进制消息帧   │
│                                         │
│  【步骤 3.2】解析封包格式，提取消息ID和数据 │
│  - TCP: util.RecvLTVPacket() 解析 LTV 格式 │
│    * 从包体中提取消息ID (Type字段)       │
│    * 提取消息数据 (Value字段)            │
│  - KCP: RecvPacket() 解析 LTV 格式      │
│  - WebSocket: 从消息帧中提取消息ID(前2字节)和数据 │
│                                         │
│  【步骤 3.3】调用 Codec 解码消息数据     │
│  - 调用: codec.DecodeMessage(msgID, msgData) │
│  - 根据消息ID查找消息元信息 (MessageMeta) │
│  - 使用对应的 Codec (Protobuf/JSON/Binary等) 解码 │
│  - 返回解码后的消息对象 (msg interface{}) │
│                                         │
│  - 返回: (msg interface{}, err error)  │
└──────┬──────────────────────────────────┘
       │
       ↓ [4] 返回到 recvLoop，收到解码后的消息对象
┌─────────────────────────────────────────┐
│  Session.recvLoop()                     │  接收循环（继续执行）
│                                         │
│  msg, err := self.ReadMessage(self)    │  ← 从步骤2返回，得到 msg
│  if err != nil { ... }                 │  处理错误情况
│                                         │
│  创建: &RecvMsgEvent{Ses: self, Msg: msg} │
│  调用: self.ProcEvent(ev)              │
└──────┬──────────────────────────────────┘
       │
       ↓ [5] 调用 CoreProcBundle.ProcEvent()
┌─────────────────────────────────────────┐
│  CoreProcBundle.ProcEvent(ev)           │  事件处理入口
│                                         │
│  1. 调用: self.hooker.OnInboundEvent(ev) │
└──────┬──────────────────────────────────┘
       │
       ↓ [6] 调用 Hooker 处理入站事件
┌─────────────────────────────────────────┐
│  Hooker.OnInboundEvent(ev)              │  事件钩子
│  (MsgHooker.OnInboundEvent)             │
│                                         │
│  - 处理 RPC 请求/响应                   │
│  - 处理 Relay 消息                      │
│  - 记录接收日志                         │
│  - 返回处理后的 Event（可能为 nil）      │
└──────┬──────────────────────────────────┘
       │
       ↓ [7] 返回处理后的 Event
┌─────────────────────────────────────────┐
│  CoreProcBundle.ProcEvent(ev)           │
│                                         │
│  2. 如果 ev != nil:                    │
│     调用: self.callback(ev)            │
└──────┬──────────────────────────────────┘
       │
       ↓ [8] 调用队列化的回调函数
┌─────────────────────────────────────────┐
│  NewQueuedEventCallback(callback)       │  队列化回调包装器
│                                         │
│  调用: SessionQueuedCall(ev.Session(),  │
│        func() { callback(ev) })        │
└──────┬──────────────────────────────────┘
       │
       ↓ [9] 获取 Peer 的事件队列并投递
┌─────────────────────────────────────────┐
│  SessionQueuedCall()                    │  会话队列调用
│                                         │
│  获取: ses.Peer().Queue()              │
│  调用: QueuedCall(queue, callback)     │
└──────┬──────────────────────────────────┘
       │
       ↓ [10] 投递到事件队列
┌─────────────────────────────────────────┐
│  QueuedCall(queue, callback)            │  队列调用
│                                         │
│  if queue != nil:                      │
│    调用: queue.Post(callback)          │
│  else:                                 │
│    立即执行: callback()                │
└──────┬──────────────────────────────────┘
       │
       ↓ [11] 添加到 Pipe 队列（非阻塞）
┌─────────────────────────────────────────┐
│  EventQueue.Post(callback)              │  事件队列投递
│  (eventQueue.Post)                      │
│                                         │
│  调用: self.Add(callback)              │
│  (Pipe.Add - 添加到无界队列)            │
│  - 将 callback 追加到队列列表           │
│  - 通过条件变量通知等待的接收者         │
│  - 立即返回，不阻塞调用者               │
└──────┬──────────────────────────────────┘
       │
       ↓ [12] 事件循环从队列取出并执行（异步）
┌─────────────────────────────────────────┐
│  EventQueue.StartLoop()                 │  事件循环（独立 goroutine）
│  (eventQueue.StartLoop)                 │
│  【在应用启动时已启动，持续运行】         │
│                                         │
│  for {                                 │
│    - 调用: self.Pick(&writeList)       │
│      * 如果队列为空，阻塞等待（条件变量） │
│      * 当有新事件时被唤醒，取出所有事件   │
│    - 遍历 writeList 中的每个 callback  │
│    - 调用: self.protectedCall(callback) │
│      * 执行回调函数（带 panic 保护）     │
│  }                                     │
└──────┬──────────────────────────────────┘
       │
       ↓ [13] 执行用户回调
┌─────────────────────────────────────────┐
│  EventCallback(ev)                      │  用户注册的回调函数
│  (用户业务逻辑处理)                      │
│                                         │
│  - 处理消息                             │
│  - 可以调用: ev.Session().Send(msg)    │
└─────────────────────────────────────────┘
```

**详细调用链：**

1. **Session.recvLoop()** - 在 Session 启动时创建的独立 goroutine 中运行，持续循环（TCP/KCP/WebSocket 都有）
2. **CoreProcBundle.ReadMessage()** - Session 嵌入的 CoreProcBundle 提供的方法，内部调用 transmit.OnRecvMessage()
3. **Transmitter.OnRecvMessage()** - 由 Processor 注册的具体实现（TCPMessageTransmitter/KCPMessageTransmitter/WSMessageTransmitter）
   - 3.1 从网络读取原始数据（TCP: util.RecvLTVPacket / KCP: 从 KCP Session / WebSocket: conn.ReadMessage）
   - 3.2 解析封包格式，提取消息ID和消息数据
   - 3.3 调用 `codec.DecodeMessage(msgID, msgData)` 解码消息数据，根据消息ID查找 MessageMeta，使用对应的 Codec 解码，返回消息对象
4. **返回到 Session.recvLoop()** - OnRecvMessage() 返回后，recvLoop 收到解码后的消息对象，创建 RecvMsgEvent 并调用 ProcEvent
5. **CoreProcBundle.ProcEvent()** - 事件处理入口，先调用 Hooker
6. **Hooker.OnInboundEvent()** - 处理 RPC、Relay、日志等，返回处理后的 Event
7. **CoreProcBundle.ProcEvent()** - 如果 Event 不为 nil，调用 callback
8. **NewQueuedEventCallback** - 由 Processor 注册时包装的队列化回调
9. **SessionQueuedCall()** - 获取 Session 对应 Peer 的事件队列
10. **QueuedCall()** - 判断是否有队列，有则调用 queue.Post()，无则立即执行 callback
11. **EventQueue.Post()** - 将回调函数添加到 Pipe 队列（非阻塞），通过条件变量通知等待的接收者
12. **EventQueue.StartLoop()** - 事件循环在独立 goroutine 中持续运行（应用启动时已启动），调用 Pipe.Pick() 阻塞等待新事件，当有新事件时被唤醒并取出执行
13. **EventCallback** - 最终执行用户注册的回调函数

#### 2. 无连接协议（UDP）

UDP 协议一个 socket 接收多个源的数据包，接收循环在 **Peer 层面**：

```text
┌─────────────┐
│  网络数据   │
└──────┬──────┘
       │
       ↓ [1] 在 Acceptor/Connector 的 goroutine 中循环
┌─────────────────────────────────────────┐
│  Peer.accept() / Peer.connect()        │  接收循环（在 Peer 中）
│  (udpAcceptor.accept / udpConnector.connect) │
│                                         │
│  - 从 UDP socket 读取数据包             │
│  - 获取源地址 (remoteAddr)              │
└──────┬──────────────────────────────────┘
       │
       ↓ [2] 根据地址获取或创建 Session
┌─────────────────────────────────────────┐
│  Peer.getSession(remoteAddr)            │  获取 Session
│  (udpAcceptor.getSession)               │
│                                         │
│  - 根据地址查找或创建 Session            │
│  - 更新 Session 超时时间                │
└──────┬──────────────────────────────────┘
       │
       ↓ [3] 调用 Session 的接收方法
┌─────────────────────────────────────────┐
│  Session.Recv(data)                     │  接收数据包
│  (udpSession.Recv)                      │
│                                         │
│  - 保存数据包到 pkt                     │
│  - 调用: self.ReadMessage(self)        │
└──────┬──────────────────────────────────┘
       │
       ↓ [4] 后续流程与 TCP 相同
┌─────────────────────────────────────────┐
│  CoreProcBundle.ReadMessage()           │  → [步骤 2-13 与 TCP 相同]
│  Transmitter.OnRecvMessage()            │
│  CoreProcBundle.ProcEvent()             │
│  Hooker.OnInboundEvent()                │
│  EventQueue.Post()                      │
│  EventCallback()                        │
└─────────────────────────────────────────┘
```

**关键差异说明：**

- **TCP/KCP/WebSocket**：每个 Session 有独立的 `recvLoop()` goroutine，持续从连接读取数据
  - **TCP**：使用 TCP socket，LTV（Length-Type-Value）封包格式
    - **LTV 格式**：Length(2字节) + Type(2字节消息ID) + Value(n字节消息数据)
    - **为什么需要长度字段**：TCP 是流式协议，多个消息可能粘在一起，需要长度字段来区分消息边界
  - **KCP**：基于 UDP 的可靠连接，LTV 封包格式
    - **为什么需要长度字段**：虽然底层 UDP 是数据报协议（每个 UDP 包都是独立的），但 KCP 在 UDP 之上实现了可靠传输，提供给应用层的接口是**流式的**（类似 TCP）
    - KCP 的 `Read()` 方法返回的是字节流，不是完整的数据报，可能会将多个应用层消息合并到一个 UDP 数据包，或者将一个大的应用层消息分割成多个 UDP 数据包
    - 因此需要长度字段来区分应用层消息的边界，解决类似 TCP 的粘包问题
  - **WebSocket**：使用 WebSocket 连接，二进制消息格式（2字节消息ID + 消息数据）
    - **为什么不需要长度字段**：WebSocket 协议在应用层已经保证了消息的完整性，每次 `ReadMessage()` 都会返回一个完整的消息帧，不会出现粘包问题，因此不需要额外的长度字段
- **UDP**：在 Acceptor/Connector 层面运行接收循环，根据数据包的源地址获取或创建对应的 Session，然后调用 `Session.Recv()` 方法处理
- **后续流程统一**：UDP 从 `Session.Recv()` 之后的流程与 TCP/KCP/WebSocket 完全相同

### 发送流程（统一）

所有协议的发送流程都是统一的，但实现细节略有不同：

#### TCP/KCP/WebSocket 发送流程

```text
┌─────────────────────────────────────────┐
│  Session.Send(msg)                      │  用户调用发送接口
│  (tcpSession.Send / KcpSession.Send / wsSession.Send) │
│                                         │
│  调用: self.sendQueue.Add(msg)         │
└──────┬──────────────────────────────────┘
       │
       ↓ [1] 添加到发送队列
┌─────────────────────────────────────────┐
│  sendQueue.Add(msg)                     │  发送队列（Pipe）
│  (Pipe.Add)                             │
│                                         │
│  - 非阻塞添加到队列                     │
│  - 通知 sendLoop 有新消息               │
└──────┬──────────────────────────────────┘
       │
       ↓ [2] 发送循环从队列取出消息
┌─────────────────────────────────────────┐
│  Session.sendLoop()                     │  发送循环（独立 goroutine）
│  (tcpSession.sendLoop / KcpSession.sendLoop / wsSession.sendLoop) │
│                                         │
│  - 循环调用: sendQueue.Pick(&writeList) │
│  - 遍历消息列表                         │
│  - 调用: self.SendMessage(ev)          │
└──────┬──────────────────────────────────┘
       │
       ↓ [3] 创建发送事件并处理
┌─────────────────────────────────────────┐
│  Session.sendLoop()                     │
│                                         │
│  创建: &SendMsgEvent{Ses: self, Msg: msg} │
│  调用: self.SendMessage(ev)            │
└──────┬──────────────────────────────────┘
       │
       ↓ [4] 调用 CoreProcBundle.SendMessage()
┌─────────────────────────────────────────┐
│  CoreProcBundle.SendMessage(ev)         │  消息发送入口
│                                         │
│  1. 调用: self.hooker.OnOutboundEvent(ev) │
└──────┬──────────────────────────────────┘
       │
       ↓ [5] 调用 Hooker 处理出站事件
┌─────────────────────────────────────────┐
│  Hooker.OnOutboundEvent(ev)             │  事件钩子
│  (MsgHooker.OnOutboundEvent)            │
│                                         │
│  - 处理 RPC 请求/响应                   │
│  - 处理 Relay 消息                      │
│  - 记录发送日志                         │
│  - 返回处理后的 Event（可能为 nil）      │
└──────┬──────────────────────────────────┘
       │
       ↓ [6] 返回处理后的 Event
┌─────────────────────────────────────────┐
│  CoreProcBundle.SendMessage(ev)         │
│                                         │
│  2. 如果 ev != nil && transmit != nil: │
│     调用: self.transmit.OnSendMessage(  │
│            ev.Session(), ev.Message())  │
└──────┬──────────────────────────────────┘
       │
       ↓ [7] 调用具体的 Transmitter 实现
┌─────────────────────────────────────────┐
│  Transmitter.OnSendMessage(ses, msg)    │  消息传输器
│  (TCPMessageTransmitter / KCPMessageTransmitter / WSMessageTransmitter) │
│                                         │
│  - TCP/KCP: 编码消息为 LTV 格式数据包    │
│    * 格式: Length(2字节) + Type(2字节) + Value(n字节) │
│    * 长度字段用于接收端正确解析消息边界   │
│  - WebSocket: 编码消息为二进制格式（2字节消息ID + 消息数据） │
│    * 格式: Type(2字节消息ID) + Value(n字节消息数据) │
│    * WebSocket 协议层自动处理消息边界，无需长度字段 │
│  - 写入网络连接                         │
└──────┬──────────────────────────────────┘
       │
       ↓
┌─────────────┐
│  网络数据   │
└─────────────┘
```

#### UDP 发送流程

```text
┌─────────────────────────────────────────┐
│  Session.Send(msg)                      │  用户调用发送接口
│  (udpSession.Send)                      │
│                                         │
│  直接调用: self.SendMessage(ev)        │
│  (不经过队列，立即发送)                  │
└──────┬──────────────────────────────────┘
       │
       ↓ [1] 后续流程与 TCP 相同
┌─────────────────────────────────────────┐
│  CoreProcBundle.SendMessage()           │  → [步骤 4-7 与 TCP 相同]
│  Hooker.OnOutboundEvent()               │
│  Transmitter.OnSendMessage()            │
└─────────────────────────────────────────┘
       │
       ↓
┌─────────────┐
│  网络数据   │
└─────────────┘
```

**发送流程关键差异：**

- **TCP/KCP/WebSocket**：使用发送队列（sendQueue），消息先入队，由 `sendLoop()` 异步发送
  - **TCP**：LTV 格式封包（Length + Type + Value）
    - 需要长度字段，因为 TCP 是流式协议，发送端需要明确告知接收端消息长度
  - **KCP**：LTV 格式封包（基于 UDP 的可靠传输）
    - 需要长度字段，因为 KCP 虽然基于 UDP（数据报协议），但它实现了可靠传输，提供给应用层的接口是流式的
    - KCP 的 `Read()` 方法返回字节流，可能包含多个应用层消息或部分消息，需要长度字段来区分消息边界
  - **WebSocket**：二进制消息格式（Type + Value，无长度字段）
    - 不需要长度字段，WebSocket 协议层会自动处理消息边界，`WriteMessage()` 会将整个消息作为一个完整的 WebSocket 帧发送
- **UDP**：直接调用 `SendMessage()`，立即编码并发送，不使用队列
- **后续处理统一**：都经过 Hooker 处理和 Transmitter 编码发送

## 高级特性

### 1. RPC（远程过程调用）

Cellnet 内置 RPC 支持，提供同步和异步两种调用方式。

#### 同步 RPC

```go
result, err := rpc.CallSync(session, &RequestMsg{}, time.Second*5)
```

适用于后台服务器向其他服务器请求数据后再继续处理事务。

#### 异步 RPC

```go
rpc.CallAsync(session, &RequestMsg{}, time.Second*5, func(result interface{}, err error) {
    // 处理结果
})
```

适用于单线程服务器逻辑，避免阻塞。

### 2. Relay（消息接力）

Relay 提供消息转发功能，支持服务器间消息传递。

### 3. 消息日志

Cellnet 内置消息日志功能，可以方便地查看收发消息的每个字段，便于调试和监控。

### 4. 定时器

提供定时器接口，支持延迟执行和循环执行。

## 目录结构

```text
cellnet/
├── codec/              # 编码支持
│   ├── binary/        # 二进制编码
│   │   └── binary.go  # Binary 编码器实现，使用 goobjfmt 进行内存流序列化
│   ├── json/          # JSON 编码
│   │   └── json.go    # JSON 编码器实现，用于 JSON 格式消息编解码
│   ├── gogopb/        # Protobuf 编码
│   │   └── gogopb.go  # Google Protocol Buffers 编码器实现（gogo 版本）
│   ├── protoplus/     # ProtoPlus 编码
│   │   └── protoplus.go # ProtoPlus 编码器实现，优化的 Protobuf 编码格式
│   ├── sproto/        # Sproto 编码
│   │   └── sproto.go  # Sproto 编码器实现，轻量级二进制协议
│   ├── httpform/      # HTTP 表单编码
│   │   ├── form.go    # HTTP 表单编码器实现
│   │   └── mapping.go # HTTP 表单字段映射工具
│   ├── httpjson/      # HTTP JSON 编码
│   │   └── json.go    # HTTP JSON 编码器实现
│   ├── codecreg.go    # 编码器注册管理，提供 RegisterCodec、GetCodec 等函数
│   └── msgcodec.go    # 消息编解码工具函数，提供 EncodeMessage、DecodeMessage 等
├── peer/              # 传输层实现
│   ├── tcp/           # TCP 端实现
│   │   ├── acceptor.go    # TCP 服务器端（Acceptor），接受客户端连接
│   │   ├── connector.go   # TCP 客户端（Connector），连接到服务器，支持自动重连
│   │   ├── session.go     # TCP 会话实现，管理单个 TCP 连接
│   │   └── syncconn.go    # TCP 同步连接器，提供同步连接接口
│   ├── udp/           # UDP 端实现
│   │   ├── acceptor.go    # UDP 服务器端（Acceptor），接受 UDP 数据包
│   │   ├── connector.go   # UDP 客户端（Connector），发送 UDP 数据包
│   │   ├── session.go     # UDP 会话实现，管理 UDP 连接状态
│   │   └── trackkey.go    # UDP 连接跟踪键，用于识别 UDP 连接
│   ├── http/          # HTTP 端实现
│   │   ├── acceptor.go        # HTTP 服务器端（Acceptor），处理 HTTP 请求
│   │   ├── connector.go       # HTTP 客户端（Connector），发送 HTTP 请求
│   │   ├── session.go         # HTTP 会话实现，封装 HTTP 请求/响应
│   │   ├── file.go            # HTTP 文件服务支持
│   │   ├── respond_html.go    # HTTP HTML 响应工具
│   │   ├── respond_msg.go     # HTTP 消息响应工具
│   │   └── respond_status.go  # HTTP 状态码响应工具
│   ├── gorillaws/     # WebSocket 端实现（基于 gorilla/websocket）
│   │   ├── acceptor.go    # WebSocket 服务器端（Acceptor）
│   │   ├── connector.go   # WebSocket 客户端（Connector）
│   │   ├── session.go     # WebSocket 会话实现
│   │   └── syncconn.go    # WebSocket 同步连接器
│   ├── kcp/           # KCP 端实现（可靠 UDP）
│   │   ├── acceptor.go    # KCP 服务器端（Acceptor）
│   │   ├── connector.go   # KCP 客户端（Connector）
│   │   ├── session.go     # KCP 会话实现
│   │   ├── syncconn.go    # KCP 同步连接器
│   │   └── trackkey.go    # KCP 连接跟踪键
│   ├── mysql/         # MySQL 数据库连接
│   │   ├── connector.go   # MySQL 连接器实现
│   │   └── wrapper.go     # MySQL 连接包装器
│   ├── redix/         # Redis 连接
│   │   └── connector.go   # Redis 连接器实现
│   ├── peerreg.go         # Peer 注册管理，提供 RegisterPeerCreator、NewPeer 等函数
│   ├── peerprop.go        # Peer 基础属性实现（CorePeerProperty）
│   ├── property.go        # Peer 属性接口定义
│   ├── sesmgr.go          # 会话管理器实现（CoreSessionManager），管理所有 Session
│   ├── sesidentify.go     # 会话标识符管理
│   ├── procbundle.go      # 处理器资源包接口和实现
│   ├── socketoption.go    # Socket 选项接口（TCP/UDP Socket 配置）
│   ├── iopanic.go         # IO 层异常捕获接口和实现
│   ├── runningtag.go      # 运行状态标记接口和实现
│   ├── sysmsgreg.go       # 系统消息注册
│   ├── redisparam.go      # Redis 参数配置
│   └── sqlparam.go        # SQL 参数配置
├── proc/              # 处理器实现
│   ├── tcp/           # TCP 处理器
│   │   ├── setup.go       # TCP 处理器设置，注册 "tcp.ltv" 处理器
│   │   ├── transmitter.go # TCP 消息传输器，实现 LTV（Length-Type-Value）封包
│   │   └── hooker.go      # TCP 事件钩子，处理连接事件
│   ├── udp/           # UDP 处理器
│   │   ├── setup.go       # UDP 处理器设置，注册 "udp.ltv" 处理器
│   │   ├── recv.go        # UDP 消息接收处理
│   │   └── send.go        # UDP 消息发送处理
│   ├── kcp/           # KCP 处理器
│   │   ├── setup.go       # KCP 处理器设置，注册 "kcp.ltv" 处理器
│   │   ├── recv.go        # KCP 消息接收处理
│   │   ├── send.go        # KCP 消息发送处理
│   │   └── hooker.go      # KCP 事件钩子
│   ├── http/          # HTTP 处理器
│   │   └── setup.go       # HTTP 处理器设置，注册 "http" 处理器
│   ├── gorillaws/     # WebSocket 处理器
│   │   ├── setup.go       # WebSocket 处理器设置，注册 "gorillaws.ltv" 处理器
│   │   ├── transmitter.go # WebSocket 消息传输器
│   │   └── hooker.go      # WebSocket 事件钩子
│   ├── procreg.go         # 处理器注册管理，提供 RegisterProcessor、BindProcessorHandler 等函数
│   ├── procbundle.go      # 处理器资源包接口定义
│   ├── msgdispatcher.go   # 消息派发器，根据消息类型自动派发到处理函数
│   └── syncrecv.go        # 同步接收处理
├── rpc/               # RPC 支持
│   ├── req.go             # RPC 请求核心实现，管理 RPC 请求生命周期
│   ├── req_sync.go        # 同步 RPC 调用实现
│   ├── req_async.go       # 异步 RPC 调用实现
│   ├── req_type.go        # RPC 请求类型定义
│   ├── event.go           # RPC 事件定义
│   ├── proc.go            # RPC 处理器，处理 RPC 请求和响应
│   ├── util.go            # RPC 工具函数
│   ├── msg.go             # RPC 消息定义
│   ├── msg.proto          # RPC 消息 Protobuf 定义
│   └── msg_gen.go         # RPC 消息生成的代码
├── relay/             # 消息接力
│   ├── relay.go           # 消息转发核心实现，支持服务器间消息传递
│   ├── broadcast.go       # 消息广播功能
│   ├── event.go           # Relay 事件定义
│   ├── proc.go            # Relay 处理器，处理转发消息
│   ├── msg.proto          # Relay 消息 Protobuf 定义
│   └── msg_gen.go         # Relay 消息生成的代码
├── timer/             # 定时器
│   ├── loop.go            # 循环定时器实现（Loop），支持持续 Tick 循环
│   ├── after.go           # 延迟执行定时器实现（After），支持单次延迟执行
│   └── loop_test.go       # 定时器测试代码
├── msglog/            # 消息日志
│   ├── proc.go            # 消息日志处理器，记录收发消息的详细信息
│   ├── blocker.go         # 消息日志阻塞器，控制哪些消息需要记录
│   ├── logcolor.go        # 消息日志颜色输出
│   └── listbase.go        # 消息日志列表基础实现
├── util/              # 工具库
│   ├── addr.go            # 地址处理工具，拆分和组合网络地址
│   ├── codec.go           # 编解码工具函数
│   ├── ioutil.go          # IO 工具函数
│   ├── packet.go          # 数据包处理工具
│   ├── queue.go           # 队列工具函数
│   ├── sys.go             # 系统工具函数
│   ├── kvfile.go          # 键值对文件读写工具
│   └── *_test.go          # 各工具函数的测试代码
├── log/               # 日志框架
│   └── framework_log.go   # 日志框架实现，基于 zap，提供统一的日志接口
├── protoc-gen-msg/    # 代码生成工具
│   ├── main.go            # Protobuf 消息代码生成器主程序
│   └── file.go            # 文件生成工具
├── examples/          # 示例代码
│   ├── chat/              # 聊天服务器示例
│   ├── echo/              # Echo 服务器示例（支持同步/异步 RPC）
│   ├── fileserver/        # 文件服务器示例
│   ├── kcp/               # KCP 协议示例
│   └── websocket/         # WebSocket 示例
├── peer.go            # Peer 接口定义（Peer、PeerProperty、SessionAccessor 等）
├── session.go         # Session 接口定义和 RawPacket 实现
├── processor.go       # Processor 相关接口定义（MessageTransmitter、EventHooker 等）
├── event.go           # Event 接口定义和事件类型（RecvMsgEvent、SendMsgEvent 等）
├── queue.go           # EventQueue 接口定义和实现（事件队列）
├── meta.go            # MessageMeta 消息元信息管理和注册
├── codec.go           # Codec 接口定义
├── pipe.go            # Pipe 无界队列实现，EventQueue 的底层实现
├── err.go             # Error 错误类型定义
├── sysmsg.go          # 系统消息定义（SessionAccepted、SessionConnected 等）
├── peer_tcp.go        # TCP Peer 便捷创建函数
├── peer_udp.go        # UDP Peer 便捷创建函数
├── peer_http.go       # HTTP Peer 便捷创建函数
├── peer_ws.go         # WebSocket Peer 便捷创建函数
└── peer_db.go         # 数据库 Peer 便捷创建函数
```

## 扩展机制

### 自定义 Peer

实现 `Peer` 接口和相关扩展接口，注册到系统中：

```go
peer.RegisterPeerCreator("custom.Acceptor", func() Peer {
    return &CustomAcceptor{}
})
```

### 自定义 Processor

实现 `MessageTransmitter` 和 `EventHooker` 接口，注册处理器：

```go
proc.RegisterProcessor("custom.processor", func(bundle ProcessorBundle, callback EventCallback, args ...interface{}) {
    bundle.SetTransmitter(&CustomTransmitter{})
    bundle.SetHooker(&CustomHooker{})
    bundle.SetCallback(callback)
})
```

### 自定义 Codec

实现 `Codec` 接口，注册编码器：

```go
codec.RegisterCodec(&CustomCodec{})
```

## 使用示例

### 基本服务器

```go
// 创建事件队列
queue := cellnet.NewEventQueue()

// 创建服务器端
peerIns := peer.NewGenericPeer("tcp.Acceptor", "server", "127.0.0.1:17701", queue)

// 绑定处理器
proc.BindProcessorHandler(peerIns, "tcp.ltv", func(ev cellnet.Event) {
    switch msg := ev.Message().(type) {
    case *cellnet.SessionAccepted:
        fmt.Println("client connected")
    case *YourMessage:
        // 处理消息
        ev.Session().Send(&YourResponse{})
    case *cellnet.SessionClosed:
        fmt.Println("client disconnected")
    }
})

// 启动
peerIns.Start()
queue.StartLoop()
```

### 基本客户端

```go
// 创建事件队列
queue := cellnet.NewEventQueue()

// 创建客户端
peerIns := peer.NewGenericPeer("tcp.Connector", "client", "127.0.0.1:17701", queue)

// 绑定处理器
proc.BindProcessorHandler(peerIns, "tcp.ltv", func(ev cellnet.Event) {
    switch msg := ev.Message().(type) {
    case *cellnet.SessionConnected:
        // 连接成功，发送消息
        ev.Session().Send(&YourMessage{})
    case *YourResponse:
        // 处理响应
    }
})

// 启动
peerIns.Start()
queue.StartLoop()
```

## 性能特性

1. **低 GC 压力**：二进制协议直接内存序列化，减少对象分配
2. **异步非阻塞**：基于事件队列的异步处理，避免阻塞
3. **连接池管理**：高效的会话管理和连接复用
4. **零拷贝优化**：在可能的情况下使用零拷贝技术

## 最佳实践

1. **消息注册**：在应用启动时统一注册所有消息
2. **异常处理**：启用 IO 层异常捕获，避免程序崩溃
3. **队列选择**：根据业务场景选择合适的队列处理模型
4. **编码选择**：服务器间通信使用 Protobuf，与 Web 服务使用 JSON
5. **会话管理**：合理使用 SessionAccessor 管理会话生命周期

## 总结

Cellnet 通过清晰的架构设计和灵活的扩展机制，为 Go 语言开发者提供了一个强大而灵活的网络框架。其组件化设计使得开发者可以根据具体需求选择合适的传输协议和编码格式，同时通过事件队列机制实现了灵活的消息处理模型。无论是简单的客户端-服务器通信，还是复杂的分布式系统，Cellnet 都能提供良好的支持。
