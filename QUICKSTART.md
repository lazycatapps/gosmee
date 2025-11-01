# 快速开始

欢迎使用 Gosmee Web UI！本指南将帮助您快速上手,将公网 Webhook 转发到您的内网服务。

## 什么是 Webhook 中继?

当您的内网服务需要接收来自公网服务(如 GitHub、GitLab)的 Webhook 回调时,通常需要:
- 暴露公网 IP 或域名
- 配置防火墙端口转发
- 处理复杂的网络配置

**Gosmee 提供了更简单的方案**:
1. 在公网 Webhook 中继服务(如 Smee.io)上生成一个公网地址
2. 将 GitHub 等服务的 Webhook 配置指向这个公网地址
3. 使用 Gosmee 将请求转发到您的内网服务
4. 无需暴露端口,无需配置防火墙

## 准备工作

### 1. 获取公网中继地址

您可以使用以下任一服务生成免费的公网中继地址:

#### 方案一: Pipelines as Code Smee 服务 (推荐)

访问 [https://hook.pipelinesascode.com/](https://hook.pipelinesascode.com/)

页面会**自动生成**一个唯一的公网 URL,例如:
```
https://hook.pipelinesascode.com/bhlSZRHGVKbh
```

直接复制这个 URL 即可,无需任何操作。

#### 方案二: Smee.io 服务

访问 [https://smee.io/](https://smee.io/)

1. 点击 "Start a new channel" 按钮
2. 系统会生成一个唯一的 URL,例如:
   ```
   https://smee.io/abc123xyz
   ```
3. 复制这个 URL

### 2. 确认您的目标服务地址

确认需要接收 Webhook 的目标服务地址。目标可以是:

**内网服务** (生产场景):
- `http://localhost:8080/webhooks`
- `http://192.168.1.100:3000/api/github/webhook`
- `http://agola.internal.com/webhooks?projectid=myproject`

**另一个公网中继地址** (本地测试场景):
- 如果您暂时没有内网服务,可以使用另一个 Smee URL 作为目标地址
- 例如: 从 `https://hook.pipelinesascode.com/abc123` 转发到 `https://smee.io/xyz789`
- 这样可以在两个公网服务之间测试转发功能

## 使用步骤

### 第一步: 创建 Gosmee Client 实例

1. 打开 Gosmee Web UI (通常是 `https://gosmee.{box-name}.heiyu.space/`)

2. 点击右下角的 **"创建新实例"** 按钮 (+ 图标)

3. 填写表单:

   **基本信息**:
   - **实例名称**: 给这个实例起一个便于识别的名字,例如 "GitHub Webhook"
   - **描述**: (可选) 例如 "用于接收 GitHub push 事件"

   **Gosmee 配置**:
   - **Smee URL**: 粘贴第一步获取的公网地址
     ```
     https://hook.pipelinesascode.com/GTzCkZZwEGTv
     ```
   - **Target URL**: 填写您的内网服务地址
     ```
     http://localhost:8080/webhooks
     ```

4. 点击 **"创建"** 按钮

### 第二步: 启动实例

1. 在实例列表中找到刚创建的实例

2. 点击 **"启动"** 按钮

3. 实例状态变为 🟢 **运行中**

4. 此时 Gosmee 已经开始监听公网地址,并会自动将收到的请求转发到您的内网服务

### 第三步: 配置 GitHub Webhook

1. 打开您的 GitHub 仓库

2. 进入 **Settings** → **Webhooks** → **Add webhook**

3. 填写 Webhook 配置:
   - **Payload URL**: 粘贴您的 Smee URL
     ```
     https://hook.pipelinesascode.com/GTzCkZZwEGTv
     ```
   - **Content type**: 选择 `application/json`
   - **Events**: 选择您需要的事件 (如 Push events, Pull request events)

4. 点击 **"Add webhook"** 保存

### 第四步: 测试转发

有多种方式可以测试转发功能:

#### 方法一: 使用 curl 直接触发 (推荐用于快速测试)

在终端中执行以下命令,向您的 Smee URL 发送测试请求:

```bash
curl -X POST https://hook.pipelinesascode.com/bhlSZRHGVKbh \
  -H "Content-Type: application/json" \
  -d '{"event": "test", "message": "Hello from Gosmee!"}'
```

将 URL 替换为您自己的 Smee URL。

#### 方法二: 通过 GitHub 触发真实事件

1. 在您的 GitHub 仓库中触发一个事件 (例如 push 一个 commit)
2. GitHub 会自动发送 Webhook 到您配置的 Smee URL

#### 查看转发结果

1. 回到 Gosmee Web UI,点击实例的 **"查看日志"** 按钮

2. 您应该能看到类似这样的日志:
   ```
   [2025-01-15 14:23:15] [INFO] Received event: test
   [2025-01-15 14:23:15] [INFO] Forwarding to http://localhost:8080/webhooks
   [2025-01-15 14:23:16] [INFO] Response: 200 OK (125ms)
   ```

3. 同时检查您的目标服务日志,确认收到了 Webhook 请求

## 常见使用场景

### 场景一: 本地测试转发功能

**需求**: 在没有内网服务的情况下测试 Gosmee 转发功能

**配置**:
- Smee URL: `https://hook.pipelinesascode.com/abc123`
- Target URL: `https://smee.io/xyz789` (另一个公网中继地址)

**测试步骤**:
1. 使用 curl 向 Smee URL 发送请求:
   ```bash
   curl -X POST https://hook.pipelinesascode.com/abc123 \
     -H "Content-Type: application/json" \
     -d '{"test": "data"}'
   ```
2. 在浏览器中打开 Target URL (`https://smee.io/xyz789`)
3. 可以在 Smee.io 页面上看到转发过来的请求

**用途**: 快速验证 Gosmee 工作是否正常,无需搭建本地服务

### 场景二: 本地开发调试

**需求**: 在本地开发环境测试 GitHub Webhook

**配置**:
- Smee URL: `https://hook.pipelinesascode.com/abc456`
- Target URL: `http://localhost:3000/api/webhook`

**用途**: 无需部署到服务器,直接在本地调试 Webhook 处理逻辑

### 场景三: 内网 CI/CD 服务

**需求**: 将 GitHub 事件转发到内网的 CI/CD 服务 (如 Agola、Jenkins)

**配置**:
- Smee URL: `https://hook.pipelinesascode.com/xyz789`
- Target URL: `http://agola.internal.com/webhooks?projectid=myapp`

**用途**: 内网 CI/CD 服务可以接收 GitHub 事件并自动触发构建

### 场景四: 防火墙内的测试环境

**需求**: 公司网络有严格的防火墙策略,无法开放端口

**配置**:
- Smee URL: `https://smee.io/test456`
- Target URL: `http://192.168.1.50:8080/webhooks`

**用途**: 无需配置防火墙规则,Gosmee 主动拉取事件并转发到内网

## 高级功能

### 查看事件历史

1. 点击实例名称进入详情页

2. 切换到 **"事件历史"** 标签页

3. 您可以:
   - 查看所有转发的请求和响应
   - 搜索特定事件
   - 查看完整的 Headers 和 Payload
   - 手动重放某个事件

### 事件重放

如果您需要重新测试某个 Webhook 请求:

1. 在事件历史中找到目标事件

2. 点击 **"查看详情"**

3. 点击 **"重新转发"** 按钮

4. Gosmee 会使用相同的数据再次发送到您的内网服务

### 过滤特定事件

如果您只想转发特定类型的 GitHub 事件:

1. 编辑实例配置

2. 展开 **"高级选项"**

3. 在 **"忽略事件类型"** 中选择不需要的事件

4. 例如: 忽略 `issue_comment` 和 `release` 事件

## 故障排查

### 问题: 实例显示 "错误" 状态

**解决**:
1. 点击 **"查看日志"** 查看错误信息
2. 常见原因:
   - Smee URL 无效或已过期 (重新生成)
   - Target URL 无法访问 (检查内网服务是否运行)
   - 网络连接问题 (检查防火墙出站规则)

### 问题: 收不到 Webhook 请求

**检查清单**:
- ✅ Gosmee 实例状态为 "运行中"
- ✅ GitHub Webhook 配置正确 (URL 匹配 Smee URL)
- ✅ GitHub Webhook 最近交付中显示成功 (200 OK)
- ✅ 内网服务正常运行且端口正确

### 问题: 转发成功但内网服务没有响应

**排查**:
1. 在 Gosmee 日志中确认收到了请求
2. 检查响应状态码 (如果是 404,说明 Target URL 路径错误)
3. 查看内网服务的日志
4. 使用 **事件重放** 功能反复测试

## 多实例管理

您可以创建多个 Gosmee 实例,每个实例对应不同的用途:

**示例**:
- **实例 1**: GitHub 项目 A → Agola
  - Smee URL: `https://smee.io/projectA`
  - Target URL: `http://agola.com/webhooks?project=A`

- **实例 2**: GitHub 项目 B → Jenkins
  - Smee URL: `https://smee.io/projectB`
  - Target URL: `http://jenkins.internal:8080/github-webhook/`

- **实例 3**: GitLab → 自定义服务
  - Smee URL: `https://hook.pipelinesascode.com/gitlab123`
  - Target URL: `http://localhost:3000/api/gitlab`

所有实例互不干扰,可以独立启动、停止和管理。

## 需要帮助?

如果遇到问题,请:
1. 查看实例日志获取详细错误信息
2. 访问 [GitHub Issues](https://github.com/lazycatapps/gosmee/issues) 提交问题
3. 查看 [gosmee 官方文档](https://github.com/chmouel/gosmee) 了解更多细节
