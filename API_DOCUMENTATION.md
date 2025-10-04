# API 文档

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: Bearer Token (JWT)

## 响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "操作成功",
  "data": { ... }
}
```

### 错误响应
```json
{
  "code": 400,
  "msg": "错误描述",
  "error": "详细错误信息"
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/认证失败 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 422 | 验证错误 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

---

## 认证相关

### 用户注册

**POST** `/auth/register`

#### 请求参数

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "123456"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-50字符 |
| email | string | 否 | 邮箱地址 |
| password | string | 是 | 密码，至少6字符 |

#### 响应示例

```json
{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "testuser",
      "avatar": "",
      "status": 1,
      "created_at": "2025-10-04T12:00:00Z",
      "updated_at": "2025-10-04T12:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 限流规则
- IP限流：10秒内最多6次请求

---

### 用户登录

**POST** `/auth/login`

#### 请求参数

```json
{
  "username": "testuser",
  "password": "123456"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

#### 响应示例

```json
{
  "code": 0,
  "msg": "登录成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "testuser",
      "avatar": "",
      "status": 1,
      "created_at": "2025-10-04T12:00:00Z",
      "updated_at": "2025-10-04T12:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 限流规则
- IP限流：10秒内最多6次请求

---

## 用户相关

### 获取用户资料

**GET** `/user/profile`

🔐 **需要认证**

#### 请求头

```
Authorization: Bearer {token}
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "获取成功",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "测试用户",
    "avatar": "https://example.com/avatar.jpg",
    "status": 1,
    "created_at": "2025-10-04T12:00:00Z",
    "updated_at": "2025-10-04T12:00:00Z"
  }
}
```

#### 限流规则
- 用户限流：1分钟内最多60次请求

---

### 更新用户资料

**PUT** `/user/profile`

🔐 **需要认证**

#### 请求头

```
Authorization: Bearer {token}
```

#### 请求参数

```json
{
  "nickname": "新昵称",
  "avatar": "https://example.com/new-avatar.jpg"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称 |
| avatar | string | 否 | 头像URL |

#### 响应示例

```json
{
  "code": 0,
  "msg": "更新成功",
  "data": null
}
```

---

### 修改密码

**POST** `/user/change-password`

🔐 **需要认证**

#### 请求头

```
Authorization: Bearer {token}
```

#### 请求参数

```json
{
  "old_password": "oldpass123",
  "new_password": "newpass456"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 原密码 |
| new_password | string | 是 | 新密码，至少6字符 |

#### 响应示例

```json
{
  "code": 0,
  "msg": "密码修改成功",
  "data": null
}
```

---

### 获取用户列表

**GET** `/user/list`

🔐 **需要认证**

#### 请求头

```
Authorization: Bearer {token}
```

#### 查询参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量，最大100 |

#### 请求示例

```
GET /user/list?page=1&page_size=20
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "获取成功",
  "data": {
    "list": [
      {
        "id": 1,
        "username": "user1",
        "email": "user1@example.com",
        "nickname": "用户1",
        "avatar": "",
        "status": 1,
        "created_at": "2025-10-04T12:00:00Z",
        "updated_at": "2025-10-04T12:00:00Z"
      },
      {
        "id": 2,
        "username": "user2",
        "email": "user2@example.com",
        "nickname": "用户2",
        "avatar": "",
        "status": 1,
        "created_at": "2025-10-04T12:01:00Z",
        "updated_at": "2025-10-04T12:01:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 系统相关

### 健康检查

**GET** `/ping`

#### 响应示例

```json
{
  "code": 0,
  "msg": "pong",
  "data": null
}
```

---

## 限流说明

框架使用令牌桶算法进行限流，当请求被限流时：

**响应状态码**: `429 Too Many Requests`

**响应头**:
- `X-RateLimit-Limit`: 桶容量（最大令牌数）
- `X-RateLimit-Remaining`: 剩余令牌数
- `X-RateLimit-Retry-After`: 建议重试时间（秒）

**响应示例**:
```json
{
  "code": 429,
  "msg": "请求过于频繁，请在 30 秒后重试",
  "error": null
}
```

---

## cURL 示例

### 注册
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "123456"
  }'
```

### 登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

### 获取用户资料
```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 更新用户资料
```bash
curl -X PUT http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "新昵称",
    "avatar": "https://example.com/avatar.jpg"
  }'
```

### 修改密码
```bash
curl -X POST http://localhost:8080/api/v1/user/change-password \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "123456",
    "new_password": "newpass123"
  }'
```

### 获取用户列表
```bash
curl -X GET "http://localhost:8080/api/v1/user/list?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Postman 集合

你可以导入以下JSON到Postman进行测试：

```json
{
  "info": {
    "name": "Go-One API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080/api/v1"
    },
    {
      "key": "token",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "header": [],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"username\": \"testuser\",\n  \"email\": \"test@example.com\",\n  \"password\": \"123456\"\n}",
              "options": {
                "raw": {
                  "language": "json"
                }
              }
            },
            "url": "{{baseUrl}}/auth/register"
          }
        },
        {
          "name": "Login",
          "request": {
            "method": "POST",
            "header": [],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"username\": \"testuser\",\n  \"password\": \"123456\"\n}",
              "options": {
                "raw": {
                  "language": "json"
                }
              }
            },
            "url": "{{baseUrl}}/auth/login"
          }
        }
      ]
    }
  ]
}
```

---

## 开发建议

1. **使用环境变量** - 不同环境使用不同的 baseUrl
2. **保存token** - 登录后保存token用于后续请求
3. **处理错误** - 根据错误码进行相应处理
4. **遵守限流** - 避免频繁请求导致被限流
5. **使用HTTPS** - 生产环境必须使用HTTPS

---

## 常见问题

### Q: 如何刷新token？
A: 当前版本未实现刷新token功能，可以在 `util/jwt.go` 中添加刷新token逻辑。

### Q: token过期时间是多久？
A: 默认1小时，可在 `.env` 文件中通过 `JWT_ACCESS_TOKEN_EXPIRE` 配置。

### Q: 如何退出登录？
A: 客户端删除token即可。如需服务端退出，可将token加入黑名单（需要在Redis中实现）。

### Q: 支持第三方登录吗？
A: 当前版本不支持，可以参考OAuth2.0协议自行实现。

