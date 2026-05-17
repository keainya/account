# Account 账户管理系统

基于 OAuth 2.0 的统一账户管理与应用元数据服务。

---

## 1. 系统概述

本系统提供统一的用户身份认证与授权服务，所有接入的应用共享同一套账户体系。用户只需一个账号，即可登录所有已注册的应用。同时，系统为每个应用提供独立的元数据存储空间，支持应用读写其授权用户的元数据。

**技术栈**

| 层级 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| 数据库 | SQLite (GORM) |
| 认证协议 | OAuth 2.0 (Authorization Code) |
| 前端 | 待定 |

---

## 2. 用户与角色

### 2.1 角色定义

系统包含两种角色：

| 角色 | 说明 |
|------|------|
| **管理员 (Admin)** | 拥有系统管理权限，可管理其他用户和应用 |
| **普通用户 (User)** | 可使用所有已注册应用的常规用户 |

### 2.2 首个管理员规则

- 系统中**第一个注册的用户自动成为管理员**。
- 后续注册的用户默认角色为**普通用户**。

### 2.3 角色提升

- 管理员可以将任意**普通用户提升为管理员**。
- 该操作不可撤销（除非由另一管理员手动降级）。
- 系统始终保证至少存在一名管理员。

### 2.4 用户属性

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 用户唯一标识 |
| `username` | string | 用户名，全局唯一 |
| `password_hash` | string | 密码哈希 |
| `email` | string | 邮箱（可选） |
| `role` | enum | `admin` / `user` |
| `created_at` | datetime | 注册时间 |
| `updated_at` | datetime | 更新时间 |

---

## 3. 应用管理

### 3.1 应用注册

管理员可在系统中注册应用，每个应用注册后获得：

| 属性 | 说明 |
|------|------|
| `client_id` | OAuth 2.0 客户端标识 |
| `client_secret` | OAuth 2.0 客户端密钥 |
| `redirect_uris` | 允许的回调地址列表 |
| `name` | 应用名称 |
| `description` | 应用描述 |

### 3.2 应用属性

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 应用唯一标识 |
| `name` | string | 应用名称 |
| `description` | string | 应用描述 |
| `client_id` | string | OAuth 客户端 ID |
| `client_secret` | string | OAuth 客户端密钥（加密存储） |
| `redirect_uris` | []string | 允许的重定向 URI 列表 |
| `created_at` | datetime | 创建时间 |

### 3.3 用户-应用关系

- **所有用户均可登录任意已注册的应用**，无需额外授权。
- 用户通过 OAuth 2.0 授权后，应用即获得该用户的身份凭证。

---

## 4. OAuth 2.0 认证服务

本系统作为 OAuth 2.0 **授权服务器**，为接入的应用提供标准化的登录服务。

### 4.1 授权模式

采用 **Authorization Code Grant（授权码模式）**，这是 OAuth 2.0 中最安全、推荐的方式。

### 4.2 认证流程

```
用户                 应用（客户端）              账户系统（本服务）
 |                       |                           |
 |  1. 点击登录           |                           |
 |---------------------->|                           |
 |                       |  2. 重定向到授权页面        |
 |                       |-------------------------->|
 |                       |                           |
 |  3. 展示登录页面        |                           |
 |<----------------------------------------------|   |
 |                       |                           |
 |  4. 输入凭据登录        |                           |
 |----------------------|-------------------------->|
 |                       |                           |
 |  5. 用户确认授权        |                           |
 |----------------------|-------------------------->|
 |                       |                           |
 |                       |  6. 返回授权码 (code)       |
 |                       |<--------------------------|
 |                       |                           |
 |                       |  7. 用 code 换取 token     |
 |                       |-------------------------->|
 |                       |                           |
 |                       |  8. 返回 access_token      |
 |                       |<--------------------------|
 |                       |                           |
 |                       |  9. 用 token 获取用户信息    |
 |                       |-------------------------->|
 |                       |                           |
 |                       | 10. 返回用户信息            |
 |                       |<--------------------------|
```

