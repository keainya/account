# Account API 文档

> Base URL: `http://localhost:8080`

---

## 目录

- [1. 通用约定](#1-通用约定)
- [2. 认证 API](#2-认证-api)
- [3. OAuth 2.0 API](#3-oauth-20-api)
- [4. 应用管理 API](#4-应用管理-api)
- [5. 用户元数据 API](#5-用户元数据-api)
- [6. 管理员 API](#6-管理员-api)
- [7. 错误码](#7-错误码)

---

## 1. 通用约定

### 1.1 认证方式

本系统支持两种认证方式：

| 方式 | 适用场景 | 携带方式 |
|------|---------|---------|
| **Session Cookie** | 前端页面、管理后台 | 登录后自动携带 |
| **Bearer Token** | 第三方应用 API 调用 | `Authorization: Bearer {token}` |

### 1.2 统一响应格式

所有 API 返回统一的 JSON 结构：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 状态码，`0` 表示成功，非 `0` 表示错误 |
| `msg` | string | 提示信息 |
| `data` | any | 响应数据，错误时为 `null` |

### 1.3 分页格式

分页接口使用统一的查询参数和响应格式：

**请求参数**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码，从 1 开始 |
| `page_size` | int | 20 | 每页条数，最大 100 |

**响应格式**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [],
    "total": 42,
    "page": 1,
    "page_size": 20
  }
}
```

### 1.4 时间格式

所有时间字段使用 ISO 8601 格式，UTC 时区：

```
2026-05-17T12:00:00Z
```

---

## 2. 认证 API

### 2.1 注册

首个注册用户自动获得 `admin` 角色。后续用户默认为 `user`。

```
POST /api/auth/register
Content-Type: application/json
```

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 用户名，3-32 字符，字母开头，仅允许字母数字下划线 |
| `password` | string | 是 | 密码，6-128 字符 |
| `email` | string | 否 | 邮箱 |

**请求示例**

```json
{
  "username": "alice",
  "password": "secure_password",
  "email": "alice@example.com"
}
```

**成功响应** `201`

```json
{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "role": "user",
    "created_at": "2026-05-17T12:00:00Z"
  }
}
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 1001 | 用户名已存在 | username 重复 |
| 1002 | 参数校验失败 | 字段格式不符合要求 |

---

### 2.2 登录

```
POST /api/auth/login
Content-Type: application/json
```

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 用户名 |
| `password` | string | 是 | 密码 |

**请求示例**

```json
{
  "username": "alice",
  "password": "secure_password"
}
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "登录成功",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "role": "admin"
  }
}
```

> 同时设置 session cookie，后续请求自动携带。

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 1003 | 用户名或密码错误 | 凭据不匹配 |
| 1004 | 用户不存在 | 用户名未注册 |

---

### 2.3 登出

```
POST /api/auth/logout
Authentication: Session
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "已登出"
}
```

> 清除当前 session。

---

### 2.4 获取当前用户信息

```
GET /api/auth/me
Authentication: Session
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "role": "admin",
    "created_at": "2026-05-17T12:00:00Z",
    "updated_at": "2026-05-17T12:00:00Z"
  }
}
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 2001 | 未登录 | 没有有效 session |

---

## 3. OAuth 2.0 API

### 3.1 认证流程概述

```
应用客户端                               Account 服务
    |                                        |
    |-- GET /oauth/authorize --------------->|  用户登录并授权
    |<-- 302 ?code={code}&state={state} ----|
    |                                        |
    |-- POST /oauth/token ------------------>|  用 code 换取 token
    |<-- 200 { access_token, refresh_token } |
    |                                        |
    |-- GET /oauth/userinfo ---------------->|  获取用户信息
    |   Authorization: Bearer {access_token} |
    |<-- 200 { sub, username, ... } --------|
```

### 3.2 授权端点

用户在此端点上登录并授权第三方应用。

```
GET /oauth/authorize
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `response_type` | string | 是 | 固定值 `code` |
| `client_id` | string | 是 | 应用注册时获得的 client_id |
| `redirect_uri` | string | 是 | 授权后的回调地址，必须在应用注册的 redirect_uris 列表中 |
| `scope` | string | 否 | 请求的权限范围，空格分隔 |
| `state` | string | 是 | 防 CSRF 攻击的随机字符串 |

**请求示例**

```
GET /oauth/authorize?response_type=code&client_id=app_3f7a2b&redirect_uri=https://example.com/callback&state=x8k2m9n1
```

**用户交互**

若用户未登录，服务端返回登录页面；登录后展示授权确认页面。

**成功响应**

用户确认授权后，302 重定向至：

```
{redirect_uri}?code=abc123def456&state=x8k2m9n1
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 3001 | 无效的 client_id | 应用未注册 |
| 3002 | redirect_uri 不匹配 | 不在允许列表中 |
| 3003 | 用户拒绝授权 | 用户点击了拒绝 |

---

### 3.3 令牌端点

用授权码换取 access_token。

```
POST /oauth/token
Content-Type: application/x-www-form-urlencoded
```

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `grant_type` | string | 是 | 固定值 `authorization_code` |
| `code` | string | 是 | 授权端点返回的授权码 |
| `redirect_uri` | string | 是 | 与授权请求相同的 redirect_uri |
| `client_id` | string | 是 | 应用 client_id |
| `client_secret` | string | 是 | 应用 client_secret |

**请求示例**

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=abc123def456" \
  -d "redirect_uri=https://example.com/callback" \
  -d "client_id=app_3f7a2b" \
  -d "client_secret=sec_xyz789"
```

**成功响应** `200`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "rt_d8f3a1c9b2e45678"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `access_token` | string | JWT 格式的访问令牌 |
| `token_type` | string | 固定值 `Bearer` |
| `expires_in` | int | 有效期，单位秒（3600 = 1 小时） |
| `refresh_token` | string | 用于刷新 access_token 的令牌 |

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 3004 | 授权码无效 | code 已过期或已使用 |
| 3005 | client_secret 错误 | 密钥不匹配 |
| 3006 | grant_type 不支持 | 非 `authorization_code` |

---

### 3.4 刷新令牌

```
POST /oauth/token
Content-Type: application/x-www-form-urlencoded
```

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `grant_type` | string | 是 | 固定值 `refresh_token` |
| `refresh_token` | string | 是 | 令牌端点返回的 refresh_token |
| `client_id` | string | 是 | 应用 client_id |
| `client_secret` | string | 是 | 应用 client_secret |

**请求示例**

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=refresh_token" \
  -d "refresh_token=rt_d8f3a1c9b2e45678" \
  -d "client_id=app_3f7a2b" \
  -d "client_secret=sec_xyz789"
```

**成功响应** `200`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "rt_f9a2b1c3d4e56789"
}
```

> 每次刷新会生成新的 refresh_token，旧的立即失效。

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 3007 | refresh_token 无效或已过期 | token 已失效 |

---

### 3.5 获取用户信息

使用 access_token 获取当前授权用户的身份信息。

```
GET /oauth/userinfo
Authorization: Bearer {access_token}
```

**成功响应** `200`

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "username": "alice",
  "email": "alice@example.com",
  "role": "user"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `sub` | string | 用户唯一标识 (subject) |
| `username` | string | 用户名 |
| `email` | string | 邮箱（可能为空） |
| `role` | string | 用户角色 |

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 3008 | access_token 无效或已过期 | token 校验失败 |

---

### 3.6 Token 说明

| 属性 | Access Token | Refresh Token |
|------|-------------|---------------|
| 格式 | JWT | 随机字符串 |
| 有效期 | 1 小时 | 30 天 |
| 用途 | 调用 API | 刷新 Access Token |
| 泄露风险 | 低（短期有效） | 高（需安全存储） |

**JWT Payload 结构**

```json
{
  "sub": "user-uuid",
  "username": "alice",
  "role": "user",
  "client_id": "app_3f7a2b",
  "iat": 1747500000,
  "exp": 1747503600
}
```

| 字段 | 说明 |
|------|------|
| `sub` | 用户 ID |
| `username` | 用户名 |
| `role` | 用户角色 |
| `client_id` | 授权应用 ID |
| `iat` | 签发时间 (Unix timestamp) |
| `exp` | 过期时间 (Unix timestamp) |

---

## 4. 应用管理 API

> 需要 Session 认证，操作需要管理员权限。

### 4.1 创建应用

```
POST /api/admin/apps
Authentication: Session (admin)
Content-Type: application/json
```

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 应用名称，1-64 字符 |
| `description` | string | 否 | 应用描述 |
| `redirect_uris` | []string | 是 | 允许的 OAuth 回调地址列表 |

**请求示例**

```json
{
  "name": "我的应用",
  "description": "一个示例应用",
  "redirect_uris": [
    "https://example.com/callback",
    "https://example.com/oauth/callback"
  ]
}
```

**成功响应** `201`

```json
{
  "code": 0,
  "msg": "应用创建成功",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "我的应用",
    "description": "一个示例应用",
    "client_id": "app_3f7a2b1c",
    "client_secret": "sec_d4e5f6a7b8c9",
    "redirect_uris": [
      "https://example.com/callback",
      "https://example.com/oauth/callback"
    ],
    "created_at": "2026-05-17T12:00:00Z"
  }
}
```

> `client_secret` 仅在创建时返回一次，请妥善保存。

---

### 4.2 获取应用列表

```
GET /api/admin/apps?page=1&page_size=20
Authentication: Session (admin)
```

**请求示例**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "name": "我的应用",
        "description": "一个示例应用",
        "client_id": "app_3f7a2b1c",
        "redirect_uris": [
          "https://example.com/callback"
        ],
        "created_at": "2026-05-17T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

> 列表中**不返回 `client_secret`**。

---

### 4.3 获取单个应用

```
GET /api/admin/apps/{app_id}
Authentication: Session (admin)
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "我的应用",
    "description": "一个示例应用",
    "client_id": "app_3f7a2b1c",
    "client_secret": "sec_d4e5f6a7b8c9",
    "redirect_uris": [
      "https://example.com/callback"
    ],
    "created_at": "2026-05-17T12:00:00Z"
  }
}
```

---

### 4.4 更新应用

```
PUT /api/admin/apps/{app_id}
Authentication: Session (admin)
Content-Type: application/json
```

**请求体**（所有字段可选）

```json
{
  "name": "新名称",
  "description": "新描述",
  "redirect_uris": [
    "https://new.example.com/callback"
  ]
}
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "应用已更新",
  "data": { "...": "..." }
}
```

---

### 4.5 重置密钥

```
POST /api/admin/apps/{app_id}/reset-secret
Authentication: Session (admin)
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "密钥已重置",
  "data": {
    "client_secret": "sec_new_a1b2c3d4e5"
  }
}
```

> 重置后旧密钥立即失效。新密钥仅在此次响应中返回。

---

### 4.6 删除应用

```
DELETE /api/admin/apps/{app_id}
Authentication: Session (admin)
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "应用已删除"
}
```

> 删除应用会同时清理该应用下的所有元数据。

---

## 5. 用户元数据 API

> 应用通过 access_token 调用。应用 token 可读写本应用下任意授权用户的元数据。

### 5.1 读取用户元数据

```
GET /api/apps/{client_id}/users/{user_id}/metadata
Authorization: Bearer {access_token}
```

**路径参数**

| 参数 | 说明 |
|------|------|
| `client_id` | 应用 client_id |
| `user_id` | 目标用户 ID |

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": {
      "preferences": {
        "theme": "dark",
        "language": "zh",
        "notifications": true
      },
      "progress": {
        "level": 12,
        "score": 9500
      }
    },
    "created_at": "2026-05-01T08:00:00Z",
    "updated_at": "2026-05-17T12:00:00Z"
  }
}
```

**元数据不存在时**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": {},
    "created_at": null,
    "updated_at": null
  }
}
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 4001 | 无权访问该应用 | token 的 client_id 与请求的不匹配 |
| 4002 | 用户不存在 | user_id 无效 |

---

### 5.2 写入/更新用户元数据

传入的元数据与已有数据进行**深度合并**。若无已有数据则创建。

```
PUT /api/apps/{client_id}/users/{user_id}/metadata
Authorization: Bearer {access_token}
Content-Type: application/json
```

**请求体**

| 字段 | 类型 | 说明 |
|------|------|------|
| `metadata` | object | 要写入的元数据，任意 JSON 对象 |

**请求示例**

```json
{
  "metadata": {
    "preferences": {
      "theme": "dark",
      "notifications": true
    }
  }
}
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "元数据已更新",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": {
      "preferences": {
        "theme": "dark",
        "language": "zh",
        "notifications": true
      },
      "progress": {
        "level": 12,
        "score": 9500
      }
    },
    "updated_at": "2026-05-17T12:30:00Z"
  }
}
```

> 返回的是合并后的完整元数据。

**深度合并示例**

```
已有数据: { "a": { "x": 1, "y": 2 }, "b": 3 }
传入数据: { "a": { "y": 99, "z": 5 } }
合并结果: { "a": { "x": 1, "y": 99, "z": 5 }, "b": 3 }
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 4001 | 无权访问该应用 | token 的 client_id 与请求的不匹配 |
| 4003 | metadata 格式无效 | 不是合法的 JSON 对象 |

