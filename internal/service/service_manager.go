package service

import (
	"go-one/internal/repository"

	"gorm.io/gorm"
)

// ServiceManager 统一管理所有服务的依赖注入
type ServiceManager struct {
	// Repositories
	userRepo repository.UserRepository

	// 可以在这里添加其他依赖
	// 例如：缓存服务、消息队列、第三方API客户端等
}

// NewServiceManager 创建服务管理器
func NewServiceManager(db *gorm.DB) *ServiceManager {
	return &ServiceManager{
		userRepo: repository.NewUserRepository(db),
	}
}

// NewUserService 创建用户服务
func (sm *ServiceManager) NewUserService() *UserService {
	return NewUserService(sm.userRepo)
}

// 可以在这里添加其他服务的工厂方法
// 例如：
// func (sm *ServiceManager) NewProductService() *ProductService {
//     return NewProductService(sm.productRepo)
// }
