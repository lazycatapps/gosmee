# API 设计

本文档描述 Gosmee Web UI 的 REST API 接口规范。所有 API 端点均位于 `/api/v1` 路径下。

## 认证机制

- 当启用 OIDC 认证时,除公共端点外的所有 API 都需要有效的 session cookie
- 未认证的 API 请求返回 401 Unauthorized
- 未认证的浏览器请求自动重定向到登录页面

### 公共端点 (无需认证)

- `/api/v1/health` - 健康检查
- `/api/v1/auth/*` - 所有认证相关端点

---

## Client 实例管理

### POST /api/v1/clients

创建 gosmee client 实例

**请求参数:**

```json
{
  "name": "Agola Webhook",
  "description": "Webhook forwarder for Agola CI",
  "smeeUrl": "https://hook.pipelinesascode.com/GTzCkZZwEGTv",
  "targetUrl": "https://agola.liu.heiyu.space/webhooks?agolaid=agola&projectid=xxx",
  "targetTimeout": 60,
  "httpie": false,
  "ignoreEvents": ["push", "pull_request"],
  "noReplay": false,
  "sseBufferSize": 1048576
}
```

**字段说明:**

- `name` (必填): 实例名称,1-50 字符
- `description` (可选): 实例描述,最多 200 字符
- `smeeUrl` (必填): Gosmee server 的事件源地址 (HTTPS URL)
- `targetUrl` (必填): 目标 Webhook 接收地址 (HTTP/HTTPS URL)
- `targetTimeout` (可选): 目标连接超时时间(秒),默认 60
- `httpie` (可选): 是否生成 HTTPie 格式脚本,默认 false (使用 cURL)
- `ignoreEvents` (可选): 需要过滤的事件类型数组
- `noReplay` (可选): 仅保存事件不转发,默认 false
- `sseBufferSize` (可选): SSE 缓冲区大小(字节),默认 1048576

**成功响应 (201):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "user-123",
  "name": "Agola Webhook",
  "description": "Webhook forwarder for Agola CI",
  "status": "stopped",
  "smeeUrl": "https://hook.pipelinesascode.com/GTzCkZZwEGTv",
  "targetUrl": "https://agola.liu.heiyu.space/webhooks?agolaid=agola",
  "targetTimeout": 60,
  "httpie": false,
  "ignoreEvents": ["push", "pull_request"],
  "noReplay": false,
  "sseBufferSize": 1048576,
  "restartCount": 0,
  "todayEvents": 0,
  "totalEvents": 0,
  "createdAt": "2025-10-01T10:30:00Z",
  "updatedAt": "2025-10-01T10:30:00Z"
}
```

**错误响应:**

- **400 Bad Request** - 请求参数错误
  ```json
  {
    "error": "name is required"
  }
  ```
- **500 Internal Server Error** - 服务器内部错误

---

### GET /api/v1/clients

查询当前用户的 client 实例列表

**查询参数:**

- `page` (可选): 页码,从 1 开始,默认 1
- `pageSize` (可选): 每页数量,默认 20,最大 100
- `status` (可选): 过滤状态,可选值: `running`, `stopped`, `error`
- `search` (可选): 按名称搜索
- `sortBy` (可选): 排序字段,默认 `createdAt`
- `sortOrder` (可选): 排序方向,可选值: `asc`, `desc`,默认 `desc`

**成功响应 (200):**

```json
{
  "total": 15,
  "page": 1,
  "pageSize": 20,
  "clients": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Agola Webhook",
      "status": "running",
      "smeeUrl": "https://hook.pipelinesascode.com/GTzCkZZwEGTv",
      "targetUrl": "https://agola.liu.heiyu.space/webhooks?agolaid=agola",
      "todayEvents": 15,
      "totalEvents": 342,
      "lastActivity": "2025-10-01T14:23:15Z"
    }
  ]
}
```

**错误响应:**

- **500 Internal Server Error** - 服务器内部错误

---

### GET /api/v1/clients/:id

获取单个 client 实例详情

**路径参数:**

- `id`: Client ID (UUID 格式)

**成功响应 (200):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "user-123",
  "name": "Agola Webhook",
  "description": "Webhook forwarder for Agola CI",
  "status": "running",
  "smeeUrl": "https://hook.pipelinesascode.com/GTzCkZZwEGTv",
  "targetUrl": "https://agola.liu.heiyu.space/webhooks?agolaid=agola",
  "targetTimeout": 60,
  "httpie": false,
  "ignoreEvents": ["push"],
  "noReplay": false,
  "sseBufferSize": 1048576,
  "pid": 12345,
  "startedAt": "2025-10-01T10:30:00Z",
  "restartCount": 2,
  "todayEvents": 15,
  "totalEvents": 342,
  "lastActivity": "2025-10-01T14:23:15Z",
  "createdAt": "2025-09-15T08:00:00Z",
  "updatedAt": "2025-10-01T10:30:00Z"
}
```