---

### 5.3 批量读取用户元数据

```
GET /api/apps/{client_id}/metadata?user_ids=uuid1,uuid2,uuid3
Authorization: Bearer {access_token}
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_ids` | string | 是 | 逗号分隔的用户 ID 列表，最多 100 个 |

**请求示例**

```
GET /api/apps/app_3f7a2b1c/metadata?user_ids=uuid-1,uuid-2,uuid-3
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "user_id": "uuid-1",
      "metadata": {
        "theme": "dark"
      },
      "updated_at": "2026-05-17T12:00:00Z"
    },
    {
      "user_id": "uuid-2",
      "metadata": {
        "high_score": 9500
      },
      "updated_at": "2026-05-17T11:00:00Z"
    },
    {
      "user_id": "uuid-3",
      "metadata": {},
      "updated_at": null
    }
  ]
}
```

> 不存在的用户或尚无元数据的用户返回空 `metadata`。

---

### 5.4 读取当前用户在指定应用下的元数据

用于前端场景，用户通过 session 读取自己在某应用下的数据。

```
GET /api/apps/{client_id}/my-metadata
Authentication: Session
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "metadata": {
      "preferences": {
        "theme": "dark"
      }
    },
    "updated_at": "2026-05-17T12:00:00Z"
  }
}
```

