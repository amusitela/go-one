# Redis Stream 使用指南

Go-One 框架提供了完整的 Redis Stream 支持，包括生产者、消费者、消息处理、自动清理等功能。

## 📦 功能特性

- ✅ **简单生产者** - 直接发送消息到Stream
- ✅ **备份生产者** - 带备份队列的高可用生产者
- ✅ **智能路由** - 根据队列状态自动路由消息
- ✅ **消费者组** - 支持多消费者协同处理
- ✅ **自动清理** - 智能清理过期消息
- ✅ **优先级管理** - 不同优先级的流有不同的清理策略
- ✅ **统一管理** - StreamManager 统一管理所有流组件

## 🚀 快速开始

### 1. 初始化Stream组件

在 `internal/conf/conf.go` 的 `Init()` 函数中添加：

```go
// 如果需要使用Redis Stream，初始化Stream组件
if err := cache.InitStreams(); err != nil {
    util.Log().Panic("初始化Stream组件失败: %v", err)
}
```

### 2. 创建生产者

#### 简单生产者（适用于普通场景）

```go
import "go-one/internal/cache"

// 创建简单生产者
err := cache.CreateSimpleStreamProducer("my_producer", "my_stream")
if err != nil {
    log.Fatal(err)
}

// 发送消息
ctx := context.Background()
err = cache.AddMessage("my_producer", ctx, map[string]interface{}{
    "user_id": 123,
    "action":  "login",
    "time":    time.Now().Unix(),
})
```

#### 备份生产者（适用于高可用场景）

```go
// 创建带备份的生产者
err := cache.CreateBackupStreamProducer("important_producer", "important_stream")
if err != nil {
    log.Fatal(err)
}

// 发送消息（自动处理备份和容错）
err = cache.AddMessage("important_producer", ctx, map[string]interface{}{
    "order_id": "ORD-12345",
    "amount":   99.99,
    "status":   "pending",
})
```

### 3. 创建消费者

#### 定义消息处理器

```go
import (
    "context"
    "go-one/internal/cache"
    "go-one/util"
    "github.com/redis/go-redis/v9"
)

// 实现 MessageHandler 接口
type MyMessageHandler struct {
    // 可以注入依赖，如数据库、其他服务等
}

func (h *MyMessageHandler) HandleMessage(ctx context.Context, msg redis.XMessage) error {
    // 处理消息
    util.Log().Info("收到消息: %v", msg.Values)
    
    // 解析消息字段
    userID, _ := msg.Values["user_id"].(string)
    action, _ := msg.Values["action"].(string)
    
    // 执行业务逻辑
    // ...
    
    return nil // 返回nil表示处理成功，会自动ACK
}
```

#### 启动消费者

```go
// 创建handler实例
handler := &MyMessageHandler{}

// 初始化并启动消费者
err := cache.InitConsumerWithHandler("my_consumer", handler, "my_stream")
if err != nil {
    log.Fatal(err)
}
```

## 📝 完整示例

### 示例1：用户行为跟踪

```go
// 1. 在应用启动时创建生产者和消费者
func InitTracking() {
    // 创建生产者
    cache.CreateSimpleStreamProducer("user_tracking", "user_events")
    
    // 创建消费者
    handler := &UserEventHandler{}
    cache.InitConsumerWithHandler("tracking_consumer", handler, "user_events")
}

// 2. 在业务代码中发送事件
func TrackUserAction(userID uint, action string) {
    ctx := context.Background()
    cache.AddMessage("user_tracking", ctx, map[string]interface{}{
        "user_id":   userID,
        "action":    action,
        "timestamp": time.Now().Unix(),
        "ip":        "192.168.1.1",
    })
}

// 3. 处理器处理事件
type UserEventHandler struct{}

func (h *UserEventHandler) HandleMessage(ctx context.Context, msg redis.XMessage) error {
    // 解析事件
    userID, _ := strconv.ParseUint(msg.Values["user_id"].(string), 10, 64)
    action := msg.Values["action"].(string)
    timestamp, _ := strconv.ParseInt(msg.Values["timestamp"].(string), 10, 64)
    
    // 存储到数据库或发送到分析系统
    util.Log().Info("用户 %d 执行了操作: %s, 时间: %d", userID, action, timestamp)
    
    return nil
}
```

### 示例2：订单处理系统（高可用）

```go
// 1. 创建带备份的生产者
func InitOrderProcessing() {
    // 使用备份生产者确保订单不丢失
    cache.CreateBackupStreamProducer("order_processor", "orders")
    
    // 创建消费者
    handler := &OrderHandler{}
    cache.InitConsumerWithHandler("order_consumer", handler, "orders")
}

// 2. 提交订单
func SubmitOrder(order *Order) error {
    ctx := context.Background()
    return cache.AddMessage("order_processor", ctx, map[string]interface{}{
        "order_id":  order.ID,
        "user_id":   order.UserID,
        "amount":    order.Amount,
        "status":    "pending",
        "items":     order.Items,
    })
}

// 3. 处理订单
type OrderHandler struct {
    // 注入服务依赖
    orderService *service.OrderService
}

func (h *OrderHandler) HandleMessage(ctx context.Context, msg redis.XMessage) error {
    orderID := msg.Values["order_id"].(string)
    
    // 处理订单
    err := h.orderService.ProcessOrder(orderID)
    if err != nil {
        util.Log().Error("处理订单失败: %v", err)
        return err // 返回错误，消息不会被ACK，稍后会重试
    }
    
    util.Log().Info("订单 %s 处理成功", orderID)
    return nil
}
```

