package cache

import (
	"context"
	"go-one/util"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端实例
var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			db = parsed
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		util.Log().Error("Redis连接失败: %v", err)
		return err
	}

	util.Log().Info("Redis连接成功")
	return nil
}

// Close 关闭Redis连接
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