---

### 5.5 写入当前用户在指定应用下的元数据

```
PUT /api/apps/{client_id}/my-metadata
Authentication: Session
Content-Type: application/json
```

**请求体**

```json
{
  "metadata": {
    "preferences": {
      "theme": "light"
    }
  }
}
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "元数据已更新",
  "data": {
    "metadata": {
      "preferences": {
        "theme": "light"
      }
    },
    "updated_at": "2026-05-17T12:30:00Z"
  }
}
```

---

### 5.6 权限矩阵

| 端点 | 管理员 (Session) | 普通用户 (Session) | 应用 Token (Bearer) |
|------|:---:|:---:|:---:|
| GET /api/apps/{cid}/users/{uid}/metadata | ✅ 任意用户 | ✅ 仅自己 | ✅ 本应用任意用户 |
| PUT /api/apps/{cid}/users/{uid}/metadata | ✅ 任意用户 | ✅ 仅自己 | ✅ 本应用任意用户 |
| GET /api/apps/{cid}/metadata | ✅ | ❌ | ✅ 本应用 |
| GET /api/apps/{cid}/my-metadata | ✅ | ✅ | N/A |
| PUT /api/apps/{cid}/my-metadata | ✅ | ✅ | N/A |

---

## 6. 管理员 API

> 需要 Session 认证 + `admin` 角色。

