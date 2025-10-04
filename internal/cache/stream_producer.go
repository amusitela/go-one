package cache

import (
	"context"
	"fmt"
	"go-one/util"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	streamLockKeyPrefix = "stream_lock:"
)

// Producer 生产者接口
type Producer interface {
	AddMessage(ctx context.Context, fields map[string]interface{}) error
	GetStreamName() string
	Stop()
}

// BackupProducer 带备用队列的生产者，用于高可用场景
type BackupProducer struct {
	client *redis.Client
	config BackupStreamConfig

	// 本地缓存，用于优化性能
	backupCheckMutex  sync.RWMutex
	lastBackupCheck   time.Time
	backupHasMessages bool

	// 后台任务控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// SimpleProducer 简单生产者，不需要备用队列
type SimpleProducer struct {
	client *redis.Client
	config StreamConfig
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewBackupProducer 创建带备用队列的生产者
func NewBackupProducer(config BackupStreamConfig) (*BackupProducer, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis 客户端尚未初始化")
	}

	ctx, cancel := context.WithCancel(context.Background())

	producer := &BackupProducer{
		client: RedisClient,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 创建消费者组
	createConsumerGroup(ctx, config.Name, config.GroupName)

	// 启动后台任务
	producer.wg.Add(2)
	go producer.startBackupToMainTransfer()
	go producer.startStreamCleanup()

	return producer, nil
}

// NewSimpleProducer 创建简单生产者
func NewSimpleProducer(config StreamConfig) (*SimpleProducer, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis 客户端尚未初始化")
	}

	ctx, cancel := context.WithCancel(context.Background())

	producer := &SimpleProducer{
		client: RedisClient,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 创建消费者组
	createConsumerGroup(ctx, config.Name, config.GroupName)

	// 启动清理任务
	producer.wg.Add(1)
	go producer.startStreamCleanup()

	return producer, nil
}

// createConsumerGroup 创建消费者组的辅助函数
func createConsumerGroup(ctx context.Context, streamName, groupName string) {
	err := RedisClient.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		util.Log().Warning("创建消费者组 %s 失败: %v", groupName, err)
	}
}

// BackupProducer 接口实现

// AddMessage BackupProducer的AddMessage实现，智能路由消息
func (p *BackupProducer) AddMessage(ctx context.Context, fields map[string]interface{}) error {
	// 步骤 1. 快速路径：检查本地缓存
	if p.isBackupStreamActive() {
		return p.addMessageToStream(ctx, p.config.BackupStream, fields)
	}

	// 步骤 2. 检查 Redis 中备用队列是否有消息
	backupLength, err := p.client.XLen(ctx, p.config.BackupStream).Result()
	p.updateBackupStatus(backupLength > 0)
	if err == nil && backupLength > 0 {
		return p.addMessageToStream(ctx, p.config.BackupStream, fields)
	}

	// 步骤 3. 尝试获取分布式锁
	lockKey := streamLockKeyPrefix + p.config.Name
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	acquired, err := p.client.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
	if err != nil {
		util.Log().Warning("获取流锁失败，将消息直接写入主队列: %v", err)
		return p.addMessageToStream(ctx, p.config.Name, fields)
	}

	if !acquired {
		util.Log().Warning("流锁正忙，将消息路由到备用队列: %s", p.config.BackupStream)
		return p.addMessageToStream(ctx, p.config.BackupStream, fields)
	}
	defer p.releaseLock(ctx, lockKey, lockValue)

	// 步骤 4. 检查主队列的pending消息数量
	pendingCount, err := p.getPendingCount(ctx, p.config.Name, p.config.GroupName)
	if err != nil {
		util.Log().Warning("获取主队列未ACK消息数量失败，仍将尝试写入主队列: %v", err)
		return p.addMessageToStream(ctx, p.config.Name, fields)
	}

	// 步骤 5. 根据pending数量决定路由
	if pendingCount >= p.config.MaxLength {
		util.Log().Warning("主队列未ACK消息已达到容量上限 (%d/%d)，将消息路由到备用队列", pendingCount, p.config.MaxLength)
		return p.addMessageToStream(ctx, p.config.BackupStream, fields)
	}

	return p.addMessageToStream(ctx, p.config.Name, fields)
}

func (p *BackupProducer) GetStreamName() string {
	return p.config.Name
}

func (p *BackupProducer) Stop() {
	p.cancel()
	p.wg.Wait()
}

// SimpleProducer 接口实现

func (p *SimpleProducer) AddMessage(ctx context.Context, fields map[string]interface{}) error {
	return p.addMessageToStream(ctx, p.config.Name, fields)
}

func (p *SimpleProducer) GetStreamName() string {
	return p.config.Name
}

func (p *SimpleProducer) Stop() {
	p.cancel()
	p.wg.Wait()
}

// BackupProducer 辅助方法

func (p *BackupProducer) isBackupStreamActive() bool {
	p.backupCheckMutex.RLock()
	defer p.backupCheckMutex.RUnlock()
	timeout := p.config.BackupCheckConfig.FastTimeout
	if p.backupHasMessages {
		timeout = p.config.BackupCheckConfig.SlowTimeout
	}
	return time.Since(p.lastBackupCheck) < timeout && p.backupHasMessages
}

func (p *BackupProducer) updateBackupStatus(hasMessages bool) {
	p.backupCheckMutex.Lock()
	defer p.backupCheckMutex.Unlock()
	p.lastBackupCheck = time.Now()
	p.backupHasMessages = hasMessages
}

func (p *BackupProducer) addMessageToStream(ctx context.Context, streamName string, fields map[string]interface{}) error {
	_, err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: fields,
	}).Result()
	if err != nil {
		util.Log().Error("向 Stream %s 添加消息失败: %v", streamName, err)
		return err
	}
	return nil
}

