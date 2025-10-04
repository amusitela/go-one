# Go-One 架构设计文档

## 架构概览

Go-One 采用经典的分层架构设计，遵循关注点分离和依赖倒置原则。

```
┌─────────────────────────────────────────┐
│         API Layer (Handler)             │  ← HTTP请求处理
├─────────────────────────────────────────┤
│       Service Layer (Business)          │  ← 业务逻辑
├─────────────────────────────────────────┤
│     Repository Layer (Data Access)      │  ← 数据访问
├─────────────────────────────────────────┤
│         Model Layer (Entities)          │  ← 数据模型
└─────────────────────────────────────────┘
```

## 核心设计模式

### 1. 依赖注入 (Dependency Injection)

通过 `ServiceManager` 统一管理服务依赖：

```go
// ServiceManager 负责创建和管理所有服务
type ServiceManager struct {
    userRepo    repository.UserRepository
    articleRepo repository.ArticleRepository
    // ... 其他依赖
}

// 通过工厂方法创建服务实例
func (sm *ServiceManager) NewUserService() *UserService {
    return NewUserService(sm.userRepo)
}
```

**优点：**
- 解耦：服务不直接创建依赖
- 可测试：易于注入mock对象
- 可维护：依赖关系清晰

### 2. 仓储模式 (Repository Pattern)

数据访问层抽象为接口：

```go
type UserRepository interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    // ... 其他方法
}
```

**优点：**
- 数据访问逻辑集中管理
- 易于替换数据源（PostgreSQL → MySQL）
- 便于单元测试（mock repository）

### 3. 服务层模式 (Service Layer)

业务逻辑封装在服务层：

```go
type UserService struct {
    userRepo repository.UserRepository
}

func (s *UserService) Register(username, email, password string) (*User, error) {
    // 业务逻辑：验证、加密、创建用户
}
```

**优点：**
- 业务逻辑与HTTP层解耦
- 可复用（多个API endpoint可共用）
- 便于单元测试

## 数据流向

### 请求流程

```
HTTP Request
    ↓
Router (中间件处理)
    ↓
Handler (API Layer)
    ↓
Service (Business Layer)
    ↓
Repository (Data Access Layer)
    ↓
Database
```

### 响应流程

```
Database
    ↓
Repository 返回 Model
    ↓
Service 处理业务逻辑
    ↓
Handler 序列化响应
    ↓
Router (中间件处理)
    ↓
HTTP Response
```

## 中间件链

```
Request
  → CORS中间件
  → 安全头部中间件
  → Sentry中间件
  → 限流中间件
  → JWT认证中间件
  → Handler处理
  → Response
```

## 目录职责

### `/cmd`
- 应用程序入口
- 初始化配置、数据库、服务器
- 处理优雅关闭

### `/internal/api`
- HTTP请求处理
- 参数验证和绑定
- 响应序列化
- 调用Service层

### `/internal/service`
- 业务逻辑实现
- 跨Repository的操作协调
- 数据转换和验证

### `/internal/repository`
- 数据库CRUD操作
- 查询构建
- 数据持久化

### `/internal/model`
- 数据库实体定义
- GORM模型配置
- 表关系定义

### `/internal/middleware`
- 请求预处理
- 认证授权
- 限流
- 日志
- 安全

### `/internal/conf`
- 配置加载和管理
- 环境变量处理
- 第三方服务初始化

### `/internal/cache`
- Redis缓存操作
- 缓存策略

### `/internal/serializer`
- 统一响应格式
- 错误码定义

### `/util`
- 通用工具函数
- 日志、JWT、加密等

## 错误处理

### 分层错误处理

**Repository层：**
```go
// 返回原始数据库错误
return nil, err
```

**Service层：**
```go
// 转换为业务错误
if err == gorm.ErrRecordNotFound {
    return nil, errors.New("用户不存在")
}
```

**API层：**
```go
// 转换为HTTP响应
if err != nil {
    c.JSON(http.StatusNotFound, serializer.Err(404, err.Error(), nil))
    return
}
```

## 数据库设计原则

1. **使用GORM的钩子方法** - BeforeCreate, AfterUpdate等
2. **软删除** - 使用DeletedAt字段
3. **时间戳** - CreatedAt和UpdatedAt自动管理
4. **索引优化** - 为常查询字段添加索引
5. **外键约束** - 使用GORM的关联定义

## 缓存策略

```go
// Cache-Aside模式
func (s *Service) GetUser(id uint) (*User, error) {
    // 1. 先查缓存
    if cached := cache.Get(key); cached != nil {
        return cached, nil
    }
    
    // 2. 缓存未命中，查数据库
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    cache.Set(key, user, ttl)
    return user, nil
}
```

## 配置管理

采用12-Factor App原则：
- 所有配置通过环境变量
- 不同环境使用不同.env文件
- 敏感信息不提交到代码库

## 安全考虑

1. **密码加密** - bcrypt算法
2. **JWT令牌** - 访问令牌+刷新令牌机制
3. **限流保护** - 令牌桶算法
4. **SQL注入防护** - GORM参数化查询
5. **XSS防护** - 响应头设置
6. **CSRF防护** - SameSite Cookie属性

## 性能优化

1. **数据库连接池** - 配置MaxIdleConns和MaxOpenConns
2. **索引优化** - 为高频查询字段添加索引
3. **缓存策略** - Redis缓存热点数据
4. **分页查询** - 避免全表扫描
5. **N+1问题** - 使用GORM的Preload

## 测试策略

### 单元测试
- Repository层：使用测试数据库或mock
- Service层：mock Repository接口
- Handler层：httptest模拟请求

### 集成测试
- 测试完整的API流程
- 使用测试数据库

## 扩展点

### 添加新的数据源
实现对应的Repository接口即可

### 添加新的认证方式
在middleware中添加新的认证中间件

### 添加新的缓存后端
在cache包中抽象接口，实现不同的缓存驱动

### 添加消息队列
在service_manager中注入消息队列客户端

## 最佳实践

1. **保持层次清晰** - 不要跨层调用
2. **依赖倒置** - 高层模块不依赖低层模块
3. **接口隔离** - Repository接口应该精简
4. **单一职责** - 每个Service只负责一个业务领域
5. **代码复用** - 共享逻辑提取到util包
6. **错误处理** - 每层负责处理自己的错误
7. **日志记录** - 关键操作记录日志
8. **优雅退出** - 处理系统信号，清理资源

## 常见问题

### Q: Service层可以调用其他Service吗？
A: 可以，但应该通过ServiceManager获取，避免循环依赖。

### Q: Repository可以调用Service吗？
A: 不可以，这违反了分层原则。

### Q: 业务逻辑应该放在哪里？
A: Service层。Repository只做数据访问，Handler只做请求处理。

### Q: 如何处理事务？
A: 在Service层使用GORM的Transaction方法。

```go
func (s *Service) ComplexOperation() error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 操作1
        // 操作2
        return nil
    })
}
```

## 参考资料

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [SOLID原则](https://en.wikipedia.org/wiki/SOLID)
- [12-Factor App](https://12factor.net/)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)

