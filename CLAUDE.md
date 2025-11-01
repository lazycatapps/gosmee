# Gosmee Web UI 项目说明

## 项目概述

Gosmee Web UI 是一个基于 [gosmee](https://github.com/chmouel/gosmee) 的 Webhook 中继管理平台,提供简洁的 Web UI 界面,用于管理和监控多个 gosmee client 实例。

### 关于 Gosmee

Gosmee 是一个功能强大的 Webhook 中继器,可在任何环境中轻松运行。它通过 client-server 架构实现公网 Webhook 到内网服务的安全转发。

**核心特性**:
- **无需端口转发**: 不需要暴露本地端口,通过 SSE (Server-Sent Events) 拉取事件
- **安全隔离**: Client 主动连接 Server,无需配置入站防火墙规则
- **事件重放**: 支持保存 Webhook 请求历史并重放,便于调试
- **多格式支持**: 可生成 cURL 或 HTTPie 格式的重放脚本
- **实时日志**: 提供完整的请求转发日志和事件查看器

**与传统端口转发的区别**:

| 特性 | Gosmee | 传统端口转发 (如 ngrok) |
|-----|--------|---------------------|
| **安全性** | 无需暴露本地端口,Client 主动拉取 | 需要开放端口,存在安全风险 |
| **防火墙** | 无需配置入站规则 | 需要配置端口映射 |
| **架构** | 中继模式 (SSE 长连接) | 隧道模式 (直接代理) |
| **事件保存** | 原生支持事件存储和重放 | 通常不支持 |
| **调试能力** | Web UI 查看、JSON 解析、脚本生成 | 仅提供日志 |
| **部署灵活性** | 可自建 Server,完全掌控 | 通常依赖第三方服务 |

**典型应用场景**:
1. **本地开发调试**: 将 GitHub/GitLab Webhook 转发到 `localhost:8080`
2. **防火墙内测试**: 中继 Webhook 到私有网络或 VPN 内的服务
3. **事件重放**: 保存 Webhook 历史并反复重放调试处理逻辑
4. **Kubernetes 集成**: Server 部署为公网服务,Client 运行在集群内部转发到内部服务

### 项目目标

本项目旨在为 gosmee client 提供友好的 Web UI 管理界面,让用户可以:

1. **创建和管理多个 gosmee client 实例**
   - 通过 Web 界面配置 gosmee client 参数
   - 启动、停止、删除 client 实例
   - 为每个实例设置友好的名称和描述

2. **实时监控转发状态**
   - 查看每个 client 的运行状态 (运行中/已停止/错误)
   - 实时显示转发日志和事件统计
   - 连接状态监控 (SSE 连接状态、重连次数等)

3. **历史事件管理**
   - 查看和搜索历史转发记录
   - 事件详情查看 (Headers、Payload、响应状态)
   - 事件重放功能 (手动重新发送到目标服务)

4. **多用户隔离**
   - 通过 OIDC 实现用户认证
   - 每个用户独立的 client 实例管理
   - 用户级别的日志存储和配额管理

### 核心价值

- **便捷性**: 无需命令行操作,通过 Web 界面完成所有 Webhook 转发配置
- **可视化**: 友好的事件展示和统计图表,快速了解 Webhook 转发状况
- **可调试**: 保留完整的请求历史,支持事件重放和详细查看
- **多租户**: OIDC 认证实现用户隔离,适合团队协作使用
- **高可用**: 支持同时运行多个 client 实例,互不干扰

## 技术架构

### 后端

- **语言**: Go 1.22+
- **Web 框架**: Gin v1.10+
- **核心依赖**:
  - `github.com/google/uuid` - 实例 ID 生成
  - `github.com/gin-gonic/gin` - HTTP 服务框架
  - `github.com/coreos/go-oidc/v3` - OIDC 认证库
  - `golang.org/x/oauth2` - OAuth2 客户端库
  - `github.com/spf13/cobra` - 命令行参数解析
  - `github.com/spf13/viper` - 配置管理

- **核心工具**: gosmee client (通过命令行调用 `gosmee client` 命令)
  - **进程管理**: 每个 client 实例作为独立进程运行
  - **日志收集**: 捕获 stdout/stderr 并持久化存储
  - **事件存储**: 利用 gosmee 的 `--saveDir` 参数保存事件到用户目录

- **配置管理**: 使用 `github.com/spf13/viper` 读取环境变量
- **日志系统**: 自定义 Logger 接口,基于标准库 `log` 包
  - 支持 INFO/ERROR/DEBUG 三个日志级别
  - 统一的日志格式: `[LEVEL] timestamp message`
  - 输出到 stdout (INFO/DEBUG) 和 stderr (ERROR)

- **中间件**:
  - **CORS 中间件**
    - 默认允许所有来源 (`Access-Control-Allow-Origin: *`)
    - 可以通过环境变量配置为特定域名
    - 支持的方法: GET, POST, PUT, DELETE, OPTIONS
    - 支持的头: Content-Type, Authorization
  - **OIDC 认证中间件**
    - 支持基于 OpenID Connect (OIDC) 的统一认证
    - 自动验证会话 cookie
    - 支持公共端点白名单 (如健康检查、认证回调)
    - API 请求认证失败返回 401,浏览器请求自动跳转登录页

- **会话管理**:
  - 内存会话存储 (SessionService)
  - 会话 TTL: 7 天
  - 自动清理过期会话 (每 10 分钟)
  - 支持会话刷新和注销

### 前端

- **框架**: React 19+
- **UI 库**: Ant Design 5.27+
- **构建工具**: Create React App (react-scripts 5.0+)
- **语言**: JavaScript (非 TypeScript)
- **状态管理**: React Hooks (useState, useEffect, useRef)
  - 不使用额外的状态管理库 (Redux/MobX)
  - 组件级状态管理,适合小型应用

- **HTTP 通信**:
  - 使用浏览器原生 `fetch` API
  - 支持 EventSource (SSE) 接收实时日志流

- **开发环境**:
  - 后端 API 地址配置: 通过环境变量 `BACKEND_API_URL` 注入
    - 开发环境默认: `http://localhost:8080`
    - 生产环境示例: `https://api.example.com`
  - 无需额外的代理配置 (依赖后端 CORS 支持)

## 核心功能

### 1. Gosmee Client 实例管理

#### 1.1 创建 Client 实例

用户通过 Web 表单配置并创建新的 gosmee client 实例:

**基本信息**:
- **实例名称** (必填): 用户友好的名称,用于识别实例
  - 验证规则: 1-50 字符,支持中英文、数字、下划线、短横线
  - 同一用户下名称必须唯一
- **描述** (可选): 实例用途描述,最多 200 字符

**Gosmee 配置** (对应 gosmee client 命令行参数):

- **Smee URL** (必填): gosmee server 的事件源地址
  - 示例: `https://hook.pipelinesascode.com/GTzCkZZwEGTv`
  - 验证: 必须是有效的 HTTPS URL

- **Target URL** (必填): 目标服务的 Webhook 接收地址
  - 示例: `https://agola.liu.heiyu.space/webhooks?agolaid=agola&projectid=xxx`
  - 验证: 必须是有效的 HTTP/HTTPS URL
  - 支持查询参数和路径

- **高级选项** (折叠面板):
  - `--target-connection-timeout`: 目标连接超时时间 (秒),默认 60
  - `--saveDir`: 事件保存目录 (自动设置为用户专属目录,不可编辑)
  - `--httpie`: 是否生成 HTTPie 格式脚本 (布尔开关,默认 false,使用 cURL)
  - `--ignore-event`: 过滤事件类型 (多选下拉框)
    - 可选值: `push`, `pull_request`, `issue_comment`, `release` 等
  - `--noReplay`: 仅保存事件不转发 (布尔开关,默认 false)
  - `--sse-buffer-size`: SSE 缓冲区大小 (字节),默认 1048576

**存储路径设计**:
```
/data/users/{userID}/clients/{clientID}/
  ├── config.json           # 实例配置
  ├── events/              # 事件存储目录 (gosmee --saveDir 指向这里)
  │   ├── 2025-01-15/      # 按日期分组
  │   │   ├── event-001.json
  │   │   ├── event-001.sh
  │   │   └── ...
  │   └── 2025-01-16/
  └── logs/                # 进程日志
      ├── 2025-01-15.log
      └── 2025-01-16.log
```

**创建流程**:
1. 前端提交表单数据到 `POST /api/v1/clients`
2. 后端验证参数合法性
3. 生成唯一的 Client ID (UUID)
4. 创建用户目录结构
5. 保存配置到 `config.json`
6. 启动 gosmee client 进程
7. 返回实例信息和 ID

#### 1.2 查看 Client 实例列表

**列表展示** (Ant Design Table):

| 列名 | 说明 | 示例 |
|-----|------|------|
| 名称 | 用户设置的实例名称 | "Agola Webhook" |
| 状态 | 运行状态 (徽章样式) | 🟢 运行中 / 🔴 已停止 / 🟡 错误 |
| Smee URL | 事件源地址 (省略显示) | `https://hook.pipelines...GTzCkZZwEGTv` |
| Target URL | 目标服务地址 (省略显示) | `https://agola.liu.heiyu.space/web...` |
| 事件计数 | 今日转发事件数 / 总事件数 | 15 / 342 |
| 最后活动 | 最后一次转发时间 | 2 分钟前 |
| 操作 | 启动/停止/查看日志/编辑/删除 | 按钮组 |

**功能**:
- **搜索**: 按实例名称搜索
- **状态过滤**: 全部 / 运行中 / 已停止 / 错误
- **排序**: 按创建时间、最后活动时间排序
- **分页**: 每页 20 条记录

**实时状态更新**:
- 前端通过轮询 (5 秒间隔) 或 WebSocket 获取状态
- 显示进程 PID、运行时长、重启次数

#### 1.3 启动/停止 Client 实例

**启动** (`POST /api/v1/clients/{id}/start`):
- 读取 `config.json` 配置
- 构造 gosmee client 命令行参数
- 使用 `exec.Command` 启动进程
- 捕获 PID 和进程句柄
- 开始收集 stdout/stderr 日志
- 状态变更为 `running`

**gosmee 命令示例**:
```bash
gosmee client \
  --target-connection-timeout 60 \
  --saveDir /data/users/{userID}/clients/{clientID}/events \
  https://hook.pipelinesascode.com/GTzCkZZwEGTv \
  https://agola.liu.heiyu.space/webhooks?agolaid=agola&projectid=xxx
```

**停止** (`POST /api/v1/clients/{id}/stop`):
- 向进程发送 SIGTERM 信号
- 等待进程优雅退出 (超时 5 秒)
- 超时后强制 SIGKILL
- 关闭日志文件句柄
- 状态变更为 `stopped`

#### 1.4 编辑 Client 实例

**可编辑字段**:
- 实例名称、描述
- Target URL (Smee URL 不可更改,因为更改意味着完全不同的事件源)
- 所有高级选项参数

**编辑流程**:
1. 如果实例正在运行,必须先停止
2. 更新 `config.json` 配置
3. 前端提示用户重新启动以应用更改

#### 1.5 删除 Client 实例

**删除流程**:
1. 前端弹出确认对话框
   - 警告: "删除后所有日志和事件历史将永久丢失,是否继续?"
2. 如果实例正在运行,先停止进程
3. 后端删除实例目录 (`/data/users/{userID}/clients/{clientID}/`)
4. 从实例列表中移除

**删除选项**:
- 默认: 删除所有数据 (配置、日志、事件)
- 可选: 仅删除实例配置,保留历史事件 (归档到 `/data/users/{userID}/archive/`)

### 2. 实时日志系统

#### 2.1 日志收集

**进程日志**:
- 捕获 gosmee client 的 stdout 和 stderr
- 按日期分割日志文件: `logs/YYYY-MM-DD.log`
- 日志格式: `[时间戳] [级别] 消息内容`
- 日志轮转: 保留最近 30 天日志,自动清理旧日志

**事件日志** (由 gosmee 自动生成):
- gosmee 通过 `--saveDir` 参数将事件保存为:
  - JSON 文件: 完整的请求/响应数据
  - Shell 脚本: cURL 或 HTTPie 重放脚本
- 按日期分组: `events/YYYY-MM-DD/event-XXX.json`

#### 2.2 实时日志查看

**技术实现**:
- 使用 SSE (Server-Sent Events) 推送实时日志
- API 端点: `GET /api/v1/clients/{id}/logs/stream`
- 前端使用 EventSource API 接收日志流

**显示方式**:
- Modal 弹窗,黑底绿字终端风格
- 自动滚动到最新日志
- 支持暂停自动滚动 (用户主动滚动时)
- 搜索和过滤功能

**日志内容示例**:
```
[2025-01-15 14:23:10] [INFO] Connected to https://hook.pipelinesascode.com/GTzCkZZwEGTv
[2025-01-15 14:23:15] [INFO] Received event: push (repo: myorg/myrepo)
[2025-01-15 14:23:15] [INFO] Forwarding to https://agola.liu.heiyu.space/webhooks...
[2025-01-15 14:23:16] [INFO] Response: 200 OK (125ms)
[2025-01-15 14:23:16] [INFO] Event saved: events/2025-01-15/event-001.json
```

#### 2.3 历史日志查看

**API 端点**: `GET /api/v1/clients/{id}/logs?date=YYYY-MM-DD&page=1&limit=100`

**功能**:
- 按日期选择日志文件
- 分页加载历史日志
- 搜索关键词高亮显示
- 下载日志文件

### 3. 事件历史管理

#### 3.1 事件列表

**数据来源**:
- 读取 `events/` 目录下的 JSON 文件
- 解析事件元数据 (时间、事件类型、来源等)

**列表展示**:

| 列名 | 说明 | 示例 |
|-----|------|------|
| 时间 | 事件接收时间 | 2025-01-15 14:23:15 |
| 事件类型 | GitHub/GitLab 事件类型 | push / pull_request |
| 来源 | 仓库或触发源 | myorg/myrepo |
| 转发状态 | 转发结果 | ✅ 200 OK / ❌ 500 Error / ⏸️ 未转发 |
| 响应时间 | 目标服务响应耗时 | 125ms |
| 操作 | 查看详情/重放 | 按钮组 |

**功能**:
- **日期范围过滤**: 选择起止日期查看事件
- **事件类型过滤**: 按 push、pull_request 等类型筛选
- **状态过滤**: 成功 / 失败 / 未转发
- **搜索**: 按仓库名、事件 ID 搜索
- **批量操作**: 批量删除历史事件

#### 3.2 事件详情

**Modal 弹窗展示**:

**基本信息**:
- 事件 ID
- 时间戳
- 事件类型
- 来源仓库

**请求详情** (Tab 标签页):
- **Headers**: 以表格形式展示所有请求头
- **Payload**: JSON 格式化展示,支持展开/折叠
- **Raw**: 原始 JSON 文本,支持复制

**响应详情**:
- HTTP 状态码
- 响应时间
- 响应 Headers
- 响应 Body (如果有)

**重放脚本**:
- 显示 gosmee 生成的 cURL 或 HTTPie 脚本
- 一键复制到剪贴板

**操作按钮**:
- **重新转发**: 手动触发事件重放到目标服务
- **下载 JSON**: 下载完整的事件 JSON 文件
- **删除**: 删除此事件记录

#### 3.3 事件重放

**手动重放**:
1. 用户点击"重新转发"按钮
2. 后端读取事件 JSON 文件
3. 构造 HTTP 请求发送到 Target URL
4. 记录响应结果
5. 前端显示重放结果 (成功/失败、响应时间、状态码)

**批量重放**:
- 选择多个事件
- 点击"批量重放"
- 显示进度条和成功/失败统计

### 4. 状态监控和统计

#### 4.1 实例状态卡片

**单个实例详情页** (`/clients/{id}`):

**顶部统计卡片** (Ant Design Statistic):
- **运行时长**: 自上次启动以来的时间
- **今日事件**: 今天转发的事件数量
- **总事件数**: 累计转发事件
- **成功率**: 成功转发 / 总事件 × 100%
- **平均响应时间**: 目标服务平均响应耗时

**连接状态**:
- SSE 连接状态: 🟢 已连接 / 🔴 断开 / 🟡 重连中
- 重连次数统计
- 最后接收事件时间

#### 4.2 事件统计图表

**时间序列图** (近 7 天事件趋势):
- X 轴: 日期
- Y 轴: 事件数量
- 折线图显示每天的事件量

**事件类型分布** (饼图):
- 展示不同事件类型 (push、pull_request 等) 的占比

**响应时间分布** (柱状图):
- 展示不同响应时间区间的事件数量
- 区间: <100ms、100-500ms、500ms-1s、1s-5s、>5s

### 5. 配置管理

#### 5.1 用户配置隔离

**目录结构**:
- 启用 OIDC 认证时: `/data/users/{userID}/`
- 未启用 OIDC 时: `/data/default/`
- 每个用户完全隔离,互不影响

**配置限制**:
- **实例数量限制**: 默认每用户最多 50 个实例
  - 可通过 `--max-clients-per-user` 参数配置
- **存储配额**: 默认每用户 10GB
  - 可通过 `--max-storage-per-user` 参数配置
  - 达到 80% 时前端警告
  - 达到 100% 时禁止创建新实例
- **事件保留期**: 默认保留 30 天
  - 可通过 `--event-retention-days` 参数配置
  - 自动清理过期事件

#### 5.2 全局配置

**环境变量**:
- `GOSMEE_HOST`: 服务监听地址,默认 `0.0.0.0`
- `GOSMEE_PORT`: 服务监听端口,默认 `8080`
- `GOSMEE_DATA_DIR`: 数据存储根目录,默认 `/data`
- `GOSMEE_MAX_CLIENTS_PER_USER`: 每用户最大实例数,默认 `50`
- `GOSMEE_MAX_STORAGE_PER_USER`: 每用户存储配额 (字节),默认 `10737418240` (10GB)
- `GOSMEE_EVENT_RETENTION_DAYS`: 事件保留天数,默认 `30`
- `GOSMEE_LOG_RETENTION_DAYS`: 日志保留天数,默认 `30`

### 6. OIDC 认证 (可选)

**认证机制**:
- 支持基于 OpenID Connect (OIDC) 的统一认证
- 与 Lazycat Cloud 认证系统集成
- 自动获取用户 ID、邮箱信息

**配置方式**:
- 环境变量:
  - `GOSMEE_OIDC_CLIENT_ID`
  - `GOSMEE_OIDC_CLIENT_SECRET`
  - `GOSMEE_OIDC_ISSUER`
  - `GOSMEE_OIDC_REDIRECT_URL`
- 当这些环境变量都配置后,OIDC 认证自动启用

**认证端点**:
- `GET /api/v1/auth/login`: 跳转到 OIDC 登录页
- `GET /api/v1/auth/callback`: OIDC 认证回调处理
- `POST /api/v1/auth/logout`: 注销当前用户会话
- `GET /api/v1/auth/userinfo`: 获取当前用户信息

**访问控制**:
- 公共端点: 健康检查、认证相关端点
- 受保护端点: 所有 client 管理相关 API
- 未认证访问: API 返回 401,浏览器跳转登录

## Web UI 设计

### 主页面布局

**顶部导航栏**:
- 应用标题: Gosmee Web UI
- 用户信息 (启用 OIDC 时显示)
- 注销按钮 (启用 OIDC 时显示)

**主要内容区**:
- **Client 实例列表** (默认视图)
  - 卡片样式 + 表格
  - 快速操作按钮: 启动/停止/查看日志
  - 搜索和过滤工具栏

**浮动按钮** (右下角):
- "创建新实例" 按钮 (+ 图标)

### Client 详情页

**页面头部**:
- 返回按钮
- 实例名称和状态徽章
- 操作按钮组: 启动/停止/编辑/删除

**Tab 标签页**:
- **概览**: 统计卡片 + 图表
- **实时日志**: 日志流显示
- **事件历史**: 事件列表
- **配置**: 实例配置详情 (只读,点击"编辑"跳转编辑表单)

### 创建/编辑实例表单

**布局**:
- Modal 对话框 (中等尺寸)
- 分步表单或折叠面板

**表单项**:
1. **基本信息** (总是展开)
   - 实例名称 (Input)
   - 描述 (TextArea)
2. **Gosmee 配置** (总是展开)
   - Smee URL (Input + URL 验证)
   - Target URL (Input + URL 验证)
3. **高级选项** (折叠面板,默认折叠)
   - 连接超时 (InputNumber + 单位 "秒")
   - 脚本格式 (Radio: cURL / HTTPie)
   - 忽略事件类型 (Select 多选)
   - 仅保存不转发 (Switch)
   - SSE 缓冲区大小 (InputNumber)

**按钮**:
- 创建/保存
- 取消

### 事件详情 Modal

**头部**:
- 事件 ID
- 时间戳
- 事件类型徽章

**Tab 标签页**:
- **请求 Headers** (Table)
- **Payload** (JSON 树形展示)
- **响应信息** (状态码、响应时间、响应 Body)
- **重放脚本** (代码块 + 复制按钮)

**底部操作**:
- 重新转发
- 下载 JSON
- 删除
- 关闭

## 技术细节

### 1. 进程管理

**进程生命周期**:
1. **启动**: 使用 `exec.Command` 启动 gosmee client
2. **监控**: 定期检查进程状态 (PID 是否存在)
3. **重启**: 进程意外退出时可选自动重启
4. **停止**: 发送 SIGTERM 信号优雅退出

**进程信息存储**:
```json
{
  "pid": 12345,
  "status": "running",
  "started_at": "2025-01-15T14:23:10Z",
  "restart_count": 0,
  "last_error": null
}
```

**进程监控** (goroutine):
- 每 5 秒检查一次进程状态
- 如果进程退出:
  - 记录退出码和错误信息
  - 更新状态为 `stopped` 或 `error`
  - 如果配置了自动重启且重启次数 < 3,则自动重启

### 2. 日志管理

**日志收集器**:
- 为每个进程创建独立的日志收集 goroutine
- 使用 `io.Pipe` 捕获 stdout/stderr
- 实时写入日志文件
- 广播到所有 SSE 监听器

**日志轮转**:
- 按日期分割: `logs/YYYY-MM-DD.log`
- 每天凌晨 0 点自动创建新日志文件
- 保留策略: 删除超过保留期的日志文件

**日志查询**:
- API: `GET /api/v1/clients/{id}/logs?date=YYYY-MM-DD&search=keyword`
- 支持分页和关键词搜索
- 返回格式: JSON 数组,每条日志一个对象

### 3. 事件存储

**存储格式** (由 gosmee 自动生成):
- JSON 文件: `event-{timestamp}.json`
  ```json
  {
    "id": "evt_abc123",
    "timestamp": "2025-01-15T14:23:15Z",
    "event_type": "push",
    "source": "github.com/myorg/myrepo",
    "headers": { ... },
    "payload": { ... },
    "response": {
      "status_code": 200,
      "latency_ms": 125,
      "body": "..."
    }
  }
  ```
- Shell 脚本: `event-{timestamp}.sh`

**事件索引**:
- 后端定期扫描事件目录
- 构建事件索引 (ID、时间、类型、状态)
- 存储为 SQLite 数据库或 JSON 文件
- 加速查询和过滤

**事件清理**:
- 后台定时任务 (每天凌晨执行)
- 删除超过保留期的事件文件
- 更新索引

### 4. 存储配额管理

**配额计算**:
- 定期扫描用户目录,统计磁盘使用量
- 包含: 事件文件、日志文件、配置文件
- 缓存计算结果 (1 小时有效)

**配额检查**:
- 创建新实例时检查配额
- 事件保存前检查配额
- 超过配额时:
  - 前端显示警告
  - 禁止创建新实例
  - 建议删除旧事件或日志

**配额 API**:
- `GET /api/v1/quota`: 获取当前用户配额信息
  ```json
  {
    "total_bytes": 10737418240,
    "used_bytes": 2147483648,
    "percentage": 20.0,
    "clients_count": 5,
    "max_clients": 50
  }
  ```

## 架构设计

### 分层架构

1. **cmd/server** - 应用入口层
   - 主程序入口
   - 命令行参数解析
   - 服务初始化

2. **handler** - HTTP 处理层
   - ClientHandler: 实例管理接口
   - LogHandler: 日志查询接口
   - EventHandler: 事件历史接口
   - AuthHandler: 认证接口

3. **service** - 业务逻辑层
   - ClientService: 实例生命周期管理
   - ProcessService: gosmee 进程管理
   - LogService: 日志收集和查询
   - EventService: 事件索引和查询
   - QuotaService: 配额计算和检查

4. **repository** - 数据访问层
   - ClientRepository: 实例配置读写
   - EventRepository: 事件索引读写
   - QuotaRepository: 配额数据读写

5. **models** - 数据模型层
   - Client: 实例配置模型
   - Process: 进程状态模型
   - Event: 事件模型
   - Quota: 配额模型

6. **middleware** - 中间件层
   - CORS 中间件
   - OIDC 认证中间件
   - 日志记录中间件

7. **pkg** - 工具包层
   - logger: 日志工具
   - validator: 输入验证
   - fileutil: 文件操作工具

### API 设计

#### Client 管理

```
POST   /api/v1/clients              创建实例
GET    /api/v1/clients              获取实例列表
GET    /api/v1/clients/{id}         获取实例详情
PUT    /api/v1/clients/{id}         更新实例配置
DELETE /api/v1/clients/{id}         删除实例

POST   /api/v1/clients/{id}/start   启动实例
POST   /api/v1/clients/{id}/stop    停止实例
POST   /api/v1/clients/{id}/restart 重启实例
```

#### 日志管理

```
GET /api/v1/clients/{id}/logs/stream         实时日志流 (SSE)
GET /api/v1/clients/{id}/logs?date=YYYY-MM-DD&page=1&limit=100  历史日志
GET /api/v1/clients/{id}/logs/download?date=YYYY-MM-DD         下载日志
```

#### 事件管理

```
GET    /api/v1/clients/{id}/events             事件列表
GET    /api/v1/clients/{id}/events/{eventId}   事件详情
POST   /api/v1/clients/{id}/events/{eventId}/replay  重放事件
DELETE /api/v1/clients/{id}/events/{eventId}   删除事件
```

#### 统计和配额

```
GET /api/v1/clients/{id}/stats   实例统计信息
GET /api/v1/quota                用户配额信息
```

#### 认证

```
GET  /api/v1/auth/login        跳转 OIDC 登录
GET  /api/v1/auth/callback     OIDC 回调
POST /api/v1/auth/logout       注销
GET  /api/v1/auth/userinfo     获取用户信息
```

## 实施路线图

### MVP (v1.0) - 核心功能

**目标**: 提供完整的 gosmee client 实例管理和日志查看功能

**功能清单**:
1. Client 实例管理
   - 创建、启动、停止、删除实例
   - 实例列表和详情展示
   - 基本配置参数支持
2. 实时日志查看
   - SSE 日志流
   - 日志搜索和过滤
3. 进程状态监控
   - 运行状态显示
   - 基本统计信息
4. 用户认证 (可选 OIDC)
   - 用户登录/登出
   - 用户级数据隔离

**技术任务**:
- [ ] 后端实现
  - [ ] gosmee client 进程管理
  - [ ] 日志收集和 SSE 推送
  - [ ] 实例配置存储
  - [ ] OIDC 认证集成
- [ ] 前端实现
  - [ ] 实例列表和创建表单
  - [ ] 实时日志查看器
  - [ ] 状态监控面板
- [ ] 测试
  - [ ] 单元测试
  - [ ] 进程管理集成测试

**时间估算**: 2-3 周

### v1.1 - 事件管理

**目标**: 支持事件历史查看和重放

**功能清单**:
1. 事件历史列表
   - 日期范围过滤
   - 事件类型过滤
   - 状态过滤
2. 事件详情查看
   - Headers、Payload、响应详情
   - 重放脚本展示
3. 事件重放
   - 手动重放单个事件
   - 批量重放
4. 事件清理
   - 自动清理过期事件
   - 手动删除事件

**时间估算**: 1-2 周

### v1.2 - 统计和配额

**目标**: 提供可视化统计和存储配额管理

**功能清单**:
1. 实例统计图表
   - 事件趋势图
   - 事件类型分布
   - 响应时间分布
2. 存储配额管理
   - 配额使用统计
   - 配额告警
   - 存储清理建议
3. 高级配置
   - 自动重启策略
   - 日志保留期配置
   - 事件过滤规则

**时间估算**: 1-2 周

### v2.0 - 企业功能 (未来规划)

**功能清单**:
1. 告警和通知
   - 实例异常告警 (邮件/Webhook)
   - 配额超限通知
2. 批量管理
   - 批量启动/停止实例
   - 配置模板
3. 高级监控
   - 性能指标 (CPU、内存)
   - 事件延迟分析
4. 团队协作
   - 实例共享
   - 权限管理
5. 审计日志
   - 操作审计
   - 访问日志

**时间估算**: 3-4 周

## 开发状态

### 待实现

- [ ] 后端基础框架
- [ ] 进程管理服务
- [ ] 日志收集系统
- [ ] 前端 UI 开发
- [ ] OIDC 认证集成
- [ ] 事件管理功能
- [ ] 统计和配额管理
- [ ] 部署和打包

## 部署方案

### Docker Compose 部署

**最小化部署** (不包含 gosmee server):
```yaml
version: '3.8'
services:
  gosmee-webui:
    image: lazycat/gosmee-webui:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - GOSMEE_DATA_DIR=/data
      - GOSMEE_MAX_CLIENTS_PER_USER=50
      - GOSMEE_EVENT_RETENTION_DAYS=30
```

**完整部署** (包含 gosmee server):
```yaml
version: '3.8'
services:
  gosmee-server:
    image: ghcr.io/chmouel/gosmee:latest
    command: server --port 3000 --public-url https://smee.example.com
    ports:
      - "3000:3000"

  gosmee-webui:
    image: lazycat/gosmee-webui:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - GOSMEE_DATA_DIR=/data
```

### 环境变量配置

**基本配置**:
- `GOSMEE_HOST`: 服务监听地址
- `GOSMEE_PORT`: 服务监听端口
- `GOSMEE_DATA_DIR`: 数据存储目录

**限制配置**:
- `GOSMEE_MAX_CLIENTS_PER_USER`: 每用户最大实例数
- `GOSMEE_MAX_STORAGE_PER_USER`: 每用户存储配额
- `GOSMEE_EVENT_RETENTION_DAYS`: 事件保留天数
- `GOSMEE_LOG_RETENTION_DAYS`: 日志保留天数

**OIDC 配置** (可选):
- `GOSMEE_OIDC_CLIENT_ID`
- `GOSMEE_OIDC_CLIENT_SECRET`
- `GOSMEE_OIDC_ISSUER`
- `GOSMEE_OIDC_REDIRECT_URL`

## 设计评估

### 技术可行性: ⭐⭐⭐⭐⭐ (5/5)

- ✅ gosmee 是成熟的开源项目,稳定可靠
- ✅ 进程管理在 Go 中实现简单 (`os/exec` 包)
- ✅ 日志收集和 SSE 推送技术成熟
- ✅ 文件系统存储简单高效

### 用户体验: ⭐⭐⭐⭐⭐ (5/5)

- ✅ Web UI 降低使用门槛,无需命令行
- ✅ 实时日志和状态监控提升调试效率
- ✅ 事件历史和重放功能强大实用
- ✅ 多实例管理满足复杂场景需求

### 实现复杂度: ⭐⭐⭐ (3/5)

- ⚠️ 进程管理需要处理异常退出、僵尸进程等边缘情况
- ⚠️ 日志收集需要考虑并发和缓冲区管理
- ⚠️ 存储配额计算可能影响性能 (大量文件扫描)
- ✅ 其他功能实现相对简单

**总体评分**: ⭐⭐⭐⭐ (4.3/5) - **设计优秀,可以开始实施**

## 后续优化方向

1. **容器化进程隔离**: 每个 client 进程运行在独立容器中
2. **分布式部署**: 支持多节点部署,负载均衡
3. **事件搜索优化**: 使用 Elasticsearch 加速事件全文搜索
4. **WebSocket 实时更新**: 替代轮询,减少服务器压力
5. **高可用**: 进程故障自动迁移
6. **性能监控**: Prometheus 集成,Grafana 可视化
