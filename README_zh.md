# Home Provider

[English](./README.md)

API 代理服务器，通过可配置的提供商后端路由 LLM 请求（兼容 OpenAI + Anthropic），提供 API 密钥管理和使用统计。

## 功能特性

- **OpenAI 兼容 API** — `POST /v1/chat/completions`
- **Anthropic Messages API** — `POST /v1/messages`
- **虚拟模型路由** — 在不同机器和 Agent 间使用统一的虚拟模型名（如 `latest`）
- **API 密钥管理** — 生成、管理 API 密钥，支持 AES-GCM 加密
- **使用统计** — 按密钥和全局的统计数据，带时间序列图表
- **多语言支持** — 中文和英文

## 快速开始

### 1. 构建服务器

```bash
go build -o server ./cmd/server
```

### 2. 配置环境

```bash
cp .env.example .env
```

编辑 `.env`，设置 `ENCRYPTION_KEY`（用于 AES-GCM 加密的 32 字节密钥）。

### 3. 运行

```bash
./server
```

### 4. 访问 Web UI

打开 http://localhost:18427

## 使用方法

### 1. 添加提供商

进入 **Providers** 标签页 → 添加提供商：

- **名称**：如 `MiniMax`
- **端点**：如 `https://api.minimax.io`
- **API 密钥**：你的提供商 API 密钥
- **模型**：如 `"MiniMax-M2.7-highspeed"`

### 2. 创建虚拟模型

进入 **Virtual Models** 标签页 → 创建虚拟模型：

- **名称**：如 `latest`
- **提供商**：选择你添加的提供商

### 3. 生成 API 密钥

进入 **API Keys** 标签页 → 生成密钥

### 4. 发起请求

```bash
# Anthropic Messages API
curl -X POST http://127.0.0.1:18427/v1/messages \
  -H "Authorization: Bearer hpk_xxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "latest",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 1024
  }'

# OpenAI 兼容 API
curl -X POST http://127.0.0.1:18427/v1/chat/completions \
  -H "Authorization: Bearer hpk_xxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "latest",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

## 工作原理

```
客户端 → API 密钥 → 虚拟模型 → 提供商 → AI 响应
```

1. 客户端发送请求，包含 API 密钥和虚拟模型名（如 `latest`）
2. 验证 API 密钥
3. 将虚拟模型解析为具体的提供商和模型
4. 请求转发给提供商
5. 响应返回给客户端

## 环境变量

| 变量             | 默认值 | 说明                        |
| ---------------- | ------ | --------------------------- |
| `PORT`           | 18427  | 服务器端口                  |
| `ENCRYPTION_KEY` | —      | AES-GCM 32 字节密钥（必填） |
| `LOG_LEVEL`      | info   | debug\|info\|warn\|error    |
| `DATA_DIR`       | `~/.config/home-provider` | 数据目录        |

## 开发

```bash
# 构建前端
cd web && npm install && npm run build

# 运行测试
go test ./...

# 格式化代码
gofmt -w ./internal/
```