### 6.1 用户列表

```
GET /api/admin/users?page=1&page_size=20
Authentication: Session (admin)
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "alice",
        "email": "alice@example.com",
        "role": "admin",
        "created_at": "2026-05-17T12:00:00Z",
        "updated_at": "2026-05-17T12:00:00Z"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "username": "bob",
        "email": "bob@example.com",
        "role": "user",
        "created_at": "2026-05-17T12:05:00Z",
        "updated_at": "2026-05-17T12:05:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 6.2 提升用户为管理员

```
PUT /api/admin/users/{user_id}/promote
Authentication: Session (admin)
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "已提升为管理员",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "username": "bob",
    "role": "admin",
    "updated_at": "2026-05-17T12:40:00Z"
  }
}
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 5001 | 用户不存在 | user_id 无效 |
| 5002 | 用户已是管理员 | 目标用户角色已为 admin |

---

### 6.3 降级管理员为普通用户

```
PUT /api/admin/users/{user_id}/demote
Authentication: Session (admin)
```

**成功响应** `200`

```json
{
  "code": 0,
  "msg": "已降级为普通用户",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "username": "bob",
    "role": "user",
    "updated_at": "2026-05-17T12:45:00Z"
  }
}
```

**错误响应**

| code | msg | 说明 |
|------|-----|------|
| 5001 | 用户不存在 | user_id 无效 |
| 5003 | 不能降级自己 | 管理员不能降级自身 |
| 5004 | 系统最后一个管理员不可降级 | 必须至少保留一个管理员 |

