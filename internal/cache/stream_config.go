package cache

import "time"

// Priority 流优先级
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// StreamConfig 流配置
type StreamConfig struct {
	Name              string        // 流名称
	GroupName         string        // 消费者组名称
	MaxLength         int64         // 最大长度
	BatchSize         int64         // 批处理大小
	CleanupInterval   time.Duration // 清理间隔
	Priority          Priority      // 优先级
	MaxAge            time.Duration // 消息最大存活时间
	MinRetentionTime  time.Duration // 最小保留时间（关键流）
	MinRetentionCount int64         // 最小保留数量
}

// BackupStreamConfig 带备用队列的流配置
type BackupStreamConfig struct {
	StreamConfig
	BackupStream      string            // 备用流名称
	TransferBatchSize int64             // 转移批处理大小
	LockTimeout       time.Duration     // 锁超时时间
	BackupCheckConfig BackupCheckConfig // 备用队列检查配置
}

// BackupCheckConfig 备用队列检查配置
type BackupCheckConfig struct {
	FastTimeout time.Duration // 快速超时（当备用队列为空）
	SlowTimeout time.Duration // 慢速超时（当备用队列有消息）
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	StreamName      string        // 流名称
	GroupName       string        // 消费者组名称
	ConsumerName    string        // 消费者名称
	MaxMessages     int64         // 最大消息数
	CleanupInterval time.Duration // 清理间隔
	ReadCount       int64         // 每次读取数量
	BlockDuration   time.Duration // 阻塞时间
}

// DefaultStreamConfig 默认流配置
func DefaultStreamConfig(name string) StreamConfig {
	return StreamConfig{
		Name:              name,
		GroupName:         name + "_group",
		MaxLength:         1000,
		BatchSize:         10,
		CleanupInterval:   5 * time.Minute,
		Priority:          PriorityNormal,
		MaxAge:            1 * time.Hour,
		MinRetentionTime:  10 * time.Minute,
		MinRetentionCount: 100,
	}
}

// DefaultBackupStreamConfig 默认备用流配置
func DefaultBackupStreamConfig(name string) BackupStreamConfig {
	return BackupStreamConfig{
		StreamConfig:      DefaultStreamConfig(name),
		BackupStream:      name + "_backup",
		TransferBatchSize: 10,
		LockTimeout:       10 * time.Second,
		BackupCheckConfig: BackupCheckConfig{
			FastTimeout: 50 * time.Millisecond,
			SlowTimeout: 200 * time.Millisecond,
		},
	}
}

// DefaultConsumerConfig 默认消费者配置
func DefaultConsumerConfig(streamName string) ConsumerConfig {
	return ConsumerConfig{
		StreamName:      streamName,
		GroupName:       streamName + "_group",
		ConsumerName:    streamName + "_consumer",
		MaxMessages:     10000,
		CleanupInterval: 30 * time.Minute,
		ReadCount:       10,
		BlockDuration:   1 * time.Second,
	}
}
