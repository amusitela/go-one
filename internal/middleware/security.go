package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityMiddleware 安全头部中间件
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 启用浏览器XSS防护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 引荐来源策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 内容安全策略（根据实际需求调整）
		c.Header("Content-Security-Policy", "default-src 'self'")

		// 严格传输安全（仅在HTTPS时启用）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