---

## 7. 错误码

### 认证相关 (1xxx)

| code | 说明 |
|------|------|
| 1001 | 用户名已存在 |
| 1002 | 参数校验失败 |
| 1003 | 用户名或密码错误 |
| 1004 | 用户不存在 |

### 会话相关 (2xxx)

| code | 说明 |
|------|------|
| 2001 | 未登录 |
| 2002 | 无权限（非管理员） |

### OAuth 相关 (3xxx)

| code | 说明 |
|------|------|
| 3001 | 无效的 client_id |
| 3002 | redirect_uri 不匹配 |
| 3003 | 用户拒绝授权 |
| 3004 | 授权码无效或已使用 |
| 3005 | client_secret 错误 |
| 3006 | grant_type 不支持 |
| 3007 | refresh_token 无效或已过期 |
| 3008 | access_token 无效或已过期 |

### 元数据相关 (4xxx)

| code | 说明 |
|------|------|
| 4001 | 无权访问该应用的数据 |
| 4002 | 目标用户不存在 |
| 4003 | metadata 格式无效 |

### 管理员相关 (5xxx)

| code | 说明 |
|------|------|
| 5001 | 用户不存在 |
| 5002 | 用户已是管理员 |
| 5003 | 不能操作自己 |
| 5004 | 不可移除最后一个管理员 |

### 系统相关 (9xxx)

| code | 说明 |
|------|------|
| 9000 | 服务器内部错误 |
| 9001 | 数据库错误 |

---

## 附录

### A. curl 快速测试

```bash
# 1. 注册（首个用户自动为管理员）
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -c cookies.txt

# 2. 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -c cookies.txt

# 3. 查看当前用户
curl http://localhost:8080/api/auth/me -b cookies.txt

# 4. 创建应用（管理员）
curl -X POST http://localhost:8080/api/admin/apps \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"测试应用","redirect_uris":["http://localhost:3000/callback"]}'

# 5. OAuth 授权（浏览器打开）
# http://localhost:8080/oauth/authorize?response_type=code&client_id=app_xxx&redirect_uri=http://localhost:3000/callback&state=random123

# 6. 换取 token（用上一步的 code）
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code={authorization_code}" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "client_id=app_xxx" \
  -d "client_secret=sec_xxx"

# 7. 获取用户信息
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer {access_token}"

# 8. 写入元数据
curl -X PUT http://localhost:8080/api/apps/app_xxx/users/{user_id}/metadata \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{"metadata":{"key":"value"}}'

# 9. 读取元数据
curl http://localhost:8080/api/apps/app_xxx/users/{user_id}/metadata \
  -H "Authorization: Bearer {access_token}"
```
