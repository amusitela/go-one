# Go-One 架构说明

本项目是基于 Gin + GORM + Redis 的分层 Web 服务模板，围绕“控制器（API）—服务（Service）—仓储（Repository）—模型（Model）”组织代码，并辅以中间件、序列化与配置初始化，提供用户注册/登录、JWT 鉴权、限流与基础安全头等能力。

## 目录结构

- `cmd/server/main.go`：应用入口，加载配置、初始化依赖、启动 HTTP 服务。
- `internal/server/router.go`：路由注册与中间件组合。
- `internal/api/*`：HTTP 控制器层，进行参数绑定、调用服务、返回统一响应。
- `internal/service/*`：业务逻辑层（DTO/VTO、错误模型、JWT 发放与校验、上下文封装）。
- `internal/repository/*`：数据访问层（对 GORM 的封装）。
- `internal/model/*`：领域模型与数据迁移（GORM 模型与 AutoMigrate）。
- `internal/middleware/*`：横切关注点（CORS、安全头、JWT 鉴权、基于 Redis 的令牌桶限流）。
- `internal/cache/*`：Redis 客户端初始化与封装。
- `internal/serializer/*`：HTTP 统一响应与视图对象（VTO）。
- `internal/conf/conf.go`：配置与依赖初始化（.env、日志、Redis、Postgres、JWT、Sentry）。
- `util/*`：日志、时间等通用工具。

## 请求生命周期（简述）

1. 进程启动：`cmd/server/main.go:18` 调用 `conf.Init()` 读取 `.env`，初始化 Logger、Redis、Postgres、JWT、Sentry，并构建路由后启动 `http.Server`。
2. 路由匹配：`internal/server/router.go:12` 创建 `gin.Engine`，挂载中间件与 `v1` 路由分组。
3. 中间件：
   - `CORS`（`internal/middleware/cors.go:8`）
   - 安全头（`internal/middleware/security.go:6`）
   - Sentry 捕获（`internal/server/router.go:16`）
   - 受保护路由下：`JWTMiddleware` 注入 `BusinessContext`（`internal/middleware/jwt.go:12`）与用户限流（`internal/middleware/ratelimit.go:16`）。
4. 控制器：`internal/api/user.go:19` 等绑定参数，构造 Service DTO，调用 `serviceManager` 实例化相应服务，接收结果并用 `serializer` 返回统一响应。
5. 服务层：业务校验、密码哈希、仓储读写、JWT 令牌对生成（`internal/service/user_service.go:38`、`internal/service/jwt.go:20`）。
6. 仓储层：使用 GORM 访问数据库（`internal/repository/user_repository.go:9`）。
7. 响应：统一的 `serializer.Response`（`internal/serializer/common.go:3`）。

## 路由与能力

- 公共路由（无鉴权）`/api/v1`：
  - `POST /auth/register` → 用户注册
  - `POST /auth/login` → 用户登录
  - `POST /auth/refresh` → 刷新令牌对（旋转 refresh token）
  - `POST /auth/logout` → 撤销 refresh token（登出）
  - `GET /ping` → 健康检查
- 受保护路由（JWT Bearer）：
  - `GET /user/profile` → 获取资料
  - `PUT /user/profile` → 更新资料
  - `POST /user/change-password` → 修改密码
  - `GET /user/list` → 用户列表（分页）

限流：
- 公共认证接口对单 IP 应用限流（`RateLimitMiddleware`）。
- 受保护接口对用户维度应用限流（从 `BusinessContext` 取 `UserUUID`）。

## 分层设计

- API（Controller）：
  - 位置：`internal/api/*.go`
  - 职责：请求参数绑定校验、调用 Service、错误转 HTTP（`context_helper.HandleServiceError`）。
- Service：
  - 位置：`internal/service/*.go`
  - 职责：业务规则、DTO/VTO 转换、调用仓储、发放/校验 JWT、构造业务错误（`ServiceError` 家族）。
  - 依赖注入：`ServiceManager` 统一创建服务（`internal/service/service_manager.go:7`）。
- Repository：
  - 位置：`internal/repository/*.go`
  - 职责：围绕模型的持久化操作（`UserRepository`）。
- Model：
  - 位置：`internal/model/*.go`
  - 职责：领域实体与迁移，`migration()` 在启动时执行 AutoMigrate。
- Middleware：
  - CORS、安全头、JWT、限流。
- Serializer：
  - 统一响应体与 VTO（`UserVTO`、`AuthTokenVTO`、`TokenPairVTO`）。

## 鉴权与会话