### 4.3 API 端点

#### 授权端点

```
GET /oauth/authorize
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `response_type` | 是 | 固定值 `code` |
| `client_id` | 是 | 应用注册时获得的 client_id |
| `redirect_uri` | 是 | 授权后的回调地址 |
| `scope` | 否 | 请求的权限范围 |
| `state` | 是 | 防 CSRF 的随机字符串 |

返回：302 重定向至 `{redirect_uri}?code={authorization_code}&state={state}`

#### 令牌端点

```
POST /oauth/token
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `grant_type` | 是 | 固定值 `authorization_code` |
| `code` | 是 | 上一步获得的授权码 |
| `redirect_uri` | 是 | 与授权请求相同的重定向 URI |
| `client_id` | 是 | 应用 client_id |
| `client_secret` | 是 | 应用 client_secret |

返回：

```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "rt_xxx..."
}
```

#### 用户信息端点

```
GET /oauth/userinfo
Authorization: Bearer {access_token}
```

返回：

```json
{
  "sub": "user-uuid",
  "username": "alice",
  "email": "alice@example.com",
  "role": "user"
}
```

### 4.4 Token 机制

| 类型 | 有效期 | 说明 |
|------|--------|------|
| Access Token | 1 小时 | 用于调用 API，格式为 JWT |
| Refresh Token | 30 天 | 用于刷新 Access Token |

---

## 5. 用户元数据存储

系统为**每个应用**提供独立的元数据存储空间，应用可以在其中读写其授权用户的元数据。

### 5.1 设计原则

- **按应用隔离**：每个应用的元数据空间相互独立。
- **应用内可读**：应用可以读取**任意授权用户**在该应用下的元数据。
- **用户粒度**：元数据以用户为单位存储，每个用户在每个应用下拥有一份独立的元数据记录。

### 5.2 数据结构

```
应用 A ──┬── 用户 Alice ── { "preferences": {...}, "progress": {...} }
         ├── 用户 Bob   ── { "preferences": {...}, "progress": {...} }
         └── 用户 Carol ── { "preferences": {...}, "custom_field": {...} }

应用 B ──┬── 用户 Alice ── { "theme": "dark", "language": "zh" }
         └── 用户 Bob   ── { "theme": "light" }
```

### 5.3 元数据 API

#### 读取指定用户的元数据

```
GET /api/apps/{client_id}/users/{user_id}/metadata
Authorization: Bearer {app_access_token}
```

返回：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "user_id": "user-uuid",
    "metadata": {
      "preferences": {
        "theme": "dark",
        "notifications": true
      }
    },
    "updated_at": "2026-05-17T12:00:00Z"
  }
}
```

#### 写入/更新指定用户的元数据

```
PUT /api/apps/{client_id}/users/{user_id}/metadata
Authorization: Bearer {app_access_token}
Content-Type: application/json

{
  "metadata": {
    "preferences": {
      "theme": "dark",
      "notifications": true
    }
  }
}
```

- 传入的 `metadata` 会与已有的进行**深度合并 (deep merge)**，而非完全覆盖。
- 若该用户尚无元数据记录，则创建新记录。

#### 批量读取用户元数据

```
GET /api/apps/{client_id}/metadata?user_ids=uuid1,uuid2,uuid3
Authorization: Bearer {app_access_token}
```

返回：

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "user_id": "uuid1",
      "metadata": { "...": "..." },
      "updated_at": "..."
    },
    {
      "user_id": "uuid2",
      "metadata": { "...": "..." },
      "updated_at": "..."
    }
  ]
}
```

### 5.4 权限说明

| 操作 | 管理员 | 普通用户 | 应用 Token |
|------|--------|---------|------------|
| 读自己元数据 | ✅ | ✅ | ✅ |
| 写自己元数据 | ✅ | ✅ | ✅ |
| 读任意用户元数据 | ✅ | ❌ | ✅ (仅限本应用) |
| 写任意用户元数据 | ✅ | ❌ | ✅ (仅限本应用) |

