# Go-One 后端开发框架

> 一个从生产环境项目中提取的现代化 Go 后端开发脚手架

[English](README.md) | 简体中文

## 📦 框架简介

Go-One 是一个基于 Gin 框架的企业级后端开发脚手架，采用清晰的分层架构和依赖注入模式，开箱即用，帮助你快速构建高质量的 Go 后端应用。

### 核心特性

✅ **分层架构** - API层 → Service层 → Repository层 → Model层  
✅ **依赖注入** - 使用 ServiceManager 管理服务依赖，易于测试和维护  
✅ **JWT认证** - 内置完整的用户认证系统  
✅ **智能限流** - 基于令牌桶算法，支持IP和用户两种限流模式  
✅ **安全防护** - CORS、安全头部、XSS防护等  
✅ **日志系统** - 基于 lumberjack 的日志分割和轮转  
✅ **数据库** - GORM + PostgreSQL，易于扩展  
✅ **缓存支持** - Redis 集成  
✅ **Redis Stream** - 完整的消息队列支持（生产者/消费者/自动清理）  
✅ **错误监控** - Sentry 集成（可选）  
✅ **标准化** - 统一的 API 响应格式和错误处理

## 🚀 5分钟快速开始

### 1. 准备环境

确保已安装：
- Go 1.23+
- PostgreSQL 12+
- Redis 6+

### 2. 创建项目

```bash
# 复制框架
cp -r go-one my-project
cd my-project

# 修改模块名（go.mod 第一行）
# 从: module go-one
# 改为: module my-project

# 批量替换导入路径
# Linux/Mac:
find . -type f -name "*.go" -exec sed -i 's/go-one/my-project/g' {} +

# Windows PowerShell:
Get-ChildItem -Recurse -Filter *.go | ForEach-Object { (Get-Content $_.FullName) -replace 'go-one', 'my-project' | Set-Content $_.FullName }

# 安装依赖
go mod download && go mod tidy
```

### 3. 配置环境

```bash
# 复制环境变量文件
cp env.example .env

# 编辑 .env，至少配置：
# - DB_PASSWORD=你的数据库密码
# - DB_NAME=my_project_db
# - JWT_SECRET=你的密钥（生产环境必须修改！）
```

### 4. 创建数据库

```bash
psql -U postgres -c "CREATE DATABASE my_project_db;"
```

### 5. 启动服务

```bash
# 编辑 cmd/server/main.go，在 conf.Init() 后添加：
# model.DB.AutoMigrate(&model.User{})

# 启动服务
make dev
# 或
go run cmd/server/main.go
```

### 6. 测试API

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

## 📁 项目结构

```
go-one/
├── cmd/server/           # 应用入口
├── internal/
│   ├── api/             # API处理器（Controller）
│   ├── service/         # 业务逻辑层
│   ├── repository/      # 数据访问层
│   ├── model/           # 数据模型
│   ├── middleware/      # 中间件（JWT、限流、CORS等）
│   ├── conf/            # 配置管理
│   ├── cache/           # 缓存（Redis）
│   ├── serializer/      # 响应序列化
│   └── server/          # 路由配置
├── util/                # 工具函数
└── logs/                # 日志文件
```

## 📚 详细文档

- 📖 [完整功能说明](README.md) - 英文完整文档
- 🏗️ [架构设计文档](ARCHITECTURE.md) - 深入理解框架设计
- 📡 [API接口文档](API_DOCUMENTATION.md) - 所有API说明
- ⚡ [快速开始指南](QUICKSTART.md) - 详细的入门教程
- 🚀 [Redis Stream使用指南](README_STREAM.md) - 消息队列完整教程

## 🔨 开发指南

### 添加新模块（以文章模块为例）

#### 1. 创建模型 (`internal/model/article.go`)

```go
type Article struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Title     string    `gorm:"size:200" json:"title"`
    Content   string    `gorm:"type:text" json:"content"`
    UserID    uint      `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### 2. 创建Repository (`internal/repository/article_repository.go`)

```go
type ArticleRepository interface {
    Create(article *model.Article) error
    FindByID(id uint) (*model.Article, error)
}

type articleRepository struct {
    db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
    return &articleRepository{db: db}
}
```

#### 3. 创建Service (`internal/service/article_service.go`)

```go
type ArticleService struct {
    articleRepo repository.ArticleRepository
}

func (s *ArticleService) CreateArticle(title, content string, userID uint) error {
    article := &model.Article{
        Title:   title,
        Content: content,
        UserID:  userID,
    }
    return s.articleRepo.Create(article)
}
```

#### 4. 更新ServiceManager

在 `service_manager.go` 中添加：
```go
articleRepo repository.ArticleRepository

func (sm *ServiceManager) NewArticleService() *ArticleService {
    return NewArticleService(sm.articleRepo)
}
```

#### 5. 创建API Handler (`internal/api/article.go`)

```go
func (h *Handler) CreateArticle(c *gin.Context) {
    var req CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, serializer.ParamErr("参数错误", err))
        return
    }
    
    articleService := h.serviceManager.NewArticleService()
    // ... 业务逻辑
    
    c.JSON(200, serializer.Success("创建成功", nil))
}
```

#### 6. 添加路由 (`internal/server/router.go`)

```go
article := protected.Group("/article")
{
    article.POST("", h.CreateArticle)
    article.GET("/:id", h.GetArticle)
}
```

### 中间件使用