- JWT 配置（密钥与过期）由 `.env` 驱动（`internal/service/jwt.go`）。
- Access Token 最小负载：仅包含 `user_id` 与 `token_type=access`。
- Refresh Token：包含 `user_id`、`token_type=refresh` 与唯一 `jti`；所有 refresh token 落库持久化，支持撤销与旋转。
- 中间件从 `Authorization: Bearer <token>` 解析访问令牌，验证后注入 `BusinessContext`（`internal/middleware/jwt.go`）。
- 刷新流程：校验签名→查库校验 JTI→撤销旧 JTI→生成新 JTI 并落库→下发新 token 对（`internal/service/user_service.go`）。
- 登出：校验 refresh token 并将其 JTI 标记撤销（幂等）。

### cURL 示例

- 注册
  - `curl -X POST http://localhost:8080/api/v1/auth/register -H 'Content-Type: application/json' -d '{"username":"alice","password":"secret123"}'`
- 登录
  - `curl -X POST http://localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' -d '{"username":"alice","password":"secret123"}'`
- 刷新
  - `curl -X POST http://localhost:8080/api/v1/auth/refresh -H 'Content-Type: application/json' -d '{"refresh_token":"<refresh>"}'`
- 登出
  - `curl -X POST http://localhost:8080/api/v1/auth/logout -H 'Content-Type: application/json' -d '{"refresh_token":"<refresh>"}'`

## 数据与存储

- 数据库：PostgreSQL（GORM `postgres` driver），连接由 `.env` 拼接（`internal/conf/conf.go:29`），会话时区由 `DB_TIMEZONE` 设置（`internal/model/init.go:41`）。
- 迁移：启动时执行 `AutoMigrate(&User{})`（`internal/model/migration.go:5`）。
- Redis：连接与探活（`internal/cache/redis.go:10`），用于令牌桶限流（`internal/middleware/ratelimit.go:16`）。

## 错误与返回

- Service 层定义 `ValidationError`、`DatabaseError`、`AuthError`、`NotFoundError` 等，通过 `HandleServiceError` 映射为 HTTP 状态与统一响应（`internal/api/context_helper.go:21`）。
- 控制器直接返回 `serializer.Response`，包含 `code/msg/data/error`。

## 关键文件引用

- 入口：`cmd/server/main.go:18`, `cmd/server/main.go:36`
- 路由：`internal/server/router.go:12`
- 控制器聚合：`internal/api/handler.go:8`
- 用户控制器：`internal/api/user.go:19`
- JWT 中间件：`internal/middleware/jwt.go:12`
- 业务上下文：`internal/service/context.go:6`
- JWT 生成/校验：`internal/service/jwt.go:20`
- 用户服务：`internal/service/user_service.go:20`
- 用户仓储：`internal/repository/user_repository.go:9`
- 模型：`internal/model/user.go:6`

## 改进建议（可选）

- 刷新令牌持久化与撤销
  - 现状：刷新令牌仅为签名校验，未落库；无法主动撤销。
  - 建议：为 refresh token 引入 `jti`，落地到 Redis/DB，支持旋转与黑名单（登出、密码修改、风控）。
- JWT 负载最小化
  - 现状：access token 内含完整 `Account`，体积较大且易过期失真。
  - 建议：仅放必要 claim（`uid/role/scope`），业务查询按需读取。
- 限流策略与键设计
  - 现状：键为 `method+path` 维度，桶计算用 `period` 粗粒度换算。
  - 建议：加入突发（burst）与滑动窗口/漏桶实现，或使用 Redis Lua 保证原子性与更准的时间粒度。
- 配置安全
  - 将 `JWT_SECRET`、数据库口令迁移至安全配置（环境变量管理、密钥服务），避免默认值运行。
- 数据库与迁移
  - 为用户唯一键与常用查询添加索引（如 `username`/`email` 已有 unique，考虑联合索引视业务而定）。
  - 在迁移中引入版本化（例如 `golang-migrate`），避免生产环境直接 AutoMigrate。
- 日志与追踪
  - 统一 request-id/trace-id 注入日志上下文；为关键路径添加结构化字段（用户、路由、耗时）。
- API 一致性
  - 统一 `msg/code` 语义与错误码区间划分，沉淀错误码表与开发规范。
- 测试
  - 为 Service 与 Repository 层补充单元测试与集成测试（内存/容器化 DB、Redis）。

---

若你需要，我可以：
- 根据当前路由补充 `README` 使用说明与 cURL 示例。
- 实现 refresh token 持久化与旋转（含登出）。
- 加入更细的限流实现与可观测性（metrics/trace）。