**错误响应:**

- **404 Not Found** - Client 不存在
  ```json
  {
    "error": "Client not found"
  }
  ```

---

### PUT /api/v1/clients/:id

更新 client 实例配置

**路径参数:**

- `id`: Client ID (UUID 格式)

**请求参数:**

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "smeeUrl": "https://hook.pipelinesascode.com/GTzCkZZwEGTv",
  "targetUrl": "https://new-target.example.com/webhooks",
  "targetTimeout": 90,
  "httpie": true,
  "ignoreEvents": ["release"],
  "noReplay": false,
  "sseBufferSize": 2097152
}
```

**说明:**

- 如果实例正在运行,需要先停止才能更新配置
- 所有字段与创建时相同

**成功响应 (200):**

返回更新后的完整 client 对象 (同 GET /api/v1/clients/:id)

**错误响应:**

- **400 Bad Request** - 请求参数错误或实例正在运行
- **404 Not Found** - Client 不存在
- **500 Internal Server Error** - 服务器内部错误

---

### DELETE /api/v1/clients/:id

删除 client 实例及其所有数据

**路径参数:**

- `id`: Client ID (UUID 格式)

**说明:**

- 如果实例正在运行,会先停止进程
- 删除实例目录及所有日志、事件数据

**成功响应 (200):**

```json
{
  "message": "Client deleted successfully"
}
```

**错误响应:**

- **404 Not Found** - Client 不存在
- **500 Internal Server Error** - 删除失败

---

### POST /api/v1/clients/:id/start

启动 client 实例

**路径参数:**

- `id`: Client ID (UUID 格式)

**成功响应 (200):**

```json
{
  "message": "Client started successfully"
}
```

**错误响应:**

- **500 Internal Server Error** - 启动失败
  ```json
  {
    "error": "Failed to start client: process already running"
  }
  ```

---

### POST /api/v1/clients/:id/stop

停止 client 实例

**路径参数:**

- `id`: Client ID (UUID 格式)

**说明:**

- 向进程发送 SIGTERM 信号,等待优雅退出 (超时 5 秒)
- 超时后强制 SIGKILL

**成功响应 (200):**

```json
{
  "message": "Client stopped successfully"
}
```

**错误响应:**

- **500 Internal Server Error** - 停止失败

---

### POST /api/v1/clients/:id/restart

重启 client 实例

**路径参数:**

- `id`: Client ID (UUID 格式)

**说明:**

- 先停止进程,然后重新启动
- 等效于依次调用 stop 和 start

**成功响应 (200):**

```json
{
  "message": "Client restarted successfully"
}
```

**错误响应:**

- **500 Internal Server Error** - 重启失败

---

### POST /api/v1/clients/batch/start

批量启动多个 client 实例

**请求参数:**

```json
{
  "clientIds": ["id1", "id2", "id3"],
  "all": false
}
```

**字段说明:**

- `clientIds`: 要启动的 Client ID 数组
- `all`: 是否启动所有实例 (true 时忽略 clientIds)

**成功响应 (200):**

```json
{
  "total": 3,
  "successful": 2,
  "failed": 1,
  "results": [
    {
      "clientId": "id1",
      "success": true
    },
    {
      "clientId": "id2",
      "success": true
    },
    {
      "clientId": "id3",
      "success": false,
      "message": "already running"
    }
  ]
}
```

**错误响应:**

- **400 Bad Request** - clientIds 为空且 all 为 false
- **500 Internal Server Error** - 批量操作失败

---

### POST /api/v1/clients/batch/stop

批量停止多个 client 实例

**请求参数:**

同 `/api/v1/clients/batch/start`

**成功响应 (200):**

同 `/api/v1/clients/batch/start`

**错误响应:**

- **400 Bad Request** - clientIds 为空且 all 为 false
- **500 Internal Server Error** - 批量操作失败

---

### GET /api/v1/clients/:id/stats

获取 client 实例的统计信息

**路径参数:**

- `id`: Client ID (UUID 格式)

**成功响应 (200):**

```json
{
  "todayEvents": 15,
  "totalEvents": 342,
  "successRate": 95.5,
  "averageLatencyMs": 125,
  "uptime": 3600,
  "lastActivity": "2025-10-01T14:23:15Z"
}
```

**字段说明:**

- `todayEvents`: 今日事件数
- `totalEvents`: 总事件数
- `successRate`: 成功率 (百分比)
- `averageLatencyMs`: 平均响应时间 (毫秒)
- `uptime`: 运行时长 (秒)
- `lastActivity`: 最后活动时间

**错误响应:**

- **404 Not Found** - Client 不存在
- **500 Internal Server Error** - 获取统计失败

---

## 日志管理

### GET /api/v1/clients/:id/logs

获取历史日志

**路径参数:**

- `id`: Client ID (UUID 格式)

**查询参数:**

- `date` (可选): 日期 (YYYY-MM-DD 格式),默认今天
- `page` (可选): 页码,默认 1
- `pageSize` (可选): 每页行数,默认 100,最大 1000
- `search` (可选): 搜索关键词

**成功响应 (200):**

```json
{
  "total": 1500,
  "page": 1,
  "pageSize": 100,
  "logs": [
    "[2025-10-01 14:23:10] [INFO] Connected to https://hook.pipelinesascode.com/GTzCkZZwEGTv",
    "[2025-10-01 14:23:15] [INFO] Received event: push (repo: myorg/myrepo)",
    "[2025-10-01 14:23:15] [INFO] Forwarding to https://agola.liu.heiyu.space/webhooks...",
    "[2025-10-01 14:23:16] [INFO] Response: 200 OK (125ms)"
  ]
}
```

**错误响应:**

- **404 Not Found** - Client 不存在
- **500 Internal Server Error** - 获取日志失败

---

### GET /api/v1/clients/:id/logs/stream

实时日志流 (Server-Sent Events)

**路径参数:**

- `id`: Client ID (UUID 格式)

**响应头:**

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Transfer-Encoding: chunked
```