```go
// JWT认证
protected.Use(middleware.JWTMiddleware())

// IP限流：10秒内最多6次
auth.Use(middleware.RateLimitMiddleware(6, 10*time.Second, "ip"))

// 用户限流：1分钟内最多60次
protected.Use(middleware.RateLimitMiddleware(60, 1*time.Minute, "user"))
```

### 日志使用

```go
import "my-project/util"

util.Log().Debug("调试信息: %v", data)
util.Log().Info("操作成功")
util.Log().Warning("警告信息")
util.Log().Error("错误信息: %v", err)
```

### Redis Stream 使用（可选）

框架提供完整的 Redis Stream 支持，适用于消息队列、异步任务等场景。

**启用Stream功能：**

在 `internal/conf/conf.go` 的 `Init()` 函数中添加：
```go
// 初始化Redis Stream组件（可选）
if err := cache.InitStreams(); err != nil {
    util.Log().Panic("初始化Stream组件失败: %v", err)
}
```

**快速示例：**

```go
import "my-project/internal/cache"

// 1. 创建生产者
cache.CreateSimpleStreamProducer("my_producer", "my_stream")

// 2. 发送消息
ctx := context.Background()
cache.AddMessage("my_producer", ctx, map[string]interface{}{
    "user_id": 123,
    "action":  "login",
})

// 3. 创建消费者处理器
type MyHandler struct{}

func (h *MyHandler) HandleMessage(ctx context.Context, msg redis.XMessage) error {
    util.Log().Info("收到消息: %v", msg.Values)
    return nil
}

// 4. 启动消费者
handler := &MyHandler{}
cache.InitConsumerWithHandler("my_consumer", handler, "my_stream")
```

**详细文档：**[Redis Stream 使用指南](README_STREAM.md)

## 🔧 配置说明

所有配置通过 `.env` 文件管理：

```env
# 服务器
GIN_MODE=debug              # debug/release
SERVER_PORT=8080
SERVER_TIMEZONE=Asia/Shanghai

# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=my_db

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your_secret_key
JWT_ACCESS_TOKEN_EXPIRE=3600       # 秒
JWT_REFRESH_TOKEN_EXPIRE=604800    # 秒

# 日志
LOG_LEVEL=debug            # debug/info/warning/error
LOG_FILE=./logs/app.log
LOG_MAX_SIZE_MB=100
LOG_MAX_BACKUPS=7
LOG_MAX_AGE_DAYS=30
```

## 🛡️ 安全建议

1. ⚠️ **生产环境必须修改 JWT_SECRET**
2. 🔒 使用强密码策略
3. 🔐 启用 HTTPS
4. 🌐 配置正确的 CORS 域名白名单
5. 📊 启用 Sentry 错误监控
6. 🔄 定期更新依赖包

## 📊 API响应格式

### 成功
```json
{
  "code": 0,
  "msg": "操作成功",
  "data": { ... }
}
```

### 错误
```json
{
  "code": 400,
  "msg": "错误描述",
  "error": "详细错误信息"
}
```

### 错误码
- `0` - 成功
- `400` - 请求参数错误
- `401` - 未授权
- `403` - 禁止访问
- `404` - 资源不存在
- `429` - 请求过于频繁
- `500` - 服务器内部错误

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/service/...

# 查看覆盖率
go test -cover ./...
```

## 📦 部署

### Docker 部署

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/.env .
EXPOSE 8080
CMD ["./server"]
```

```bash
docker build -t my-app .
docker run -p 8080:8080 my-app
```

### 二进制部署

```bash
# 编译
make build

# 运行
./bin/server
```

## 🔄 数据库迁移

### 方式1：GORM AutoMigrate（简单）

在 `cmd/server/main.go` 中：
```go
model.DB.AutoMigrate(
    &model.User{},
    &model.Article{},
    // ... 其他模型
)
```

### 方式2：golang-migrate（推荐生产环境）

```bash
# 安装工具
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移文件
migrate create -ext sql -dir migrations -seq create_users

# 运行迁移
migrate -database "postgresql://user:pass@localhost/db?sslmode=disable" -path migrations up
```

## 🛠️ Make 命令

```bash
make help      # 查看所有命令
make deps      # 安装依赖
make build     # 编译项目
make run       # 运行项目
make dev       # 开发模式（热重载）
make test      # 运行测试
make clean     # 清理编译文件
```

## 📖 推荐阅读

- [Gin框架文档](https://gin-gonic.com/zh-cn/docs/)
- [GORM文档](https://gorm.io/zh_CN/docs/)
- [Go语言标准库](https://pkg.go.dev/std)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## ❓ 常见问题

### 如何切换数据库？

**MySQL:**
```go
// go.mod 添加
require gorm.io/driver/mysql v1.5.0

// internal/model/init.go 修改
import "gorm.io/driver/mysql"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{...})
```

**SQLite:**
```go
require gorm.io/driver/sqlite v1.5.0

import "gorm.io/driver/sqlite"
db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{...})
```

### 如何添加更多中间件？

在 `internal/server/router.go` 中添加：
```go
r.Use(YourMiddleware())
```

### 如何自定义日志格式？

修改 `util/logger.go` 中的 `Println` 方法。

### Service层可以调用其他Service吗？

可以，通过 ServiceManager 获取其他服务实例。

---

## 📮 联系方式

如有问题，请提交 Issue。

**Happy Coding! 🎉**

