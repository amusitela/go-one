package cache

import (
	"context"
	"fmt"
	"go-one/util"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamCleanupManager 流清理管理器
type StreamCleanupManager struct {
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewStreamCleanupManager 创建流清理管理器
func NewStreamCleanupManager() *StreamCleanupManager {
	if RedisClient == nil {
		util.Log().Error("Redis客户端尚未初始化")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &StreamCleanupManager{
		client: RedisClient,
		ctx:    ctx,
		cancel: cancel,
	}
}

// CleanupStream 智能清理流消息
func (scm *StreamCleanupManager) CleanupStream(config StreamConfig) error {
	streamName := config.Name

	streamInfo, err := scm.client.XInfoStream(scm.ctx, streamName).Result()
	if err != nil {
		if err == redis.Nil {
			util.Log().Debug("流不存在: %s", streamName)
			return nil
		}
		return fmt.Errorf("获取流信息失败: %w", err)
	}

	if streamInfo.Length == 0 {
		return nil
	}

	safeMinID, err := scm.calculateSafeMinID(config)
	if err != nil {
		return fmt.Errorf("计算安全清理ID失败: %w", err)
	}

	return scm.performCleanup(streamName, safeMinID, config)
}

// calculateSafeMinID 计算安全的最小ID
func (scm *StreamCleanupManager) calculateSafeMinID(config StreamConfig) (string, error) {
	ageBasedMinID := scm.calculateMinIDByAge(config.MaxAge)

	ackBasedMinID, err := scm.calculateMinIDByACK(config.Name, config.GroupName)
	if err != nil {
		util.Log().Warning("计算基于ACK的最小ID失败: %v", err)
		ackBasedMinID = "0-0"
	}

	priorityMinID := scm.calculateMinIDByPriority(config)

	return scm.selectSafestID(ageBasedMinID, ackBasedMinID, priorityMinID), nil
}

// calculateMinIDByAge 基于最大年龄计算最小ID
func (scm *StreamCleanupManager) calculateMinIDByAge(maxAge time.Duration) string {
	if maxAge <= 0 {
		return "0-0"
	}

	cutoffTime := time.Now().Add(-maxAge)
	timestampMs := cutoffTime.UnixMilli()
	return fmt.Sprintf("%d-0", timestampMs)
}

// calculateMinIDByACK 基于未ACK消息计算最小ID
func (scm *StreamCleanupManager) calculateMinIDByACK(streamName, groupName string) (string, error) {
	pendingInfo, err := scm.client.XPending(scm.ctx, streamName, groupName).Result()
	if err != nil {
		if err == redis.Nil || strings.Contains(err.Error(), "NOGROUP") {
			return "+", nil
		}
		return "", err
	}

	if pendingInfo.Count == 0 {
		return "+", nil
	}

	return pendingInfo.Lower, nil
}

// calculateMinIDByPriority 基于流优先级计算最小保留ID
func (scm *StreamCleanupManager) calculateMinIDByPriority(config StreamConfig) string {
	switch config.Priority {
	case PriorityCritical, PriorityHigh:
		if config.MinRetentionTime > 0 {
			cutoffTime := time.Now().Add(-config.MinRetentionTime)
			timestampMs := cutoffTime.UnixMilli()
			return fmt.Sprintf("%d-0", timestampMs)
		}
	}
	return "0-0"
}

// selectSafestID 选择最安全的ID
func (scm *StreamCleanupManager) selectSafestID(ids ...string) string {
	safestID := "+"
	for _, id := range ids {
		if id == "0-0" || id == "" {
			continue
		}

		if safestID == "+" || scm.compareStreamIDs(id, safestID) < 0 {
			safestID = id
		}
	}
	return safestID
}

// compareStreamIDs 比较两个流ID
func (scm *StreamCleanupManager) compareStreamIDs(id1, id2 string) int {
	if id1 == "+" || id2 == "+" {
		if id1 == id2 {
			return 0
		}
		if id1 == "+" {
			return 1
		}
		return -1
	}

	parts1 := strings.Split(id1, "-")
	parts2 := strings.Split(id2, "-")

	if len(parts1) != 2 || len(parts2) != 2 {
		return 0
	}

	timestamp1, _ := strconv.ParseInt(parts1[0], 10, 64)
	timestamp2, _ := strconv.ParseInt(parts2[0], 10, 64)

	if timestamp1 < timestamp2 {
		return -1
	} else if timestamp1 > timestamp2 {
		return 1
	}

	seq1, _ := strconv.ParseInt(parts1[1], 10, 64)
	seq2, _ := strconv.ParseInt(parts2[1], 10, 64)

	if seq1 < seq2 {
		return -1
	} else if seq1 > seq2 {
		return 1
	}
	return 0
}

// performCleanup 执行实际的清理操作
func (scm *StreamCleanupManager) performCleanup(streamName, minID string, config StreamConfig) error {
	if minID == "0-0" || minID == "" {
		return nil
	}

	var deletedCount int64
	var err error

	if minID == "+" {
		streamInfo, infoErr := scm.client.XInfoStream(scm.ctx, streamName).Result()
		if infoErr != nil {
			return fmt.Errorf("获取流信息失败: %w", infoErr)
		}

		keepCount := config.MaxLength
		if config.MinRetentionCount > 0 && keepCount < config.MinRetentionCount {
			keepCount = config.MinRetentionCount
		}

		if streamInfo.Length > keepCount {
			deletedCount, err = scm.client.XTrimMaxLen(scm.ctx, streamName, keepCount).Result()
		}
	} else {
		deletedCount, err = scm.client.XTrimMinID(scm.ctx, streamName, minID).Result()

		if err == nil && config.MinRetentionCount > 0 {
			streamInfo, infoErr := scm.client.XInfoStream(scm.ctx, streamName).Result()
			if infoErr == nil && streamInfo.Length < config.MinRetentionCount {
				util.Log().Debug("流 [%s] 已达到最小保留数量要求 (%d), 停止清理",
					streamName, config.MinRetentionCount)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("清理失败: %w", err)
	}

	if deletedCount > 0 {
		util.Log().Info("流清理完成 [%s]: 删除了 %d 条消息 (优先级: %v, 最小ID: %s)",
			streamName, deletedCount, config.Priority, minID)
	}

	return nil
}

// Stop 停止清理管理器
func (scm *StreamCleanupManager) Stop() {
	if scm.cancel != nil {
		scm.cancel()
	}
}
