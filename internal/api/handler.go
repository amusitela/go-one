package api

import (
	"go-one/internal/service"

	"gorm.io/gorm"
)

var (
	// HandlerApi 全局handler实例
	HandlerApi *Handler
)

// Handler 聚合所有API处理器的服务依赖
type Handler struct {
	serviceManager *service.ServiceManager
}

// NewHandler 创建Handler实例
func NewHandler(db *gorm.DB) *Handler {
	handler := &Handler{
		serviceManager: service.NewServiceManager(db),
	}
	HandlerApi = handler
	return handler
}
