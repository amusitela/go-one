package server

import (
	"go-one/internal/api"
	"go-one/internal/middleware"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// NewRouter 路由配置
func NewRouter() *gin.Engine {
	r := gin.Default()
	h := api.HandlerApi

	// 应用中间件
	r.Use(middleware.Cors())
	r.Use(middleware.SecurityMiddleware())

	// Sentry middleware 捕获请求中的 panic 与错误
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: true,
		Timeout:         2 * time.Second,
	}))

	// API版本1
	v1 := r.Group("/api/v1")

	// 公开路由（不需要认证）
	public := v1.Group("")
	{
		// 用户认证相关
		auth := public.Group("/auth")
		{
			// 对登录和注册接口应用IP限流（1分钟内最多6次）
			auth.Use(middleware.RateLimitMiddleware(6, 10*time.Second, "ip"))
			auth.POST("/register", h.UserRegister)
			auth.POST("/login", h.UserLogin)
			auth.POST("/refresh", h.RefreshToken) // 刷新令牌
			auth.POST("/logout", h.UserLogout)
		}

		// 健康检查
		public.GET("/ping", api.Ping)
	}

	// 受保护的路由（需要JWT认证）
	protected := v1.Group("")
	protected.Use(middleware.JWTMiddleware())
	protected.Use(middleware.RateLimitMiddleware(60, 1*time.Minute, "user"))
	{
		// 用户相关
		user := protected.Group("/user")
		{
			user.GET("/profile", h.GetUserProfile)
			user.PUT("/profile", h.UpdateUserProfile)
			user.POST("/change-password", h.ChangePassword)
			user.GET("/list", h.ListUsers)
		}
	}

	return r
}