**响应格式 (SSE):**

```
event: log
data: [2025-10-01 14:23:10] [INFO] Connected to https://hook.pipelinesascode.com/GTzCkZZwEGTv

event: log
data: [2025-10-01 14:23:15] [INFO] Received event: push

event: log
data: [2025-10-01 14:23:16] [INFO] Response: 200 OK (125ms)
```

**说明:**

- 使用 EventSource API 接收实时日志
- 连接保持打开直到客户端断开或进程停止

**错误响应:**

- **400 Bad Request** - 无法启动日志流 (例如实例未运行)
- **404 Not Found** - Client 不存在

---

### GET /api/v1/clients/:id/logs/download

下载日志文件

**路径参数:**

- `id`: Client ID (UUID 格式)

**查询参数:**

- `date` (必填): 日期 (YYYY-MM-DD 格式)

**成功响应 (200):**

- 返回文件流
- 响应头:
  ```
  Content-Type: text/plain
  Content-Disposition: attachment; filename="gosmee-{clientId}-{date}.log"
  ```

**错误响应:**

- **400 Bad Request** - date 参数缺失
- **404 Not Found** - Client 不存在或日志文件不存在
- **500 Internal Server Error** - 下载失败

---

## 事件管理

### GET /api/v1/clients/:id/events

获取事件列表

**路径参数:**

- `id`: Client ID (UUID 格式)

**查询参数:**

- `page` (可选): 页码,默认 1
- `pageSize` (可选): 每页数量,默认 20,最大 100
- `eventType` (可选): 按事件类型过滤 (如 push, pull_request)
- `status` (可选): 按状态过滤,可选值: `success`, `failed`, `not_replayed`
- `search` (可选): 在 source 字段中搜索
- `dateFrom` (可选): 开始日期 (ISO 8601)
- `dateTo` (可选): 结束日期 (ISO 8601)
- `sortBy` (可选): 排序字段,默认 `timestamp`
- `sortOrder` (可选): 排序方向,默认 `desc`

**成功响应 (200):**

```json
{
  "total": 342,
  "page": 1,
  "pageSize": 20,
  "events": [
    {
      "id": "evt_abc123",
      "timestamp": "2025-10-01T14:23:15Z",
      "eventType": "push",
      "source": "github.com/myorg/myrepo",
      "status": "success",
      "statusCode": 200,
      "latencyMs": 125
    }
  ]
}
```

**错误响应:**

- **404 Not Found** - Client 不存在
- **500 Internal Server Error** - 获取事件列表失败

