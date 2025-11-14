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

### 完整流程

```text
┌─────────────┐
│  网络数据   │
└──────┬──────┘
       │
       ↓
┌─────────────────────┐
│  Session.recvLoop   │  接收循环
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ Transmitter         │  消息传输器
│ OnRecvMessage       │  解码网络数据
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ Hooker              │  事件钩子
│ OnInboundEvent      │  入站事件处理
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ EventQueue.Post     │  投递到事件队列
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ EventCallback       │  用户回调处理
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ Session.Send        │  发送响应
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ Hooker              │  事件钩子
│ OnOutboundEvent     │  出站事件处理
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│ Transmitter         │  消息传输器
│ OnSendMessage       │  编码并发送
└──────┬──────────────┘
       │
       ↓
┌─────────────┐
│  网络数据   │
└─────────────┘
```

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
│   ├── json/          # JSON 编码
│   ├── gogopb/        # Protobuf 编码
│   ├── protoplus/     # ProtoPlus 编码
│   ├── sproto/        # Sproto 编码
│   └── httpform/      # HTTP 表单编码
├── peer/              # 传输层实现
│   ├── tcp/           # TCP 端实现
│   ├── udp/           # UDP 端实现
│   ├── http/          # HTTP 端实现
│   ├── gorillaws/     # WebSocket 端实现
│   └── kcp/           # KCP 端实现
├── proc/              # 处理器实现
│   ├── tcp/           # TCP 处理器
│   ├── udp/           # UDP 处理器
│   ├── http/          # HTTP 处理器
│   └── gorillaws/     # WebSocket 处理器
├── rpc/               # RPC 支持
├── relay/             # 消息接力
├── timer/             # 定时器
├── msglog/            # 消息日志
├── util/              # 工具库
└── examples/          # 示例代码
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
