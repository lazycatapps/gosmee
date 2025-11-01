# Gosmee Web UI - Webhook 中继管理平台

基于 [gosmee](https://github.com/chmouel/gosmee) 的 Webhook 中继管理平台，提供简洁的 Web UI 界面，用于管理和监控多个 gosmee client 实例。

## 项目简介

Gosmee Web UI 为 gosmee client 提供友好的 Web 管理界面，让您可以通过浏览器轻松管理多个 Webhook 中继实例，无需繁琐的命令行操作。

### 关于 Gosmee

Gosmee 是一个功能强大的 Webhook 中继器，通过 client-server 架构实现公网 Webhook 到内网服务的安全转发。

**核心特性**:
- 🔒 **无需端口转发**: 不需要暴露本地端口，通过 SSE (Server-Sent Events) 拉取事件
- 🛡️ **安全隔离**: Client 主动连接 Server，无需配置入站防火墙规则
- 🔄 **事件重放**: 支持保存 Webhook 请求历史并重放，便于调试
- 📝 **多格式支持**: 可生成 cURL 或 HTTPie 格式的重放脚本
- 📊 **实时日志**: 提供完整的请求转发日志和事件查看器

**典型应用场景**:
1. 本地开发调试：将 GitHub/GitLab Webhook 转发到 `localhost:8080`
2. 防火墙内测试：中继 Webhook 到私有网络或 VPN 内的服务
3. 事件重放：保存 Webhook 历史并反复重放调试处理逻辑
4. Kubernetes 集成：Server 部署为公网服务，Client 运行在集群内部

## 功能特性

- 🎯 **可视化实例管理**: 创建、启动、停止、删除 gosmee client 实例
- 📊 **实时状态监控**: 查看运行状态、事件统计、连接状态
- 📋 **实时日志查看**: SSE 推送实时日志，支持搜索和过滤
- 📚 **事件历史管理**: 查看和搜索历史转发记录，支持事件重放
- 🔐 **多用户隔离**: 支持 OIDC 认证，每个用户独立管理实例
- 💾 **配额管理**: 存储配额监控和自动清理
- ⚡ **前后端分离**: 易于部署和扩展

## 技术栈

**后端:**
- Go 1.25+
- Gin Web Framework v1.11+
- gosmee client (进程管理)
- Cobra (命令行参数解析)
- Viper (配置管理)
- OIDC 认证支持 (可选)

**前端:**
- React 19.2+
- Ant Design 5.27+
- JavaScript (非 TypeScript)

## 快速开始

### 前置要求

#### 部署 Gosmee Server

本项目使用 **Client-Server 模式**，需要先部署 Gosmee Server。

**Docker Compose 部署（推荐）:**

```yaml
version: '3.8'
services:
  gosmee-server:
    image: ghcr.io/chmouel/gosmee:latest
    command: server --port 3000 --public-url https://smee.example.com
    ports:
      - "3000:3000"
    restart: unless-stopped
```

启动服务:
```bash
docker-compose up -d
```

### 方式一：本地开发运行

#### 1. 启动后端服务

```bash
# 使用 Makefile（推荐）
make dev-backend

# 或手动启动
cd backend
go mod download
export GOSMEE_DATA_DIR="/tmp/gosmee-data"
export GOSMEE_MAX_CLIENTS_PER_USER=50
go run cmd/server/main.go
```

后端服务默认运行在 `http://localhost:8080`

可以通过 `-p` 或 `--port` 参数指定端口：

```bash
go run cmd/server/main.go --port 9090
```

#### 2. 启动前端服务

```bash
# 使用 Makefile（推荐）
make dev-frontend

# 或手动启动
cd frontend
npm install
npm start
```

前端服务默认运行在 `http://localhost:3000`

**配置说明：**

后端支持通过环境变量或命令行参数配置。主要配置项：
- `--data-dir`: 数据存储根目录，默认 `/data`
- `--max-clients-per-user`: 每用户最大实例数，默认 `50`
- `--max-storage-per-user`: 每用户存储配额（字节），默认 `10737418240` (10GB)
- `--event-retention-days`: 事件保留天数，默认 `30`
- `--log-retention-days`: 日志保留天数，默认 `30`

环境变量格式：`GOSMEE_` + 参数名（横线替换为下划线），例如 `GOSMEE_DATA_DIR`

### 方式二：LPK 部署（推荐用于 Lazycat Cloud 平台）

#### 1. 构建前端

```bash
make build-frontend
# 或手动构建
sh build.sh
```

#### 2. 构建后端镜像

```bash
make push-backend
# 这会构建生产环境后端镜像并推送到 registry
```

#### 3. 构建 LPK 包

```bash
make build-lpk
# 或手动构建
lzc build
```

#### 4. 部署 LPK

将生成的 `.lpk` 文件上传到 Lazycat Cloud 平台进行部署。

**环境变量说明：**

后端环境变量：
- `GOSMEE_DATA_DIR`: 数据存储根目录（必需）
- `GOSMEE_MAX_CLIENTS_PER_USER`: 每用户最大实例数，默认 `50`
- `GOSMEE_MAX_STORAGE_PER_USER`: 每用户存储配额（字节），默认 `10737418240` (10GB)
- `GOSMEE_EVENT_RETENTION_DAYS`: 事件保留天数，默认 `30`
- `GOSMEE_LOG_RETENTION_DAYS`: 日志保留天数，默认 `30`

OIDC 认证环境变量（可选）：
- `GOSMEE_OIDC_CLIENT_ID=${LAZYCAT_AUTH_OIDC_CLIENT_ID}`
- `GOSMEE_OIDC_CLIENT_SECRET=${LAZYCAT_AUTH_OIDC_CLIENT_SECRET}`
- `GOSMEE_OIDC_ISSUER=${LAZYCAT_AUTH_OIDC_ISSUER}`
- `GOSMEE_OIDC_REDIRECT_URL=https://${LAZYCAT_APP_DOMAIN}/api/v1/auth/callback`

### 使用说明

1. 打开浏览器访问前端地址（如 `http://localhost:3000`）
2. 点击"创建新实例"按钮
3. 填写实例配置：
   - **实例名称**: 用户友好的名称
   - **描述**: 实例用途描述（可选）
   - **Smee URL**: gosmee server 的事件源地址（例如：`https://hook.pipelinesascode.com/GTzCkZZwEGTv`）
   - **Target URL**: 目标服务的 Webhook 接收地址（例如：`https://your-service.com/webhooks`）
   - **高级选项**（可选）:
     - 连接超时时间
     - 脚本格式（cURL / HTTPie）
     - 忽略事件类型
     - 其他 gosmee client 参数
4. 点击"创建"按钮
5. 实例创建后，点击"启动"开始转发 Webhook
6. 查看实时日志和事件历史

**功能特性**:
- 实例列表：查看所有实例的状态、统计信息
- 实时日志：通过 SSE 查看实时转发日志
- 事件历史：查看和搜索历史事件，支持按日期、类型、状态过滤
- 事件重放：手动重新发送历史事件到目标服务
- 配额监控：查看存储使用情况和配额限制

## 项目结构

```
gosmee-webui/
├── backend/                   # Go 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        # 应用入口
│   ├── internal/              # 内部包
│   │   ├── models/            # 数据模型
│   │   ├── types/             # 类型定义
│   │   ├── repository/        # 数据访问层
│   │   ├── service/           # 业务逻辑层
│   │   ├── handler/           # HTTP 处理层
│   │   ├── middleware/        # 中间件
│   │   ├── router/            # 路由管理
│   │   └── pkg/               # 工具包
│   ├── go.mod
│   ├── Dockerfile
│   └── .gitignore
├── frontend/                  # React 前端
│   ├── src/
│   │   ├── App.js             # 主组件
│   │   └── App.css            # 样式
│   ├── package.json
│   └── .gitignore
├── dist/                      # 构建输出目录（由 build.sh 生成）
│   └── web/                   # 前端静态文件
├── build.sh                   # 前端构建脚本
├── lzc-build.yml              # LPK 构建配置
├── lzc-manifest.yml           # LPK 应用清单
├── icon.png                   # 应用图标
├── Makefile                   # 构建命令
├── CLAUDE.md                  # Claude 参考文档
└── README.md                  # 项目说明
```

## API 接口

### Client 管理

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

### 日志管理

```
GET /api/v1/clients/{id}/logs/stream         实时日志流 (SSE)
GET /api/v1/clients/{id}/logs?date=YYYY-MM-DD&page=1&limit=100  历史日志
GET /api/v1/clients/{id}/logs/download?date=YYYY-MM-DD         下载日志
```

### 事件管理

```
GET    /api/v1/clients/{id}/events             事件列表
GET    /api/v1/clients/{id}/events/{eventId}   事件详情
POST   /api/v1/clients/{id}/events/{eventId}/replay  重放事件
DELETE /api/v1/clients/{id}/events/{eventId}   删除事件
```

### 统计和配额

```
GET /api/v1/clients/{id}/stats   实例统计信息
GET /api/v1/quota                用户配额信息
```

### 认证（OIDC）

```
GET  /api/v1/auth/login        跳转 OIDC 登录
GET  /api/v1/auth/callback     OIDC 回调
POST /api/v1/auth/logout       注销
GET  /api/v1/auth/userinfo     获取用户信息
```

详细 API 文档请参考 [API.md](API.md)

## Makefile 命令

```bash
make help            # 显示所有可用命令
make dev-backend     # 启动后端开发服务
make dev-frontend    # 启动前端开发服务
make build-frontend  # 构建前端到 dist 目录
make build-local-bin # 本地编译后端二进制
make push-backend    # 构建并推送生产环境后端镜像
make push-backend-dev # 构建并推送开发环境后端镜像
make build-lpk       # 构建 LPK 包（需要 lzc-cli）
make deploy          # 生产部署（后端生产镜像 + 前端 + lpk）
make deploy-frontend # 部署前端（前端 + lpk）
make deploy-backend-dev # 部署后端开发版（后端开发镜像 + lpk）
make deploy-full-dev # 完整开发部署（后端开发镜像 + 前端 + lpk）
make audit           # 扫描前端依赖漏洞
make clean           # 清理构建输出
```

## 开发计划

- [x] 基础 Web UI 界面
- [x] 实例管理功能（创建、启动、停止、删除）
- [x] 实时日志查看功能（SSE）
- [x] 进程管理和监控
- [x] OIDC 认证支持
- [x] 后端分层架构
- [x] LPK 打包支持
- [ ] 事件历史管理（列表、搜索、过滤）
- [ ] 事件详情查看
- [ ] 事件重放功能
- [ ] 统计图表（事件趋势、类型分布）
- [ ] 配额管理和告警
- [ ] 自动清理过期数据
- [ ] 批量管理功能

## 文档

- [快速开始](QUICKSTART.md) - 5 分钟快速上手指南
- [API 文档](API.md) - 完整的 API 接口文档
- [Claude 开发文档](CLAUDE.md) - 项目技术架构和开发指南

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！