---

### GET /api/v1/clients/:id/events/:eventId

获取事件详情

**路径参数:**

- `id`: Client ID (UUID 格式)
- `eventId`: Event ID

**成功响应 (200):**

```json
{
  "id": "evt_abc123",
  "clientId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-10-01T14:23:15Z",
  "eventType": "push",
  "source": "github.com/myorg/myrepo",
  "status": "success",
  "statusCode": 200,
  "latencyMs": 125,
  "headers": {
    "X-GitHub-Event": "push",
    "Content-Type": "application/json",
    "User-Agent": "GitHub-Hookshot/abc123"
  },
  "payload": "{\"ref\":\"refs/heads/main\",\"commits\":[...]}",
  "response": "{\"status\":\"ok\"}",
  "errorMessage": ""
}
```

**字段说明:**

- `headers`: 请求头键值对
- `payload`: 请求体 (JSON 字符串)
- `response`: 响应体 (JSON 字符串)
- `errorMessage`: 错误消息 (仅在失败时有值)

**错误响应:**

- **404 Not Found** - Event 不存在
  ```json
  {
    "error": "Event not found"
  }
  ```

---

### DELETE /api/v1/clients/:id/events/:eventId

删除事件记录

**路径参数:**

- `id`: Client ID (UUID 格式)
- `eventId`: Event ID

**成功响应 (200):**

```json
{
  "message": "Event deleted successfully"
}
```

**错误响应:**

- **404 Not Found** - Event 不存在
- **500 Internal Server Error** - 删除失败

---

### POST /api/v1/clients/:id/events/replay

重放事件到目标 URL

**路径参数:**

- `id`: Client ID (UUID 格式)

**请求参数:**

```json
{
  "eventIds": ["evt_abc123", "evt_def456", "evt_ghi789"]
}
```

**字段说明:**

- `eventIds` (必填): 要重放的事件 ID 数组

**成功响应 (200):**

```json
{
  "total": 3,
  "successful": 2,
  "failed": 1,
  "results": [
    {
      "eventId": "evt_abc123",
      "success": true,
      "statusCode": 200,
      "latencyMs": 150
    },
    {
      "eventId": "evt_def456",
      "success": true,
      "statusCode": 200,
      "latencyMs": 120
    },
    {
      "eventId": "evt_ghi789",
      "success": false,
      "statusCode": 500,
      "latencyMs": 80,
      "errorMessage": "Internal Server Error"
    }
  ]
}
```

**错误响应:**

- **400 Bad Request** - eventIds 为空
- **404 Not Found** - Client 不存在
- **500 Internal Server Error** - 重放失败

---

## 配额管理

### GET /api/v1/quota

获取当前用户的配额信息

**成功响应 (200):**

```json
{
  "quota": {
    "userId": "user-123",
    "totalBytes": 10737418240,
    "usedBytes": 2147483648,
    "percentage": 20.0,
    "clientsCount": 5,
    "maxClients": 50,
    "updatedAt": "2025-10-01T14:30:00Z"
  },
  "warning": "Storage usage is above 80%, please clean up old logs or events"
}
```

**字段说明:**

- `totalBytes`: 总配额 (字节)
- `usedBytes`: 已使用存储 (字节)
- `percentage`: 使用百分比 (0-100)
- `clientsCount`: 当前实例数
- `maxClients`: 最大实例数
- `warning`: 警告信息 (可选,仅在配额超过 80% 时返回)

**错误响应:**

- **500 Internal Server Error** - 获取配额失败

---

## 认证管理

### GET /api/v1/auth/login

跳转到 OIDC Provider 进行认证登录

**说明:**

- 生成随机 state 用于 CSRF 防护
- 将 state 保存到 cookie (10 分钟有效期)
- 重定向到 OIDC Provider 的授权页面
- 仅在启用 OIDC 认证时可用

**响应:**

- **302 Found** - 重定向到 OIDC Provider 授权页面
- **503 Service Unavailable** - OIDC 认证未启用
  ```json
  {
    "error": "OIDC authentication is not enabled"
  }
  ```

---

### GET /api/v1/auth/callback

OIDC 认证回调处理

**查询参数:**

- `code` (必填): 授权码
- `state` (必填): CSRF 防护令牌

**处理流程:**

1. 验证 state 与 cookie 中的 state 是否匹配
2. 使用授权码交换访问令牌和 ID Token
3. 验证 ID Token 签名
4. 提取用户信息 (sub, email, groups)
5. 创建会话并设置 session cookie
6. 重定向到首页