func (p *SimpleProducer) addMessageToStream(ctx context.Context, streamName string, fields map[string]interface{}) error {
	_, err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: fields,
	}).Result()
	if err != nil {
		util.Log().Error("向 Stream %s 添加消息失败: %v", streamName, err)
		return err
	}
	return nil
}

func (p *BackupProducer) getPendingCount(ctx context.Context, streamName, groupName string) (int64, error) {
	pendingResult, err := p.client.XPending(ctx, streamName, groupName).Result()
	if err != nil {
		return 0, err
	}
	return pendingResult.Count, nil
}

func (p *BackupProducer) releaseLock(ctx context.Context, key, value string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	p.client.Eval(ctx, script, []string{key}, value)
}

// 后台任务

func (p *BackupProducer) startBackupToMainTransfer() {
	defer p.wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			transferred, err := p.transferFromBackupToMain(p.ctx)
			if err != nil {
				util.Log().Error("后台转移备用队列消息时出错: %v", err)
			}

			// 动态调整检查频率
			if transferred > 0 {
				ticker.Reset(500 * time.Millisecond)
			} else {
				ticker.Reset(2 * time.Second)
			}
		}
	}
}

func (p *BackupProducer) transferFromBackupToMain(ctx context.Context) (int, error) {
	lockKey := streamLockKeyPrefix + p.config.Name
	lockValue := fmt.Sprintf("transfer_%d", time.Now().UnixNano())
	acquired, err := p.client.SetNX(ctx, lockKey, lockValue, p.config.LockTimeout).Result()
	if err != nil || !acquired {
		return 0, err
	}
	defer p.releaseLock(ctx, lockKey, lockValue)

	pendingCount, err := p.getPendingCount(ctx, p.config.Name, p.config.GroupName)
	if err != nil {
		return 0, fmt.Errorf("无法获取主队列未ACK消息数量: %w", err)
	}

	availableSpace := p.config.MaxLength - pendingCount
	if availableSpace <= 0 {
		return 0, nil
	}

	countToTransfer := p.config.TransferBatchSize
	if availableSpace < countToTransfer {
		countToTransfer = availableSpace
	}

	messages, err := p.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{p.config.BackupStream, "0-0"},
		Count:   countToTransfer,
	}).Result()
	if err != nil || len(messages) == 0 || len(messages[0].Messages) == 0 {
		p.updateBackupStatus(false)
		return 0, nil
	}

	pipe := p.client.Pipeline()
	msgsToTransfer := messages[0].Messages
	var idsToDelete []string
	for _, msg := range msgsToTransfer {
		pipe.XAdd(ctx, &redis.XAddArgs{Stream: p.config.Name, Values: msg.Values})
		idsToDelete = append(idsToDelete, msg.ID)
	}
	pipe.XDel(ctx, p.config.BackupStream, idsToDelete...)

	_, err = pipe.Exec(ctx)
	if err != nil {
		util.Log().Error("消息转移 Pipeline 执行失败: %v", err)
		return 0, fmt.Errorf("pipeline exec failed: %w", err)
	}

	util.Log().Debug("成功将 %d 条消息从备用队列转移到主队列", len(msgsToTransfer))
	return len(msgsToTransfer), nil
}

func (p *BackupProducer) startStreamCleanup() {
	defer p.wg.Done()

	cleanupManager := NewStreamCleanupManager()
	if cleanupManager == nil {
		util.Log().Error("创建清理管理器失败")
		return
	}
	defer cleanupManager.Stop()

	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if err := cleanupManager.CleanupStream(p.config.StreamConfig); err != nil {
				util.Log().Error("清理主流失败: %v", err)
			}
		}
	}
}

func (p *SimpleProducer) startStreamCleanup() {
	defer p.wg.Done()

	cleanupManager := NewStreamCleanupManager()
	if cleanupManager == nil {
		util.Log().Error("创建清理管理器失败")
		return
	}
	defer cleanupManager.Stop()

	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if err := cleanupManager.CleanupStream(p.config); err != nil {
				util.Log().Error("清理流失败: %v", err)
			}
		}
	}
}
