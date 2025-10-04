# Go-One 企业级后端开发框架

> 🚀 一个从生产环境提炼的现代化 Go 后端脚手架，开箱即用，快速构建企业级应用

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Framework](https://img.shields.io/badge/framework-Gin-orange)](https://gin-gonic.com)

[English](README_EN.md) | **简体中文**

---

## 📖 目录

- [核心特性](#-核心特性)
- [5分钟快速开始](#-5分钟快速开始)
- [项目架构](#-项目架构)
- [开发指南](#-开发指南)
- [API文档](#-api文档)
- [部署指南](#-部署指南)
- [配置说明](#-配置说明)
- [常见问题](#-常见问题)

---

## ✨ 核心特性

### 🏗️ 清洁架构设计

- **完全解耦的分层架构** - Service层与HTTP传输层完全分离
- **BusinessContext** - 统一的业务上下文，独立于Web框架
- **ServiceError** - 标准化的错误处理机制
- **依赖注入** - ServiceManager管理所有依赖

```
HTTP Layer (Gin) → BusinessContext → Service Layer (框架无关) → Repository → Database
```

### 🔒 企业级安全

- ✅ **JWT认证** - 完整的用户认证系统，Token包含用户信息
- ✅ **智能限流** - 基于令牌桶算法（支持IP和用户两种模式）
- ✅ **CORS防护** - 可配置的跨域资源共享
- ✅ **安全头部** - XSS、点击劫持等防护
- ✅ **密码加密** - bcrypt加密存储

### ⚡ 高性能组件

- ✅ **Redis缓存** - 完整的缓存支持
- ✅ **Redis Stream** - 消息队列（生产者/消费者/自动清理）
- ✅ **连接池** - 数据库连接池优化
- ✅ **日志系统** - 分级日志+自动分割+轮转

### 🛠️ 开发友好

- ✅ **标准化响应** - 统一的JSON响应格式
- ✅ **错误监控** - Sentry集成（可选）
- ✅ **热重载支持** - 开发环境自动重启
- ✅ **Make命令** - 简化常用操作

---

## 🚀 5分钟快速开始

### 前置条件

确保已安装以下软件：
- **Go 1.23+**
- **PostgreSQL 12+**
- **Redis 6+**

### 第一步：创建项目

```bash
# 1. 复制框架
git clone https://github.com/your-repo/go-one.git my-project
cd my-project

# 2. 修改 go.mod 第一行
# module go-one → module my-project

# 3. 批量替换导入路径
# Linux/Mac:
find . -type f -name "*.go" -exec sed -i 's/go-one/my-project/g' {} +

# Windows PowerShell:
Get-ChildItem -Recurse -Filter *.go | ForEach-Object { 
    (Get-Content $_.FullName) -replace 'go-one', 'my-project' | 
    Set-Content $_.FullName 
}

# 4. 安装依赖
go mod download && go mod tidy
```

### 第二步：配置环境

```bash
# 1. 创建配置文件
cp env.example .env

# 2. 编辑 .env，修改以下配置：
# DB_PASSWORD=你的数据库密码
# DB_NAME=my_project_db
# JWT_SECRET=生产环境必须修改！
```

### 第三步：初始化数据库

```bash
# 创建数据库
psql -U postgres -c "CREATE DATABASE my_project_db;"

# 编辑 cmd/server/main.go，在 conf.Init() 后添加：
# model.DB.AutoMigrate(&model.User{})
```

### 第四步：启动服务

```bash
# 启动
go run cmd/server/main.go

# 或使用 Make
make dev
```

### 第五步：测试API

```bash
# 健康检查
curl http://localhost:8080/api/v1/ping

# 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"admin123"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

🎉 **成功！** 你的服务已经运行在 http://localhost:8080

---

## 🏗️ 项目架构

### 目录结构

```
go-one/
├── cmd/
│   └── server/           # 应用入口
│       └── main.go
├── internal/
│   ├── api/              # API层（HTTP处理器）
│   │   ├── context_helper.go   # BusinessContext适配器
│   │   ├── handler.go          # Handler基础
│   │   └── user.go             # 用户API
│   ├── service/          # Service层（业务逻辑）
│   │   ├── context.go          # BusinessContext定义
│   │   ├── errors.go           # ServiceError定义
│   │   ├── jwt.go              # JWT工具
│   │   ├── user_service.go     # 用户服务
│   │   └── service_manager.go  # 服务管理器
│   ├── repository/       # Repository层（数据访问）
│   │   └── user_repository.go
│   ├── model/            # 数据模型
│   │   ├── init.go
│   │   └── user.go
│   ├── middleware/       # 中间件
│   │   ├── jwt.go              # JWT认证
│   │   ├── ratelimit.go        # 限流
│   │   ├── cors.go             # CORS
│   │   └── security.go         # 安全头部
│   ├── cache/            # 缓存（Redis）
│   │   └── redis.go
│   ├── conf/             # 配置管理
│   │   └── conf.go
│   ├── serializer/       # 响应序列化
│   │   └── common.go
│   └── server/           # 路由配置
│       └── router.go
├── util/                 # 工具函数
│   ├── logger.go         # 日志
│   └── helpers.go        # 辅助函数
├── logs/                 # 日志文件目录
├── env.example           # 环境变量示例
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 分层架构

```
┌─────────────────────────────────────────┐
│   HTTP Layer (Gin Router & Middleware)  │  ← 处理HTTP请求
├─────────────────────────────────────────┤
│   API Layer (Handler)                   │  ← 参数绑定、响应序列化
│   - 转换 Request → DTO                   │
│   - 转换 ServiceError → HTTP Response   │
├─────────────────────────────────────────┤
│   BusinessContext (适配器)               │  ← 解耦HTTP与业务层
│   - 提取用户信息、请求元数据             │
├─────────────────────────────────────────┤
│   Service Layer (业务逻辑)               │  ← 框架无关
│   - 接收 BusinessContext 和 DTO         │
│   - 返回 ServiceError                   │
├─────────────────────────────────────────┤
│   Repository Layer (数据访问)            │  ← 数据库操作
├─────────────────────────────────────────┤
│   Model Layer (数据模型)                 │  ← GORM模型
└─────────────────────────────────────────┘
```

### 核心设计：Service层解耦

#### 1. BusinessContext - 业务上下文

替代 `gin.Context`，让Service层独立于HTTP框架：

```go
type BusinessContext struct {
    Context context.Context  // Go标准上下文
    
    // 用户信息
    UserUUID string
    Claims   *JWTClaims
    Account  *model.User
    
    // 请求元数据
    RequestID   string
    ClientIP    string
    UserAgent   string
    RequestTime int64
}
```

#### 2. ServiceError - 统一错误处理

```go
type ServiceError interface {
    error
    GetCode() int
    GetMessage() string
}

// 错误类型
type ValidationError   // 参数验证错误 (40000)
type AuthError         // 认证错误 (40001)
type NotFoundError     // 资源未找到 (40004)
type BusinessError     // 业务逻辑错误 (40xxx)
type DatabaseError     // 数据库错误 (50001)
```

#### 3. 完整的请求流程

```go
// API层：处理HTTP请求
func (h *Handler) UserRegister(c *gin.Context) {
    // 1. 获取BusinessContext
    bizCtx := GetBusinessContext(c)
    
    // 2. 绑定请求参数
    var req RegisterRequest
    c.ShouldBindJSON(&req)
    
    // 3. 转换为DTO
    dto := &service.RegisterDTO{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
    }
    
    // 4. 调用Service（传入BusinessContext）
    result, serviceErr := userService.Register(bizCtx, dto)
    if serviceErr != nil {
        HandleServiceError(c, serviceErr)  // 转换为HTTP响应
        return
    }
    
    // 5. 返回成功响应
    ResponseWithMessage(c, "注册成功", result)
}

// Service层：纯业务逻辑
func (s *UserService) Register(ctx *BusinessContext, dto *RegisterDTO) (*RegisterResult, ServiceError) {
    // 参数验证
    if len(dto.Username) < 3 {
        return nil, &ValidationError{
            Message: "用户名长度至少为3个字符",
            Code:    40000,
        }
    }
    
    // 业务逻辑
    user := &model.User{...}
    if err := s.userRepo.Create(user); err != nil {
        return nil, &DatabaseError{
            Message: "创建用户失败",
            Err:     err,
        }
    }
    
    return &RegisterResult{User: user, Token: token}, nil
}
```

**优势**：
- ✅ Service层完全独立，可在任何环境使用（HTTP、gRPC、CLI、消息队列）
- ✅ 易于编写单元测试，无需模拟HTTP上下文
- ✅ 清晰的错误处理流程
- ✅ 类型安全，代码可维护性高

---

## 🔨 开发指南

### 添加新模块（以文章为例）

#### 1. 创建模型 `internal/model/article.go`

```go
package model

type Article struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Title     string    `gorm:"size:200;not null" json:"title"`
    Content   string    `gorm:"type:text" json:"content"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### 2. 创建Repository `internal/repository/article_repository.go`

```go
package repository

type ArticleRepository interface {
    Create(article *model.Article) error
    FindByID(id uint) (*model.Article, error)
    List(page, pageSize int) ([]model.Article, int64, error)
}

type articleRepository struct {
    db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
    return &articleRepository{db: db}
}

func (r *articleRepository) Create(article *model.Article) error {
    return r.db.Create(article).Error
}
```

#### 3. 创建Service `internal/service/article_service.go`

```go
package service

// 创建文章DTO
type CreateArticleDTO struct {
    Title   string
    Content string
}

// 文章服务
type ArticleService struct {
    articleRepo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) *ArticleService {
    return &ArticleService{articleRepo: repo}
}

// 创建文章
func (s *ArticleService) CreateArticle(ctx *BusinessContext, dto *CreateArticleDTO) (*model.Article, ServiceError) {
    // 验证参数
    if dto.Title == "" {
        return nil, &ValidationError{Message: "标题不能为空", Code: 40000}
    }
    
    // 检查认证
    account, err := ctx.GetRequiredAccount()
    if err != nil {
        return nil, err
    }
    
    // 创建文章
    article := &model.Article{
        Title:   dto.Title,
        Content: dto.Content,
        UserID:  account.ID,
    }
    
    if err := s.articleRepo.Create(article); err != nil {
        return nil, &DatabaseError{Message: "创建文章失败", Err: err}
    }
    
    return article, nil
}
```

#### 4. 更新ServiceManager `internal/service/service_manager.go`

```go
func NewServiceManager(db *gorm.DB) *ServiceManager {
    userRepo := repository.NewUserRepository(db)
    articleRepo := repository.NewArticleRepository(db)  // 新增
    
    return &ServiceManager{
        userRepo:    userRepo,
        articleRepo: articleRepo,  // 新增
    }
}

func (sm *ServiceManager) NewArticleService() *ArticleService {
    return NewArticleService(sm.articleRepo)
}
```

#### 5. 创建API Handler `internal/api/article.go`

```go
package api

type CreateArticleRequest struct {
    Title   string `json:"title" binding:"required,max=200"`
    Content string `json:"content" binding:"required"`
}

func (h *Handler) CreateArticle(c *gin.Context) {
    // 1. 获取BusinessContext
    bizCtx := GetBusinessContext(c)
    
    // 2. 绑定请求
    var req CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, serializer.ParamErr("参数错误", err))
        return
    }
    
    // 3. 转换DTO
    dto := &service.CreateArticleDTO{
        Title:   req.Title,
        Content: req.Content,
    }
    
    // 4. 调用Service
    articleService := h.serviceManager.NewArticleService()
    article, serviceErr := articleService.CreateArticle(bizCtx, dto)
    if serviceErr != nil {
        HandleServiceError(c, serviceErr)
        return
    }
    
    // 5. 返回响应
    ResponseWithMessage(c, "创建成功", article)
}
```

#### 6. 添加路由 `internal/server/router.go`

```go
protected := v1.Group("")
protected.Use(middleware.JWTMiddleware())
{
    // 文章路由
    article := protected.Group("/article")
    {
        article.POST("", h.CreateArticle)
        article.GET("/:id", h.GetArticle)
        article.PUT("/:id", h.UpdateArticle)
        article.DELETE("/:id", h.DeleteArticle)
    }
}
```

### 中间件使用

```go
// JWT认证（必需）
protected.Use(middleware.JWTMiddleware())

// IP限流：10秒内最多6次
auth.Use(middleware.RateLimitMiddleware(6, 10*time.Second, "ip"))

// 用户限流：1分钟内最多60次
protected.Use(middleware.RateLimitMiddleware(60, 1*time.Minute, "user"))

// CORS（已全局配置）
r.Use(middleware.CORSMiddleware())

// 安全头部（已全局配置）
r.Use(middleware.SecurityMiddleware())
```

### 日志使用

```go
import "my-project/util"

// 不同级别的日志
util.Log().Debug("调试信息: %v", data)
util.Log().Info("操作成功: 用户ID=%d", userID)
util.Log().Warning("警告: %s", message)
util.Log().Error("错误: %v", err)
```

### Redis缓存使用

```go
import "my-project/internal/cache"

ctx := context.Background()

// 设置缓存
cache.RedisClient.Set(ctx, "key", "value", 10*time.Minute)

// 获取缓存
val, err := cache.RedisClient.Get(ctx, "key").Result()

// 删除缓存
cache.RedisClient.Del(ctx, "key")

// Hash操作
cache.RedisClient.HSet(ctx, "user:1", "name", "张三")
cache.RedisClient.HGet(ctx, "user:1", "name")
```

---

## 📡 API文档

### 认证相关

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "admin",
  "email": "admin@example.com",
  "password": "admin123"
}
```

**响应**：
```json
{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "nickname": "admin",
      "created_at": "2025-10-04T12:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

### 用户相关（需要认证）

所有以下接口需要在Header中携带Token：
```http
Authorization: Bearer {your_token}
```

#### 获取用户资料
```http
GET /api/v1/user/profile
```

#### 更新用户资料
```http
PUT /api/v1/user/profile
Content-Type: application/json

{
  "nickname": "新昵称",
  "avatar": "https://example.com/avatar.jpg"
}
```

#### 修改密码
```http
PUT /api/v1/user/password
Content-Type: application/json

{
  "old_password": "old123",
  "new_password": "new456"
}
```

#### 用户列表
```http
GET /api/v1/users?page=1&page_size=20
```

### 响应格式

**成功响应**：
```json
{
  "code": 0,
  "msg": "操作成功",
  "data": { ... }
}
```

**错误响应**：
```json
{
  "code": 40000,
  "msg": "参数验证失败",
  "error": "用户名长度至少为3个字符"
}
```

**错误码说明**：
- `40000` - 参数验证错误
- `40001` - 认证失败
- `40003` - 权限不足
- `40004` - 资源未找到
- `40009` - 资源冲突（如用户名已存在）
- `429` - 请求过于频繁
- `50000+` - 服务器内部错误

---

## 🔧 配置说明

### 环境变量 (`.env`)

```env
# ========== 服务器配置 ==========
GIN_MODE=debug                    # debug/release
SERVER_PORT=8080
SERVER_TIMEZONE=Asia/Shanghai

# ========== 数据库配置 ==========
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password         # ⚠️ 必须修改
DB_NAME=go_one_db
DB_TIMEZONE=Asia/Shanghai

# ========== Redis配置 ==========
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# ========== JWT配置 ==========
JWT_SECRET=your_secret_key        # ⚠️ 生产环境必须修改！
JWT_ACCESS_TOKEN_EXPIRE=3600      # 秒（1小时）
JWT_REFRESH_TOKEN_EXPIRE=604800   # 秒（7天）

# ========== 日志配置 ==========
LOG_LEVEL=debug                   # debug/info/warning/error
LOG_FILE=./logs/app.log
LOG_MAX_SIZE_MB=100               # 单个日志文件最大大小
LOG_MAX_BACKUPS=7                 # 保留的旧日志文件数
LOG_MAX_AGE_DAYS=30               # 日志保留天数
LOG_COMPRESS=true                 # 是否压缩旧日志
LOG_CONSOLE=true                  # 是否同时输出到控制台

# ========== Sentry配置（可选）==========
SENTRY_DSN=
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=1.0.0
SENTRY_TRACES_SAMPLE_RATE=0.1
SENTRY_PROFILES_SAMPLE_RATE=0.1
```

---

## 🚀 部署指南

### Docker部署

#### 1. 创建Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# 下载依赖
RUN go mod download

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 复制编译好的二进制文件
COPY --from=builder /app/server .
COPY --from=builder /app/.env .

# 创建日志目录
RUN mkdir -p logs

EXPOSE 8080

CMD ["./server"]
```

#### 2. 构建和运行

```bash
# 构建镜像
docker build -t my-app:latest .

# 运行容器
docker run -d \
  --name my-app \
  -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e REDIS_ADDR=host.docker.internal:6379 \
  my-app:latest

# 查看日志
docker logs -f my-app
```

#### 3. Docker Compose部署

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=go_one_db
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

### 二进制部署

```bash
# 1. 编译
make build
# 或
go build -o bin/server cmd/server/main.go

# 2. 创建systemd服务文件 /etc/systemd/system/myapp.service
[Unit]
Description=My Go App
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/myapp
ExecStart=/opt/myapp/bin/server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target

# 3. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp
sudo systemctl status myapp
```

### 使用Nginx反向代理

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

---

## ❓ 常见问题

### 如何切换数据库？

**切换到MySQL：**
```go
// 1. 修改 go.mod
require gorm.io/driver/mysql v1.5.0

// 2. 修改 internal/model/init.go
import "gorm.io/driver/mysql"

dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

**切换到SQLite：**
```go
require gorm.io/driver/sqlite v1.5.0

import "gorm.io/driver/sqlite"

db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
```

### Service层可以调用其他Service吗？

可以，通过ServiceManager获取：

```go
type OrderService struct {
    orderRepo repository.OrderRepository
    userService *UserService  // 注入其他Service
}

// 在ServiceManager中配置依赖
func (sm *ServiceManager) NewOrderService() *OrderService {
    return &OrderService{
        orderRepo:   sm.orderRepo,
        userService: sm.NewUserService(),
    }
}
```

### 如何处理数据库事务？

在Service层使用GORM的Transaction：

```go
func (s *Service) ComplexOperation(ctx *BusinessContext) ServiceError {
    err := model.DB.Transaction(func(tx *gorm.DB) error {
        // 操作1
        if err := tx.Create(&user).Error; err != nil {
            return err
        }
        
        // 操作2
        if err := tx.Create(&profile).Error; err != nil {
            return err
        }
        
        return nil
    })
    
    if err != nil {
        return &DatabaseError{Message: "事务失败", Err: err}
    }
    return nil
}
```

### 如何编写单元测试？

Service层测试示例：

```go
func TestUserService_Register(t *testing.T) {
    // 1. 创建mock repository
    mockRepo := &MockUserRepository{}
    
    // 2. 创建service
    service := NewUserService(mockRepo)
    
    // 3. 创建BusinessContext
    ctx := NewBusinessContext(context.Background())
    
    // 4. 调用service
    dto := &RegisterDTO{
        Username: "test",
        Email:    "test@example.com",
        Password: "test123",
    }
    
    result, err := service.Register(ctx, dto)
    
    // 5. 断言
    assert.Nil(t, err)
    assert.NotNil(t, result.User)
    assert.NotEmpty(t, result.Token)
}
```

### 如何添加自定义中间件？

```go
// 1. 创建中间件 internal/middleware/custom.go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        startTime := time.Now()
        
        // 继续处理请求
        c.Next()
        
        // 后置处理
        duration := time.Since(startTime)
        util.Log().Info("请求耗时: %v", duration)
    }
}

// 2. 在router中使用
r.Use(CustomMiddleware())
```

---

## 🛡️ 安全建议

1. ⚠️ **生产环境必须修改 `JWT_SECRET`**
2. 🔒 使用强密码策略（密码长度、复杂度要求）
3. 🔐 启用 HTTPS（使用Let's Encrypt免费证书）
4. 🌐 配置正确的CORS域名白名单
5. 📊 启用Sentry错误监控
6. 🔄 定期更新依赖包 `go get -u && go mod tidy`
7. 🚫 不要在日志中记录敏感信息
8. 🔑 使用环境变量管理敏感配置
9. 🛡️ 启用限流保护API
10. 📝 定期审计日志

---

## 🛠️ Make命令

```bash
make help      # 查看所有命令
make deps      # 安装依赖
make build     # 编译项目（输出到bin/）
make run       # 运行编译后的二进制
make dev       # 开发模式（go run）
make test      # 运行测试
make lint      # 代码检查
make fmt       # 格式化代码
make clean     # 清理编译文件
```

---

## 📚 推荐阅读

- [Gin框架文档](https://gin-gonic.com/zh-cn/docs/)
- [GORM文档](https://gorm.io/zh_CN/docs/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go语言标准库](https://pkg.go.dev/std)
- [SOLID原则](https://en.wikipedia.org/wiki/SOLID)
- [12-Factor App](https://12factor.net/)

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

在提交PR前，请确保：
- [ ] 代码通过 `make test`
- [ ] 代码通过 `make lint`
- [ ] 代码已格式化 `make fmt`
- [ ] 添加了必要的测试
- [ ] 更新了相关文档

---

## 📮 联系与支持

- 📧 提交Issue：[GitHub Issues](https://github.com/your-repo/go-one/issues)
- 💬 讨论区：[GitHub Discussions](https://github.com/your-repo/go-one/discussions)

---

**Happy Coding! 🎉**

*Built with ❤️ using Go*
