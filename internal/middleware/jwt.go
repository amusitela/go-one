package middleware

import (
	"go-one/internal/serializer"
	"go-one/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware JWT认证中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "未提供认证令牌", nil))
			c.Abort()
			return
		}

		// Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "认证令牌格式错误", nil))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, &service.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(service.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "认证令牌无效或已过期", nil))
			c.Abort()
			return
		}

		// 获取claims并创建BusinessContext
		if claims, ok := token.Claims.(*service.JWTClaims); ok {
			// 创建BusinessContext并注入上下文
			bizCtx := service.NewBusinessContext(c.Request.Context()).
				WithClaims(claims).
				WithClientIP(c.ClientIP()).
				WithUserAgent(c.GetHeader("User-Agent"))

			c.Set("business_context", bizCtx)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, serializer.Err(serializer.CodeUnauthorized, "无法解析认证信息", nil))
			c.Abort()
			return
		}
	}
}
