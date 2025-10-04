# Go-One 后端开发框架

一个基于 Gin 框架的现代化 Go 后端开发脚手架，提供完整的项目结构、中间件、依赖注入等开箱即用的功能。

## ✨ 特性

- 🏗️ **清晰的项目结构** - 采用分层架构设计（API层、Service层、Repository层、Model层）
- 💉 **依赖注入** - 使用 ServiceManager 管理服务依赖
- 🔐 **JWT认证** - 内置 JWT 认证中间件
- 🚦 **限流保护** - 基于令牌桶算法的灵活限流中间件
- 🛡️ **安全防护** - CORS、安全头部、XSS防护等
- 📝 **日志管理** - 基于 lumberjack 的日志分割和轮转
- 💾 **数据库支持** - GORM + PostgreSQL，易于扩展其他数据库
- 🔄 **Redis缓存** - 集成 Redis 支持
- 📊 **Sentry监控** - 可选的错误监控和性能追踪
- 🎯 **标准化响应** - 统一的 API 响应格式

## 📁 项目结构

```
go-one/
├── cmd/
│   └── server/
│       └── main.go           # 应用入口
├── internal/
│   ├── api/                  # API处理器层
│   │   ├── handler.go        # Handler 聚合器
│   │   └── user.go           # 用户相关API
│   ├── cache/                # 缓存层
│   │   └── redis.go          # Redis客户端
│   ├── conf/                 # 配置管理
│   │   ├── conf.go           # 配置初始化
│   │   └── jwt.go            # JWT配置
│   ├── middleware/           # 中间件
│   │   ├── cors.go           # CORS中间件
│   │   ├── jwt.go            # JWT认证中间件
│   │   ├── ratelimit.go      # 限流中间件
│   │   └── security.go       # 安全头部中间件
│   ├── model/                # 数据模型
│   │   ├── init.go           # 数据库初始化
│   │   └── user.go           # 用户模型
│   ├── repository/           # 数据访问层
│   │   └── user_repository.go
│   ├── serializer/           # 序列化器
│   │   └── common.go         # 通用响应格式
│   ├── server/               # 服务器配置
│   │   └── router.go         # 路由配置
│   └── service/              # 业务逻辑层
│       ├── service_manager.go
│       └── user_service.go
├── util/                     # 工具包
│   ├── helpers.go            # 辅助函数
│   ├── jwt.go                # JWT工具
│   └── logger.go             # 日志工具
├── logs/                     # 日志文件目录
├── .gitignore
├── env.example               # 环境变量示例
├── go.mod
├── Makefile
└── README.md
```

## 🚀 快速开始

### 前置要求

- Go 1.23+
- PostgreSQL 12+
- Redis 6+

### 安装步骤

1. **克隆或复制框架**

```bash
# 复制 go-one 文件夹到你的项目目录
cp -r go-one my-project
cd my-project
```

2. **修改模块名称**

修改 `go.mod` 中的模块名：
```go
module my-project  // 改为你的项目名
```

批量替换代码中的导入路径：
```bash
# Linux/Mac
find . -type f -name "*.go" -exec sed -i 's/go-one/my-project/g' {} +

# Windows PowerShell
Get-ChildItem -Recurse -Filter *.go | ForEach-Object { (Get-Content $_.FullName) -replace 'go-one', 'my-project' | Set-Content $_.FullName }
```

3. **安装依赖**

```bash
make deps
# 或
go mod download
go mod tidy
```

4. **配置环境变量**

```bash
cp env.example .env
# 编辑 .env 文件，填入你的配置
```

5. **创建数据库**

```sql
CREATE DATABASE go_one_db;
```

6. **运行数据库迁移**

在 `cmd/server/main.go` 的 `conf.Init()` 之后添加：
```go
// 自动迁移
model.DB.AutoMigrate(&model.User{})
```

7. **启动服务**

```bash
make dev
# 或
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

### 测试API

```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"123456"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'

# 获取用户资料（需要token）
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 健康检查
curl http://localhost:8080/api/v1/ping
```

## 📝 开发指南

### 添加新模块

假设要添加一个"文章"模块：

1. **创建模型** (`internal/model/article.go`)

```go
package model

import "time"

type Article struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Title     string    `gorm:"size:200;not null" json:"title"`
    Content   string    `gorm:"type:text" json:"content"`
    UserID    uint      `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (Article) TableName() string {
    return "articles"
}
```

2. **创建Repository** (`internal/repository/article_repository.go`)

```go
package repository

import (
    "go-one/internal/model"
    "gorm.io/gorm"
)

type ArticleRepository interface {
    Create(article *model.Article) error
    FindByID(id uint) (*model.Article, error)
    // ... 其他方法
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

func (r *articleRepository) FindByID(id uint) (*model.Article, error) {
    var article model.Article
    err := r.db.Where("id = ?", id).First(&article).Error
    return &article, err
}
```

3. **创建Service** (`internal/service/article_service.go`)

```go
package service

import (
    "go-one/internal/model"
    "go-one/internal/repository"
)

type ArticleService struct {
    articleRepo repository.ArticleRepository
}

func NewArticleService(articleRepo repository.ArticleRepository) *ArticleService {
    return &ArticleService{articleRepo: articleRepo}
}

func (s *ArticleService) CreateArticle(title, content string, userID uint) (*model.Article, error) {
    article := &model.Article{
        Title:   title,
        Content: content,
        UserID:  userID,
    }
    err := s.articleRepo.Create(article)
    return article, err
}
```

4. **更新ServiceManager** (`internal/service/service_manager.go`)

