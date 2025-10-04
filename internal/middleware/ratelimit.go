package middleware

import (
	"context"
	"fmt"
	"go-one/internal/cache"
	"go-one/internal/serializer"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware 创建一个基于令牌桶算法的限流中间件
// limit: 桶容量（最大令牌数）
// period: 补充令牌的周期
// identifierType: 限流标识类型（"user" 或 "ip"）
func RateLimitMiddleware(limit int64, period time.Duration, identifierType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var identifier string
		switch identifierType {
		case "user":
			userID, exists := c.Get("userID")
			if !exists {
				c.JSON(http.StatusInternalServerError, serializer.Err(serializer.CodeError, "无法获取用户信息进行限流", nil))
				c.Abort()
				return
			}
			identifier = fmt.Sprintf("%v", userID)
		case "ip":
			identifier = c.ClientIP()
		default:
			c.JSON(http.StatusInternalServerError, serializer.Err(serializer.CodeError, "无效的限流标识符类型", nil))
			c.Abort()
			return
		}

		bucketKey := fmt.Sprintf("token_bucket:%s:%s:%s:%s",
			identifierType,
			identifier,
			c.Request.Method,
			c.FullPath(),
		)

		ctx := context.Background()
		now := time.Now().Unix()

		// 获取令牌桶状态
		pipe := cache.RedisClient.Pipeline()
		tokensCmd := pipe.HGet(ctx, bucketKey, "tokens")
		lastRefillCmd := pipe.HGet(ctx, bucketKey, "last_refill")
		_, err := pipe.Exec(ctx)

		var currentTokens int64 = limit
		var lastRefill int64 = now

		if err == nil {
			if tokens, err := tokensCmd.Result(); err == nil {
				if parsedTokens, err := strconv.ParseInt(tokens, 10, 64); err == nil {
					currentTokens = parsedTokens
				}
			}
			if refill, err := lastRefillCmd.Result(); err == nil {
				if parsedRefill, err := strconv.ParseInt(refill, 10, 64); err == nil {
					lastRefill = parsedRefill
				}
			}
		}

		// 计算需要补充的令牌
		tokensToAdd := (now - lastRefill) * int64(time.Second) / int64(period)
		currentTokens += tokensToAdd
		if currentTokens > limit {
			currentTokens = limit
		}

		if currentTokens <= 0 {
			nextRefillTime := lastRefill + int64(period)/int64(time.Second)
			retryAfter := nextRefillTime - now
			if retryAfter <= 0 {
				retryAfter = 1
			}

			c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Retry-After", strconv.FormatInt(retryAfter, 10))

			c.JSON(http.StatusTooManyRequests, serializer.Err(serializer.CodeTooManyRequests,
				fmt.Sprintf("请求过于频繁，请在 %d 秒后重试", retryAfter), nil))
			c.Abort()
			return
		}

		// 消耗一个令牌
		currentTokens--

		// 更新令牌桶状态
		pipe = cache.RedisClient.Pipeline()
		pipe.HSet(ctx, bucketKey, "tokens", currentTokens)
		pipe.HSet(ctx, bucketKey, "last_refill", now)
		pipe.Expire(ctx, bucketKey, period*time.Duration(limit)*2) // 设置过期时间为补满桶时间的2倍
		_, err = pipe.Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, serializer.Err(serializer.CodeError, "限流服务异常", err))
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(currentTokens, 10))

		c.Next()
	}
}
