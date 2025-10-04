# 快速开始指南

这个指南将帮助你在5分钟内启动并运行 Go-One 框架。

## 前置条件检查

在开始之前，请确保已安装：

```bash
# 检查 Go 版本 (需要 1.23+)
go version

# 检查 PostgreSQL (需要 12+)
psql --version

# 检查 Redis (需要 6+)
redis-cli --version
```

如果未安装，请先安装这些依赖。

## 第一步：项目设置

### 1. 复制框架到你的项目

```bash
# 复制 go-one 文件夹
cp -r go-one my-awesome-project
cd my-awesome-project
```

### 2. 修改模块名称

**编辑 `go.mod`**，将第一行改为你的项目名：
```go
module my-awesome-project
```

**批量替换导入路径**：

**Linux/Mac:**
```bash
find . -type f -name "*.go" -exec sed -i 's/go-one/my-awesome-project/g' {} +
```

**Windows PowerShell:**
```powershell
Get-ChildItem -Recurse -Filter *.go | ForEach-Object { 
    (Get-Content $_.FullName) -replace 'go-one', 'my-awesome-project' | 
    Set-Content $_.FullName 
}
```

### 3. 安装依赖

```bash
go mod download
go mod tidy
```

## 第二步：配置环境

### 1. 创建环境变量文件

```bash
cp env.example .env
```

### 2. 编辑 `.env` 文件

**最小配置（开发环境）：**

```env
# 服务器配置
GIN_MODE=debug
SERVER_PORT=8080
SERVER_TIMEZONE=Asia/Shanghai

# 数据库配置
POSTGRES_URL=postgresql://
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=my_awesome_db
DB_TIMEZONE=Asia/Shanghai

# Redis配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置（生产环境请修改！）
JWT_SECRET=your_super_secret_key_change_this_in_production
JWT_ACCESS_TOKEN_EXPIRE=3600
JWT_REFRESH_TOKEN_EXPIRE=604800

# 日志配置
LOG_LEVEL=debug
LOG_FILE=./logs/app.log
LOG_CONSOLE=true
```

## 第三步：数据库设置

### 1. 创建数据库

```bash
# 连接到 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE my_awesome_db;

# 退出
\q
```

### 2. 运行数据库迁移

编辑 `cmd/server/main.go`，在 `conf.Init()` 之后添加：

```go
// 自动迁移数据库表
model.DB.AutoMigrate(&model.User{})
```

## 第四步：启动服务

### 开发模式启动

```bash
make dev
# 或
go run cmd/server/main.go
```

你应该看到：
```
[Info] 2025-10-04 12:00:00 | conf.go:70 | Redis连接成功
[Info] 2025-10-04 12:00:00 | init.go:72 | 数据库连接成功
[Info] 2025-10-04 12:00:00 | jwt.go:42 | JWT配置初始化完成
[Info] 2025-10-04 12:00:00 | main.go:48 | 启动服务器在 :8080
```

## 第五步：测试API

### 1. 健康检查

```bash
curl http://localhost:8080/api/v1/ping
```

**预期响应：**
```json
{
  "code": 0,
  "msg": "pong",
  "data": null
}
```

### 2. 注册用户

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123"
  }'
```

**预期响应：**
```json
{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      ...
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**保存返回的 token！** 你需要它来访问受保护的接口。

### 3. 获取用户资料

```bash
# 替换 YOUR_TOKEN 为上一步获取的 token
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. 更新用户资料

```bash
curl -X PUT http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "超级管理员",
    "avatar": "https://example.com/avatar.jpg"
  }'
```

## 🎉 成功！

恭喜！你已经成功运行了 Go-One 框架。

## 下一步

### 1. 查看完整文档

- 📖 [README.md](README.md) - 完整功能说明
- 🏗️ [ARCHITECTURE.md](ARCHITECTURE.md) - 架构设计文档
- 📡 [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - API接口文档

### 2. 添加你的业务逻辑

参考 README.md 中的"添加新模块"章节，开始构建你的应用。

### 3. 自定义配置

根据你的需求调整：
- 中间件配置
- CORS域名白名单
- 限流策略
- 日志级别

## 常见问题排查

### 问题1：无法连接数据库

**错误信息：**
```
postgres连接失败: connection refused
```

**解决方案：**
1. 确认 PostgreSQL 已启动
2. 检查 `.env` 中的数据库配置
3. 确认数据库已创建

```bash
# 启动 PostgreSQL (根据系统不同)
# Linux
sudo systemctl start postgresql

# Mac
brew services start postgresql

# Windows
# 在服务管理器中启动 PostgreSQL 服务
```

### 问题2：无法连接 Redis

**错误信息：**
```
Redis连接失败: connection refused
```

**解决方案：**
```bash
# 启动 Redis (根据系统不同)
# Linux
sudo systemctl start redis

# Mac
brew services start redis

# Windows
# 下载并启动 Redis for Windows
```

### 问题3：端口已被占用

**错误信息：**
```
bind: address already in use
```

**解决方案：**
修改 `.env` 中的 `SERVER_PORT` 为其他端口：
```env
SERVER_PORT=8081
```

### 问题4：导入路径错误

**错误信息：**
```
package go-one/internal/xxx is not in GOROOT
```

**解决方案：**
确保已经批量替换了所有导入路径：
```bash
# 检查是否还有 go-one 的导入
grep -r "go-one" --include="*.go" .

# 如果有，重新运行替换命令
find . -type f -name "*.go" -exec sed -i 's/go-one/my-awesome-project/g' {} +
```

### 问题5：JWT token 无效

**解决方案：**
1. 确认请求头格式正确：`Authorization: Bearer {token}`
2. 确认 token 未过期
3. 确认 `.env` 中的 `JWT_SECRET` 未更改

## 开发技巧

### 1. 使用热重载

安装 Air 进行热重载开发：

```bash
# 安装 Air
go install github.com/cosmtrek/air@latest

# 创建配置文件
cat > .air.toml << 'EOF'
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/server"
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "logs"]
EOF

# 启动热重载
air
```

### 2. 使用数据库迁移工具

```bash
# 安装 golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移文件
migrate create -ext sql -dir migrations -seq create_users_table

# 运行迁移
migrate -database "postgresql://user:pass@localhost:5432/db?sslmode=disable" -path migrations up
```

### 3. 使用 Make 命令

```bash
make help      # 查看所有可用命令
make deps      # 安装依赖
make build     # 编译项目
make run       # 运行编译后的二进制
make dev       # 开发模式运行
make test      # 运行测试
make clean     # 清理编译文件
```

## 推荐工具

- **Postman** - API测试
- **DBeaver** - 数据库管理
- **Redis Desktop Manager** - Redis管理
- **VS Code** + Go插件 - 代码编辑

## 获取帮助

如果遇到问题：
1. 查看日志文件：`logs/app.log`
2. 检查环境变量配置
3. 确认所有服务已启动
4. 查看完整文档

## 下一个里程碑

- [ ] 添加你的第一个业务模块
- [ ] 编写单元测试
- [ ] 配置 CI/CD
- [ ] 准备生产环境部署

**祝你开发愉快！** 🚀