```go
type ServiceManager struct {
    userRepo    repository.UserRepository
    articleRepo repository.ArticleRepository  // 添加
}

func NewServiceManager(db *gorm.DB) *ServiceManager {
    return &ServiceManager{
        userRepo:    repository.NewUserRepository(db),
        articleRepo: repository.NewArticleRepository(db),  // 添加
    }
}

func (sm *ServiceManager) NewArticleService() *ArticleService {
    return NewArticleService(sm.articleRepo)
}
```

5. **创建API Handler** (`internal/api/article.go`)

```go
package api

import (
    "go-one/internal/serializer"
    "net/http"
    "github.com/gin-gonic/gin"
)

type CreateArticleRequest struct {
    Title   string `json:"title" binding:"required"`
    Content string `json:"content" binding:"required"`
}

func (h *Handler) CreateArticle(c *gin.Context) {
    var req CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, serializer.ParamErr("参数错误", err))
        return
    }

    userID := c.GetString("userID")
    // ... 业务逻辑
    
    c.JSON(http.StatusOK, serializer.Success("创建成功", nil))
}
```

6. **添加路由** (`internal/server/router.go`)

```go
protected := v1.Group("")
protected.Use(middleware.JWTMiddleware())
{
    article := protected.Group("/article")
    {
        article.POST("", h.CreateArticle)
        article.GET("/:id", h.GetArticle)
    }
}
```

### 中间件使用

**JWT认证中间件**
```go
protected.Use(middleware.JWTMiddleware())
```

**限流中间件**
```go
// 基于IP限流：每10秒最多6次请求
auth.Use(middleware.RateLimitMiddleware(6, 10*time.Second, "ip"))

// 基于用户限流：每1分钟最多60次请求
protected.Use(middleware.RateLimitMiddleware(60, 1*time.Minute, "user"))
```

### 配置管理

所有配置通过环境变量管理，在 `.env` 文件中配置：

```env
# 服务器配置
GIN_MODE=debug
SERVER_PORT=8080
SERVER_TIMEZONE=Asia/Shanghai

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=go_one_db

# Redis配置
REDIS_ADDR=localhost:6379

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_ACCESS_TOKEN_EXPIRE=3600
```

### 日志使用

```go
import "go-one/util"

// 不同级别的日志
util.Log().Debug("调试信息: %s", value)
util.Log().Info("普通信息: %d", count)
util.Log().Warning("警告信息")
util.Log().Error("错误信息: %v", err)
util.Log().Panic("严重错误，程序将退出")
```

## 🔧 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| GIN_MODE | Gin运行模式 (debug/release) | debug |
| SERVER_PORT | 服务器端口 | 8080 |
| SERVER_TIMEZONE | 服务器时区 | Asia/Shanghai |
| DB_HOST | 数据库地址 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | postgres |
| DB_PASSWORD | 数据库密码 | - |
| DB_NAME | 数据库名称 | go_one_db |
| REDIS_ADDR | Redis地址 | localhost:6379 |
| REDIS_PASSWORD | Redis密码 | - |
| REDIS_DB | Redis数据库编号 | 0 |
| JWT_SECRET | JWT密钥 | - |
| JWT_ACCESS_TOKEN_EXPIRE | 访问令牌过期时间（秒） | 3600 |
| JWT_REFRESH_TOKEN_EXPIRE | 刷新令牌过期时间（秒） | 604800 |
| LOG_LEVEL | 日志级别 (debug/info/warning/error) | debug |
| LOG_FILE | 日志文件路径 | ./logs/app.log |

### 数据库连接

框架使用 GORM 作为 ORM，默认支持 PostgreSQL。要使用其他数据库：

**MySQL:**
```go
// go.mod
require gorm.io/driver/mysql v1.5.0

// internal/model/init.go
import "gorm.io/driver/mysql"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{...})
```

**SQLite:**
```go
// go.mod
require gorm.io/driver/sqlite v1.5.0

// internal/model/init.go
import "gorm.io/driver/sqlite"
db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{...})
```

## 🛡️ 安全最佳实践

1. **生产环境必须修改 JWT_SECRET**
2. **使用强密码策略**
3. **启用 HTTPS**（修改 middleware/security.go）
4. **配置正确的 CORS 域名**（修改 middleware/cors.go）
5. **定期更新依赖包**
6. **启用 Sentry 监控生产环境错误**

## 📊 API响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "id": 1,
    "username": "testuser"
  }
}
```

### 错误响应
```json
{
  "code": 400,
  "msg": "参数错误",
  "error": "详细错误信息"
}
```

### 错误代码
- `0` - 成功
- `400` - 请求参数错误
- `401` - 未授权
- `403` - 禁止访问
- `404` - 资源不存在
- `422` - 验证错误
- `429` - 请求过于频繁
- `500` - 服务器内部错误

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/service/...

# 运行测试并生成覆盖率报告
go test -cover ./...
```

## 📦 部署

### Docker部署

创建 `Dockerfile`:
```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./server"]
```

构建和运行:
```bash
docker build -t my-app .
docker run -p 8080:8080 --env-file .env my-app
```

### 系统服务部署

创建 systemd 服务文件 `/etc/systemd/system/myapp.service`:
```ini
[Unit]
Description=My Go App
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/myapp
ExecStart=/var/www/myapp/server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启动服务:
```bash
sudo systemctl daemon-reload
sudo systemctl start myapp
sudo systemctl enable myapp
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 📮 联系方式

如有问题，请提交 Issue 或联系维护者。

---

**Happy Coding! 🎉**