**响应:**

- **302 Found** - 认证成功,重定向到首页 (`/`)
- **400 Bad Request** - State 不匹配或缺少参数
  ```json
  {
    "error": "State mismatch"
  }
  ```
- **500 Internal Server Error** - Token 验证失败或内部错误

---

### POST /api/v1/auth/logout

注销当前用户会话

**说明:**

- 删除服务器端会话
- 清除客户端 session cookie

**成功响应 (200):**

```json
{
  "message": "Logged out successfully"
}
```

---

### GET /api/v1/auth/userinfo

获取当前登录用户信息

**成功响应 (200) - OIDC 已启用且已认证:**

```json
{
  "authenticated": true,
  "oidc_enabled": true,
  "user_id": "user-123",
  "email": "user@example.com",
  "groups": ["ADMIN", "USER"],
  "is_admin": true
}
```

**成功响应 (200) - OIDC 已启用但未认证:**

```json
{
  "authenticated": false,
  "oidc_enabled": true
}
```

**成功响应 (200) - OIDC 未启用:**

```json
{
  "authenticated": false,
  "oidc_enabled": false
}
```

**字段说明:**

- `authenticated`: 是否已认证
- `oidc_enabled`: OIDC 是否启用
- `user_id`: 用户 ID (OIDC sub claim)
- `email`: 用户邮箱
- `groups`: 用户所属组
- `is_admin`: 是否在 ADMIN 组中

---

## 健康检查

### GET /api/v1/health

健康检查接口

**说明:**

- 公共端点,无需认证
- 用于负载均衡器健康检查

**成功响应 (200):**

```json
{
  "status": "healthy",
  "service": "gosmee-webui"
}
```

---

## 错误响应格式

所有错误响应统一使用以下格式:

```json
{
  "error": "错误描述信息"
}
```

HTTP 状态码:

- `400 Bad Request` - 请求参数错误
- `401 Unauthorized` - 未认证 (需要登录)
- `403 Forbidden` - 无权限访问
- `404 Not Found` - 资源不存在
- `500 Internal Server Error` - 服务器内部错误
- `503 Service Unavailable` - 服务不可用

---

## 数据模型

### Client 对象

```typescript
{
  id: string;              // UUID
  userId: string;          // 用户 ID
  name: string;            // 实例名称
  description: string;     // 描述
  status: "running" | "stopped" | "error";

  // Gosmee 配置
  smeeUrl: string;         // Smee 服务器 URL
  targetUrl: string;       // 目标 URL
  targetTimeout: number;   // 超时时间 (秒)
  httpie: boolean;         // 使用 HTTPie 格式
  ignoreEvents: string[];  // 忽略的事件类型
  noReplay: boolean;       // 仅保存不转发
  sseBufferSize: number;   // SSE 缓冲区大小

  // 进程信息
  pid?: number;            // 进程 ID
  startedAt?: string;      // 启动时间 (ISO 8601)
  stoppedAt?: string;      // 停止时间 (ISO 8601)
  restartCount: number;    // 重启次数
  lastError?: string;      // 最后错误

  // 统计
  todayEvents: number;     // 今日事件数
  totalEvents: number;     // 总事件数
  lastActivity?: string;   // 最后活动时间 (ISO 8601)

  // 元数据
  createdAt: string;       // 创建时间 (ISO 8601)
  updatedAt: string;       // 更新时间 (ISO 8601)
}
```

### Event 对象

```typescript
{
  id: string;              // Event ID
  clientId: string;        // Client ID
  timestamp: string;       // 时间戳 (ISO 8601)
  eventType: string;       // 事件类型
  source: string;          // 事件源
  status: "success" | "failed" | "not_replayed";
  statusCode: number;      // HTTP 状态码
  latencyMs: number;       // 延迟 (毫秒)
  headers: Record<string, string>;  // 请求头
  payload: string;         // 请求体 (JSON 字符串)
  response?: string;       // 响应体 (JSON 字符串)
  errorMessage?: string;   // 错误消息
}
```

### Quota 对象

```typescript
{
  userId: string;          // 用户 ID
  totalBytes: number;      // 总配额 (字节)
  usedBytes: number;       // 已使用 (字节)
  percentage: number;      // 使用百分比 (0-100)
  clientsCount: number;    // 当前实例数
  maxClients: number;      // 最大实例数
  updatedAt: string;       // 更新时间 (ISO 8601)
}
```