---

## 6. 用户管理 API

### 6.1 注册

```
POST /api/auth/register
Content-Type: application/json

{
  "username": "alice",
  "password": "secure_password",
  "email": "alice@example.com"
}
```

- 第一个注册的用户自动获得 `admin` 角色。
- 后续注册的用户默认为 `user` 角色。

### 6.2 登录

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "alice",
  "password": "secure_password"
}
```

返回 session cookie。

### 6.3 登出

```
POST /api/auth/logout
```

### 6.4 当前用户信息

```
GET /api/auth/me
```

### 6.5 管理员接口

#### 提升用户为管理员

```
PUT /api/admin/users/{user_id}/promote
```

仅管理员可调用。

#### 获取用户列表

```
GET /api/admin/users?page=1&page_size=20
```

仅管理员可调用。

---

## 7. 系统架构

```
┌──────────────────────────────────────────────────────┐
│                    Account System                     │
│                                                       │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  │
│  │  Auth 模块   │  │  OAuth 模块   │  │ Metadata 模块│  │
│  │             │  │              │  │             │  │
│  │ · 注册/登录  │  │ · 授权码流程  │  │ · CRUD 操作  │  │
│  │ · 会话管理   │  │ · Token 签发  │  │ · 应用隔离   │  │
│  │ · 角色管理   │  │ · Token 验证  │  │ · 批量查询   │  │
│  └──────┬──────┘  └──────┬───────┘  └──────┬──────┘  │
│         │                │                  │         │
│         └────────────────┼──────────────────┘         │
│                          │                            │
│                   ┌──────┴──────┐                     │
│                   │   SQLite    │                     │
│                   │  (GORM)     │                     │
│                   └─────────────┘                     │
└──────────────────────────────────────────────────────┘
```

**目录结构**

```
account/
├── server/
│   ├── main.go              # 入口
│   ├── go.mod
│   ├── object/              # 数据模型 & DB 初始化
│   │   └── database.go
│   ├── router/              # 路由定义
│   │   └── default.go
│   ├── service/             # 业务逻辑 & API Handler
│   │   ├── auth.go          # 认证相关
│   │   ├── oauth.go         # OAuth 2.0 相关
│   │   ├── metadata.go      # 元数据存储
│   │   ├── admin.go         # 管理员接口
│   │   └── response.go      # 统一响应格式
│   └── utils/               # 工具函数
│       └── fs.go
└── web/                     # 前端（待开发）
```

---

## 8. 数据库表设计

### users — 用户表

| 列 | 类型 | 说明 |
|----|------|------|
| id | TEXT (UUID) | 主键 |
| username | TEXT UNIQUE | 用户名 |
| password_hash | TEXT | 密码哈希 |
| email | TEXT | 邮箱 |
| role | TEXT | admin / user |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### apps — 应用表

| 列 | 类型 | 说明 |
|----|------|------|
| id | TEXT (UUID) | 主键 |
| name | TEXT | 应用名称 |
| description | TEXT | 描述 |
| client_id | TEXT UNIQUE | OAuth Client ID |
| client_secret | TEXT | OAuth Client Secret |
| redirect_uris | TEXT (JSON) | 允许的重定向 URI 列表 |
| created_at | DATETIME | 创建时间 |

### user_metadata — 用户元数据表

| 列 | 类型 | 说明 |
|----|------|------|
| id | TEXT (UUID) | 主键 |
| user_id | TEXT | 关联 user.id |
| app_id | TEXT | 关联 app.id |
| metadata | TEXT (JSON) | 元数据内容 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

- 联合唯一索引：(user_id, app_id)

---

## 9. 快速开始

```bash
# 进入服务端目录
cd server

# 安装依赖
go mod tidy

# 运行服务（默认监听 :8080）
go run main.go
```

服务启动后：
- 访问 `http://localhost:8080/status` 检查服务状态。
- 首个注册用户将自动成为管理员。
