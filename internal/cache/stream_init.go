package cache

import (
	"fmt"
	"go-one/util"
)

// InitStreams 初始化Stream组件（可选功能）
// 如果项目需要使用Redis Stream，在conf.Init()中调用此函数
func InitStreams() error {
	// 初始化StreamManager
	if err := InitStreamManager(); err != nil {
		return fmt.Errorf("初始化StreamManager失败: %w", err)
	}

	util.Log().Info("Redis Stream组件初始化完成")
	return nil
}

// InitConsumerWithHandler 初始化消费者并启动
func InitConsumerWithHandler(name string, handler MessageHandler, streamName string) error {
	manager := GetStreamManager()
	if manager == nil {
		return fmt.Errorf("StreamManager未初始化")
	}

	// 使用默认配置
	config := DefaultConsumerConfig(streamName)

	// 创建消费者
	if err := manager.CreateConsumer(name, config, handler); err != nil {
		return fmt.Errorf("创建消费者失败: %w", err)
	}

	// 启动消费者
	if err := manager.StartConsumer(name); err != nil {
		return fmt.Errorf("启动消费者失败: %w", err)
	}

	util.Log().Info("消费者 %s 已创建并启动", name)
	return nil
}

// CreateCustomConsumer 创建自定义配置的消费者
func CreateCustomConsumer(name string, config ConsumerConfig, handler MessageHandler, autoStart bool) error {
	manager := GetStreamManager()
	if manager == nil {
		return fmt.Errorf("StreamManager未初始化")
	}

	// 创建消费者
	if err := manager.CreateConsumer(name, config, handler); err != nil {
		return fmt.Errorf("创建消费者失败: %w", err)
	}

	// 如果需要自动启动
	if autoStart {
		if err := manager.StartConsumer(name); err != nil {
			return fmt.Errorf("启动消费者失败: %w", err)
		}
		util.Log().Info("自定义消费者 %s 已创建并启动", name)
	} else {
		util.Log().Info("自定义消费者 %s 已创建", name)
	}

	return nil
}

// CreateSimpleStreamProducer 创建简单的流生产者
func CreateSimpleStreamProducer(name string, streamName string) error {
	manager := GetStreamManager()
	if manager == nil {
		return fmt.Errorf("StreamManager未初始化")
	}

	config := DefaultStreamConfig(streamName)
	err := manager.CreateSimpleProducer(name, config)
	if err != nil {
		return fmt.Errorf("创建生产者失败: %w", err)
	}

	util.Log().Info("简单生产者 %s 已创建", name)
	return nil
}

// CreateBackupStreamProducer 创建带备份的流生产者（高可用）
func CreateBackupStreamProducer(name string, streamName string) error {
	manager := GetStreamManager()
	if manager == nil {
		return fmt.Errorf("StreamManager未初始化")
	}

	config := DefaultBackupStreamConfig(streamName)
	err := manager.CreateBackupProducer(name, config)
	if err != nil {
		return fmt.Errorf("创建备份生产者失败: %w", err)
	}

	util.Log().Info("备份生产者 %s 已创建", name)
	return nil
}

// ShutdownStreams 关闭所有流组件
func ShutdownStreams() {
	manager := GetStreamManager()
	if manager != nil {
		manager.Shutdown()
	}
	util.Log().Info("所有流组件已关闭")
}
