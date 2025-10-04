package cache

import (
	"context"
	"fmt"
	"go-one/util"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// XMessage Redis Stream消息类型别名
type XMessage = redis.XMessage

// MessageHandler 消息处理接口
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg redis.XMessage) error
}

// Consumer 消费者接口
type Consumer interface {
	Start() error
	Stop()
	IsRunning() bool
}

// StreamConsumer stream消费者
type StreamConsumer struct {
	client        *redis.Client
	config        ConsumerConfig
	handler       MessageHandler
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
	cleanupTicker *time.Ticker
	wg            sync.WaitGroup
}

// NewStreamConsumer 创建stream消费者
func NewStreamConsumer(config ConsumerConfig, handler MessageHandler) (*StreamConsumer, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis 客户端尚未初始化")
	}

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &StreamConsumer{
		client:  RedisClient,
		config:  config,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
		running: false,
	}

	// 创建消费者组
	createConsumerGroup(ctx, config.StreamName, config.GroupName)

	return consumer, nil
}

// Start 启动消费者
func (sc *StreamConsumer) Start() error {
	if sc.running {
		return fmt.Errorf("Stream消费者已在运行")
	}

	sc.running = true
	sc.cleanupTicker = time.NewTicker(sc.config.CleanupInterval)

	// 启动消费者
	sc.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				util.Log().Error("Stream消费者崩溃: %v", r)
				sc.running = false
			}
			sc.wg.Done()
		}()
		sc.consume()
	}()

	// 启动定期清理
	sc.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				util.Log().Error("Stream清理器崩溃: %v", r)
			}
			sc.wg.Done()
		}()
		sc.startCleanup()
	}()

	util.Log().Info("Stream消费者已启动: %s", sc.config.StreamName)
	return nil
}

// Stop 停止消费者
func (sc *StreamConsumer) Stop() {
	if !sc.running {
		return
	}

	sc.cancel()
	if sc.cleanupTicker != nil {
		sc.cleanupTicker.Stop()
	}
	sc.running = false
	sc.wg.Wait()
	util.Log().Info("Stream消费者已停止: %s", sc.config.StreamName)
}

// IsRunning 检查消费者是否在运行
func (sc *StreamConsumer) IsRunning() bool {
	return sc.running
}

// consume 消费循环
func (sc *StreamConsumer) consume() {
	for {
		select {
		case <-sc.ctx.Done():
			util.Log().Info("Stream消费者收到停止信号")
			sc.running = false
			return
		default:
			sc.processMessages()
		}
	}
}

// processMessages 处理消息
func (sc *StreamConsumer) processMessages() {
	streams, err := sc.client.XReadGroup(sc.ctx, &redis.XReadGroupArgs{
		Group:    sc.config.GroupName,
		Consumer: sc.config.ConsumerName,
		Streams:  []string{sc.config.StreamName, ">"},
		Count:    sc.config.ReadCount,
		Block:    sc.config.BlockDuration,
	}).Result()

	if err != nil {
		if err != redis.Nil {
			util.Log().Error("读取stream失败: %v", err)
		}
		return
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return
	}

	for _, msg := range streams[0].Messages {
		if err := sc.handler.HandleMessage(sc.ctx, msg); err != nil {
			util.Log().Error("处理消息失败: %v, msgID: %s", err, msg.ID)
		} else {
			sc.client.XAck(sc.ctx, sc.config.StreamName, sc.config.GroupName, msg.ID)
		}
	}
}

// startCleanup 启动定期清理
func (sc *StreamConsumer) startCleanup() {
	cleanupManager := NewStreamCleanupManager()
	if cleanupManager == nil {
		util.Log().Error("创建清理管理器失败")
		return
	}
	defer cleanupManager.Stop()

	streamConfig := StreamConfig{
		Name:              sc.config.StreamName,
		GroupName:         sc.config.GroupName,
		MaxLength:         sc.config.MaxMessages,
		CleanupInterval:   sc.config.CleanupInterval,
		Priority:          PriorityNormal,
		MaxAge:            2 * time.Hour,
		MinRetentionTime:  30 * time.Minute,
		MinRetentionCount: 100,
	}

	for {
		select {
		case <-sc.ctx.Done():
			return
		case <-sc.cleanupTicker.C:
			if err := cleanupManager.CleanupStream(streamConfig); err != nil {
				util.Log().Error("清理流失败: %v", err)
			}
		}
	}
}
