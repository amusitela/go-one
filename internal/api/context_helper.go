package api

import (
	"go-one/internal/serializer"
	"go-one/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBusinessContext 从gin.Context获取BusinessContext
func GetBusinessContext(c *gin.Context) *service.BusinessContext {
	if bizCtx, exists := c.Get("business_context"); exists {
		if bc, ok := bizCtx.(*service.BusinessContext); ok {
			return bc
		}
	}
	// 如果没有从中间件获取到，创建一个新的
	return service.NewBusinessContext(c.Request.Context()).
		WithClientIP(c.ClientIP()).
		WithUserAgent(c.GetHeader("User-Agent"))
}

// HandleServiceError 处理ServiceError并转换为HTTP响应
// 返回适当的HTTP状态码和响应体
func HandleServiceError(c *gin.Context, err service.ServiceError) {
	if err == nil {
		return
	}

	code := err.GetCode()
	message := err.GetMessage()

	// 根据错误码范围判断HTTP状态码
	var httpStatus int
	switch {
	case code >= 40000 && code < 40100: // 参数验证错误
		httpStatus = http.StatusBadRequest
	case code >= 40001 && code < 40002: // 认证错误
		httpStatus = http.StatusUnauthorized
	case code >= 40003 && code < 40004: // 权限错误
		httpStatus = http.StatusForbidden
	case code >= 40004 && code < 40005: // 未找到错误
		httpStatus = http.StatusNotFound
	case code >= 40009 && code < 40010: // 冲突错误（如重复）
		httpStatus = http.StatusConflict
	case code >= 50000: // 服务器错误
		httpStatus = http.StatusInternalServerError
	default:
		httpStatus = http.StatusBadRequest
	}

	// 根据错误类型选择序列化器
	switch err.(type) {
	case *service.ValidationError:
		c.JSON(httpStatus, serializer.ParamErr(message, err))
	case *service.DatabaseError:
		c.JSON(httpStatus, serializer.DBErr(message, err))
	case *service.NotFoundError:
		c.JSON(httpStatus, serializer.Err(serializer.CodeNotFound, message, nil))
	case *service.AuthError:
		c.JSON(httpStatus, serializer.Err(serializer.CodeUnauthorized, message, nil))
	default:
		c.JSON(httpStatus, serializer.Err(code, message, nil))
	}
}
