package cache

import (
	"context"
	"fmt"
	"go-one/util"
	"sync"
)

// StreamManager 流管理器，负责统一管理生产者和消费者
type StreamManager struct {
	producers map[string]Producer
	consumers map[string]Consumer
	mu        sync.RWMutex
}

// 全局StreamManager实例
var globalStreamManager *StreamManager
var streamManagerOnce sync.Once

// InitStreamManager 初始化全局流管理器
func InitStreamManager() error {
	var err error
	streamManagerOnce.Do(func() {
		globalStreamManager = &StreamManager{
			producers: make(map[string]Producer),
			consumers: make(map[string]Consumer),
		}
		util.Log().Info("StreamManager初始化完成")
	})
	return err
}

// GetStreamManager 获取全局流管理器实例
func GetStreamManager() *StreamManager {
	return globalStreamManager
}

// CreateBackupProducer 创建带备用队列的生产者
func (sm *StreamManager) CreateBackupProducer(name string, config BackupStreamConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.producers[name]; exists {
		return fmt.Errorf("生产者 %s 已存在", name)
	}

	producer, err := NewBackupProducer(config)
	if err != nil {
		return fmt.Errorf("创建备用生产者失败: %w", err)
	}

	sm.producers[name] = producer
	util.Log().Info("创建备用生产者: %s", name)
	return nil
}

// CreateSimpleProducer 创建简单生产者
func (sm *StreamManager) CreateSimpleProducer(name string, config StreamConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.producers[name]; exists {
		return fmt.Errorf("生产者 %s 已存在", name)
	}

	producer, err := NewSimpleProducer(config)
	if err != nil {
		return fmt.Errorf("创建简单生产者失败: %w", err)
	}

	sm.producers[name] = producer
	util.Log().Info("创建简单生产者: %s", name)
	return nil
}

// CreateConsumer 创建消费者
func (sm *StreamManager) CreateConsumer(name string, config ConsumerConfig, handler MessageHandler) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.consumers[name]; exists {
		return fmt.Errorf("消费者 %s 已存在", name)
	}

	consumer, err := NewStreamConsumer(config, handler)
	if err != nil {
		return fmt.Errorf("创建消费者失败: %w", err)
	}

	sm.consumers[name] = consumer
	util.Log().Info("创建消费者: %s", name)
	return nil
}

// GetProducer 获取生产者
func (sm *StreamManager) GetProducer(name string) (Producer, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	producer, exists := sm.producers[name]
	if !exists {
		return nil, fmt.Errorf("生产者 %s 不存在", name)
	}
	return producer, nil
}

// GetConsumer 获取消费者
func (sm *StreamManager) GetConsumer(name string) (Consumer, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	consumer, exists := sm.consumers[name]
	if !exists {
		return nil, fmt.Errorf("消费者 %s 不存在", name)
	}
	return consumer, nil
}

// StartConsumer 启动指定消费者
func (sm *StreamManager) StartConsumer(name string) error {
	consumer, err := sm.GetConsumer(name)
	if err != nil {
		return err
	}

	return consumer.Start()
}

// StopConsumer 停止指定消费者
func (sm *StreamManager) StopConsumer(name string) error {
	consumer, err := sm.GetConsumer(name)
	if err != nil {
		return err
	}

	consumer.Stop()
	return nil
}

// RemoveProducer 移除生产者
func (sm *StreamManager) RemoveProducer(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	producer, exists := sm.producers[name]
	if !exists {
		return fmt.Errorf("生产者 %s 不存在", name)
	}

	producer.Stop()
	delete(sm.producers, name)
	util.Log().Info("移除生产者: %s", name)
	return nil
}

// RemoveConsumer 移除消费者
func (sm *StreamManager) RemoveConsumer(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	consumer, exists := sm.consumers[name]
	if !exists {
		return fmt.Errorf("消费者 %s 不存在", name)
	}

	consumer.Stop()
	delete(sm.consumers, name)
	util.Log().Info("移除消费者: %s", name)
	return nil
}

// ListProducers 列出所有生产者名称
func (sm *StreamManager) ListProducers() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.producers))
	for name := range sm.producers {
		names = append(names, name)
	}
	return names
}

// ListConsumers 列出所有消费者名称
func (sm *StreamManager) ListConsumers() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.consumers))
	for name := range sm.consumers {
		names = append(names, name)
	}
	return names
}

// Shutdown 关闭所有生产者和消费者
func (sm *StreamManager) Shutdown() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 停止所有消费者
	for name, consumer := range sm.consumers {
		consumer.Stop()
		util.Log().Info("停止消费者: %s", name)
	}

	// 停止所有生产者
	for name, producer := range sm.producers {
		producer.Stop()
		util.Log().Info("停止生产者: %s", name)
	}

	util.Log().Info("StreamManager关闭完成")
}

// --- 便捷方法 ---

// AddMessage 向指定生产者添加消息的便捷方法
func AddMessage(producerName string, ctx context.Context, fields map[string]interface{}) error {
	if globalStreamManager == nil {
		return fmt.Errorf("StreamManager未初始化")
	}

	producer, err := globalStreamManager.GetProducer(producerName)
	if err != nil {
		return err
	}

	return producer.AddMessage(ctx, fields)
}