## 🔧 高级配置

### 自定义流配置

```go
import "go-one/internal/cache"

// 创建自定义配置的生产者
config := cache.StreamConfig{
    Name:              "custom_stream",
    GroupName:         "custom_group",
    MaxLength:         5000,          // 最大消息数
    BatchSize:         20,            // 批处理大小
    CleanupInterval:   10 * time.Minute,
    Priority:          cache.PriorityHigh,
    MaxAge:            2 * time.Hour, // 消息最大存活时间
    MinRetentionTime:  30 * time.Minute,
    MinRetentionCount: 500,
}

manager := cache.GetStreamManager()
manager.CreateSimpleProducer("custom_producer", config)
```

### 自定义消费者配置

```go
// 创建自定义配置的消费者
consumerConfig := cache.ConsumerConfig{
    StreamName:      "my_stream",
    GroupName:       "my_group",
    ConsumerName:    "worker_1",
    MaxMessages:     20000,
    CleanupInterval: 1 * time.Hour,
    ReadCount:       50,                // 每次读取50条消息
    BlockDuration:   5 * time.Second,   // 阻塞5秒等待新消息
}

handler := &MyMessageHandler{}
cache.CreateCustomConsumer("my_consumer", consumerConfig, handler, true)
```

## 📊 监控和管理

### 查看所有生产者和消费者

```go
manager := cache.GetStreamManager()

// 列出所有生产者
producers := manager.ListProducers()
fmt.Println("生产者:", producers)

// 列出所有消费者
consumers := manager.ListConsumers()
fmt.Println("消费者:", consumers)
```

### 停止特定的消费者

```go
manager := cache.GetStreamManager()
err := manager.StopConsumer("my_consumer")
if err != nil {
    log.Println("停止消费者失败:", err)
}
```

### 移除生产者或消费者

```go
manager := cache.GetStreamManager()

// 移除生产者
manager.RemoveProducer("my_producer")

// 移除消费者
manager.RemoveConsumer("my_consumer")
```

## 🛡️ 最佳实践

### 1. 选择正确的生产者类型

- **简单生产者**：适用于日志、监控、非关键数据
- **备份生产者**：适用于订单、支付、关键业务数据

### 2. 合理设置消息清理策略

```go
// 关键业务流
config := cache.DefaultStreamConfig("orders")
config.Priority = cache.PriorityCritical
config.MaxAge = 24 * time.Hour          // 保留24小时
config.MinRetentionCount = 10000        // 至少保留10000条

// 普通日志流
config := cache.DefaultStreamConfig("logs")
config.Priority = cache.PriorityLow
config.MaxAge = 1 * time.Hour           // 保留1小时
config.MinRetentionCount = 100
```

### 3. 错误处理

```go
func (h *MyHandler) HandleMessage(ctx context.Context, msg redis.XMessage) error {
    // 可恢复的错误：返回error，消息会保留在pending列表等待重试
    if err := h.process(msg); err != nil {
        if isRetriable(err) {
            return err // 稍后重试
        }
        
        // 不可恢复的错误：记录日志后返回nil
        util.Log().Error("不可恢复的错误: %v, 消息: %v", err, msg)
        return nil // ACK消息，避免阻塞
    }
    
    return nil
}
```

### 4. 优雅关闭

在应用关闭时，确保正确关闭Stream组件：

```go
// 在 main.go 的关闭逻辑中添加
defer cache.ShutdownStreams()
```

## 🔍 故障排查

### 问题1：消息未被消费

检查消费者是否正常启动：
```go
manager := cache.GetStreamManager()
consumer, err := manager.GetConsumer("my_consumer")
if err != nil {
    log.Println("消费者不存在")
} else {
    log.Println("消费者运行状态:", consumer.IsRunning())
}
```

### 问题2：消息堆积

查看Redis中的stream长度：
```bash
redis-cli XLEN my_stream
redis-cli XPENDING my_stream my_group
```

### 问题3：生产者无法发送消息

确认Redis连接正常且StreamManager已初始化：
```go
if cache.GetStreamManager() == nil {
    log.Fatal("StreamManager未初始化")
}
```

## 📚 参考资料

- [Redis Streams 官方文档](https://redis.io/docs/data-types/streams/)
- [Go-One 完整文档](README_CN.md)
- [架构设计文档](ARCHITECTURE.md)

---

**Happy Streaming! 🚀**

