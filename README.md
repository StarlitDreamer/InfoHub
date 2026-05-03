# InfoHub Agent

InfoHub Agent 是一个信息汇总与决策辅助 Agent，用于从 RSS 等来源采集信息，完成去重、AI 摘要、评分、日报生成、文件落盘，并通过 CLI、定时任务或 HTTP API 对外提供能力。

## 当前能力

- 支持 Demo 数据源和多 RSS 数据源聚合。
- 支持单次运行、定时运行和 Gin HTTP 服务模式。
- 支持 mock AI 和 OpenAI 兼容 Chat Completions 接口。
- 支持 Markdown 日报输出和 Webhook 推送。
- 支持文件型日报存储，保存 Markdown 和 NewsItem JSON。
- 支持跨运行去重，默认使用本地文件保存已处理指纹。
- 支持 JSON 配置文件和环境变量配置，环境变量优先级更高。

## 快速开始

运行测试：

```bash
go test ./...
```

生成一次日报：

```bash
go run cmd/main.go run-once
```

启动定时任务：

```bash
go run cmd/main.go schedule
```

启动 HTTP 服务：

```bash
go run cmd/main.go serve
```

未传运行模式时，默认等同于 `run-once`。

## 配置方式

推荐使用 JSON 配置文件：

```bash
INFOHUB_CONFIG_PATH=configs/config.example.json go run cmd/main.go run-once
```

Windows PowerShell 示例：

```powershell
$env:INFOHUB_CONFIG_PATH="D:\code\go\InfoHub\configs\config.example.json"
go run cmd\main.go run-once
```

配置文件示例见 [configs/config.example.json](configs/config.example.json)。

也可以使用环境变量配置，示例见 [configs/config.example.env](configs/config.example.env)。当 JSON 配置和环境变量同时存在时，环境变量会覆盖 JSON 配置。

常用环境变量：

- `INFOHUB_CONFIG_PATH`：JSON 配置文件路径。
- `INFOHUB_RSS_URL`：单个 RSS 数据源。
- `INFOHUB_RSS_URLS`：多个 RSS 数据源，逗号分隔，优先于 `INFOHUB_RSS_URL`。
- `INFOHUB_AI_ENDPOINT`：OpenAI 兼容接口地址。
- `INFOHUB_AI_API_KEY`：AI API Key。
- `INFOHUB_AI_MODEL`：AI 模型名。
- `INFOHUB_WEBHOOK_URL`：Webhook 推送地址。
- `INFOHUB_SEND_EMPTY_REPORT`：无新增信息时是否仍推送 Webhook，默认 `false`。
- `INFOHUB_STORAGE_DIR`：日报和 JSON 快照保存目录，默认 `data/reports`。
- `INFOHUB_DEDUP_STORE_PATH`：跨运行去重状态文件，默认 `data/dedup/seen.json`。
- `INFOHUB_HTTP_ADDR`：Gin HTTP 服务监听地址，默认 `:8080`。
- `INFOHUB_AUTH_TOKEN`：Gin API 鉴权 token，留空时不启用鉴权。
- `INFOHUB_SCHEDULE_INTERVAL_SECONDS`：定时任务间隔秒数，默认 `3600`。

## HTTP API

启动服务：

```bash
go run cmd/main.go serve
```

如果配置了 `INFOHUB_AUTH_TOKEN` 或 JSON 中的 `auth.token`，除 `/health` 外的接口都需要请求头：

```http
Authorization: Bearer <token>
```

健康检查：

```http
GET /health
```

手动生成日报：

```http
POST /reports/run
```

响应示例：

```json
{
  "status": "generated",
  "item_count": 3
}
```

读取最新日报：

```http
GET /reports/latest
```

列出历史日报：

```http
GET /reports
```

## 数据落盘

默认目录：

```text
data/reports/
├── reports/
│   └── 20260503-153000.md
└── items/
    └── 20260503-153000.json
```

跨运行去重状态默认保存到：

```text
data/dedup/seen.json
```

这些运行数据默认被 `.gitignore` 忽略。

## 去重行为

系统会先做单次运行内标题去重，再做跨运行去重。跨运行去重 key 优先使用 URL；URL 为空时使用标题，并通过 SHA-256 生成稳定指纹。

当没有新增信息时，日报会输出：

```md
# 今日信息

今日暂无新增信息。
```

默认不会推送空日报。如需推送，设置：

```bash
INFOHUB_SEND_EMPTY_REPORT=true
```

## 当前限制

- RSS 解析目前覆盖基础 RSS 字段，尚未支持复杂 Feed 清洗。
- mock AI 可用于本地开发；真实摘要需要配置 OpenAI 兼容接口。
- 当前存储为文件版，尚未接入 MySQL。
- 当前跨运行去重为文件版，尚未接入 Redis。
- Gin API 支持 Bearer Token 鉴权，部署到公网前仍建议放在反向代理或内网访问控制之后。

## 建议下一步

- 接入 Redis 版跨运行去重。
- 接入 MySQL repository。
- 增强 RSS 内容清洗和 HTML 摘要提取。
- 增加 Dockerfile 和部署示例。
