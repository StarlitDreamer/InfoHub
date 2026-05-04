# InfoHub Agent

[English](./README.md) | [简体中文](./README.zh.md)

InfoHub Agent 是一个信息汇总与决策支持服务。它会从 RSS 与演示数据源采集内容，完成去重、AI 摘要与评分，并输出可直接辅助决策的日报。

## 当前能力

- Demo 采集器与多 RSS 聚合
- Mock AI 处理器与兼容 OpenAI 的 Chat Completions 客户端
- 单次运行内与跨运行去重
- Markdown 报告渲染
- 基于文件或 MySQL 的报告存储
- Webhook 与邮件投递
- 手动执行、定时调度模式与 Gin HTTP API
- JSON 配置与环境变量覆盖
- 基于 Redis 的去重存储

## 项目结构

```text
cmd/
configs/
internal/
  ai/
  config/
  crawler/
  delivery/
  model/
  processor/
  repository/
  scheduler/
  server/
  service/
scripts/
```

## 快速开始

运行测试：

```bash
go test ./...
```

如果你的机器不能使用默认 Go 构建缓存目录，请先在工作区内设置 `GOCACHE`：

```powershell
$env:GOCACHE="D:\code\go\InfoHub\.gocache"
go test ./...
```

针对本地 Compose 数据库运行 MySQL 集成测试：

```bash
INFOHUB_TEST_MYSQL_DSN="infohub:infohub@tcp(localhost:3307)/infohub?charset=utf8mb4&parseTime=true&loc=Local" go test ./internal/repository -run TestMySQLReportRepositoryIntegration
```

针对本地 Compose 缓存运行 Redis 集成测试：

```bash
INFOHUB_TEST_REDIS_ADDR="localhost:6379" go test ./internal/processor -run TestRedisDedupStoreIntegration
```

生成一份报告：

```bash
go run cmd/main.go run-once
```

启动定时调度模式：

```bash
go run cmd/main.go schedule
```

调度器支持两种方式：通过 `INFOHUB_SCHEDULE_INTERVAL_SECONDS` 指定固定间隔，或通过 `INFOHUB_SCHEDULE_CRON` 指定 5 段 cron 表达式，例如 `0 9,18 * * 1-5`。

启动 HTTP 服务：

```bash
go run cmd/main.go serve
```

如果没有传入模式参数，应用默认使用 `run-once`。

## 验证流程

### 1. 使用 Mock AI 与文件存储进行快速本地验证

这是主机侧最安全的验证路径，因为它不依赖 MySQL，只会写入本地文件：

```powershell
$env:GOCACHE="D:\code\go\InfoHub\.gocache"
$env:INFOHUB_RSS_URLS="https://blog.google/rss/,https://openai.com/news/rss.xml"
$env:INFOHUB_RSS_MAX_ITEMS_PER_FEED="15"
$env:INFOHUB_RSS_RECENT_WITHIN_HOURS="168"
$env:INFOHUB_REPORT_MAX_ITEMS="12"
$env:INFOHUB_STORAGE_DIR="D:\code\go\InfoHub\data\reports-verify"
$env:INFOHUB_DEDUP_STORE_PATH="D:\code\go\InfoHub\data\dedup\verify-seen.json"
go run cmd\main.go run-once
```

这条路径会使用：

- 真实 RSS 源
- Mock AI 评分与摘要
- 基于文件的报告存储
- 用于重复验证的临时去重文件

### 2. Docker Compose 验证

如果你希望把 MySQL 和 Redis 一起纳入验证，请使用：

```bash
docker compose up --build
```

### 3. 真实 AI 验证

如果要接入真实的兼容 OpenAI 的端点，请设置：

- `INFOHUB_AI_ENDPOINT`
- `INFOHUB_AI_API_KEY`
- `INFOHUB_AI_MODEL`

如果这些变量缺失，应用会回退到内置 Mock AI 处理器。

## 配置

使用仓库内置的 JSON 示例：

```bash
INFOHUB_CONFIG_PATH=configs/config.example.json go run cmd/main.go run-once
```

PowerShell 示例：

```powershell
$env:INFOHUB_CONFIG_PATH="D:\code\GolandProjects\InfoHub\configs\config.example.json"
go run cmd\main.go run-once
```

示例文件如下：

- [configs/config.example.json](D:/code/GolandProjects/InfoHub/configs/config.example.json)
- [configs/config.example.env](D:/code/GolandProjects/InfoHub/configs/config.example.env)

如果你想在不修改示例配置的前提下做本地 RSS 验证，可以使用：

- `configs/config.local.json`

注意：

- `configs/config.local.json` 内含面向 Docker Compose 网络的 MySQL DSN：`mysql:3306`
- 这个文件更适合在 Docker Compose 内部使用
- 如果你在主机上运行并希望只使用文件存储，优先使用纯环境变量，或显式清空 `INFOHUB_MYSQL_DSN`

常用环境变量：

- `INFOHUB_CONFIG_PATH`
- `INFOHUB_RSS_URL`
- `INFOHUB_RSS_URLS`
- `INFOHUB_RSS_MAX_ITEMS_PER_FEED`
- `INFOHUB_RSS_RECENT_WITHIN_HOURS`
- `INFOHUB_REPORT_MAX_ITEMS`
- `INFOHUB_AI_ENDPOINT`
- `INFOHUB_AI_API_KEY`
- `INFOHUB_AI_MODEL`
- `INFOHUB_WEBHOOK_URL`
- `INFOHUB_SMTP_HOST`
- `INFOHUB_SMTP_PORT`
- `INFOHUB_SMTP_USERNAME`
- `INFOHUB_SMTP_PASSWORD`
- `INFOHUB_EMAIL_FROM`
- `INFOHUB_EMAIL_TO`
- `INFOHUB_EMAIL_SUBJECT`
- `INFOHUB_PREFERENCE_TAGS`
- `INFOHUB_PREFERENCE_SOURCES`
- `INFOHUB_PREFERENCE_KEYWORDS`
- `INFOHUB_PREFERENCE_TAG_WEIGHT`
- `INFOHUB_PREFERENCE_SOURCE_WEIGHT`
- `INFOHUB_PREFERENCE_KEYWORD_WEIGHT`
- `INFOHUB_SEND_EMPTY_REPORT`
- `INFOHUB_STORAGE_DIR`
- `INFOHUB_DEDUP_STORE_PATH`
- `INFOHUB_HTTP_ADDR`
- `INFOHUB_AUTH_TOKEN`
- `INFOHUB_REDIS_ADDR`
- `INFOHUB_REDIS_PASSWORD`
- `INFOHUB_REDIS_DB`
- `INFOHUB_REDIS_DEDUP_KEY`
- `INFOHUB_MYSQL_DSN`
- `INFOHUB_MYSQL_TABLE`
- `INFOHUB_MYSQL_PREFERENCE_TABLE`
- `INFOHUB_SCHEDULE_INTERVAL_SECONDS`
- `INFOHUB_SCHEDULE_CRON`

数据源配置也可以在 JSON 中显式使用 `sources` 数组：

```json
{
  "sources": [
    {
      "name": "google-rss",
      "kind": "rss",
      "location": "https://blog.google/rss/",
      "enabled": true,
      "timeout_seconds": 10
    },
    {
      "name": "custom-api",
      "kind": "http_json",
      "location": "https://example.com/news.json",
      "headers": {
        "Authorization": "Bearer <token>"
      }
    }
  ]
}
```

`enabled`、`timeout_seconds` 和 `headers` 都是可选的单源运行时设置。

`http_json` 目前接受两种输入格式：顶层数组，或带有 `items` 字段的对象。每个条目可以包含 `title`、`content`、`source`、`url`、`publish_time`、`tags` 和 `score`。

如果你希望按来源对生成的 Markdown 报告进行分组，请将 `report.group_by_source` 设为 `true`。

如果你希望决策评分优先考虑选定的标签、来源和关键词，可以设置个性化偏好。同时也可以通过偏好权重参数调节标签、来源和关键词匹配的相对影响。

## HTTP API

健康检查：

```http
GET /health
```

执行报告生成：

```http
POST /reports/run
```

可选请求体：

```json
{
  "user_id": "alice",
  "preference": {
    "tags": ["AI", "Agent"],
    "sources": ["openai-news"],
    "keywords": ["workflow"]
  }
}
```

当提供 `user_id` 时，服务会先加载该用户的已存储偏好，再叠加请求体中传入的偏好覆盖项。

管理已存储的用户偏好：

```http
PUT /preferences/:userID
GET /preferences/:userID
```

示例响应：

```json
{
  "status": "generated",
  "item_count": 26,
  "display_count": 12
}
```

读取最新报告：

```http
GET /reports/latest
```

最新报告响应中包含 `display_count`，表示实际渲染到 Markdown 报告中的条目数量。

列出历史报告：

```http
GET /reports
```

每条历史记录都包含 `item_count` 和 `display_count`，便于客户端在不打开全文的情况下展示报告规模。

如果配置了 `INFOHUB_AUTH_TOKEN`，除 `/health` 外的所有接口都需要：

```http
Authorization: Bearer <token>
```

## 报告存储

### 文件存储

默认输出路径：

```text
data/reports/reports/<timestamp>.md
data/reports/items/<timestamp>.json
data/dedup/seen.json
```

### MySQL 存储

设置 `INFOHUB_MYSQL_DSN` 以启用基于 MySQL 的报告存储。如果该值为空，应用会使用文件存储。

当你在主机上运行时，不要把 `INFOHUB_MYSQL_DSN` 指向 Docker 内部主机名 `mysql:3306`。应改用以下之一：

- 留空 DSN，使用文件存储
- 使用 `localhost:3307`，当 Compose 数据库暴露给主机时

默认表名：

```text
reports
```

MySQL 初始化 SQL 位于：

```text
scripts/mysql/init/001_create_reports.sql
```

该初始化文件现在同时创建报告表和偏好接口所使用的用户偏好表。

## Docker

构建镜像：

```bash
docker build -t infohub-agent:local .
```

启动完整本地栈：

```bash
docker compose up --build
```

默认情况下，Docker Compose 现在使用：

```text
/app/configs/config.local.json
```

这样本地栈默认会走 RSS + Mock AI，更适合做验证。

如果你希望显式覆盖配置：

```bash
INFOHUB_CONFIG_PATH=/app/configs/config.local.json docker compose up --build
```

PowerShell 示例：

```powershell
$env:INFOHUB_CONFIG_PATH="/app/configs/config.local.json"
docker compose up --build
```

如果你希望在启动时显式覆盖 Markdown 展示上限：

```bash
INFOHUB_REPORT_MAX_ITEMS=12 docker compose up --build
```

这会启动：

- `mysql:8.4`，映射到 `localhost:3307`
- `redis:7`，映射到 `localhost:6379`
- `infohub-agent`，映射到 `localhost:8080`

仅启动依赖：

```bash
docker compose up -d mysql redis
```

主机默认 DSN：

```text
infohub:infohub@tcp(localhost:3307)/infohub?charset=utf8mb4&parseTime=true&loc=Local
```

Docker Compose 内默认 DSN：

```text
infohub:infohub@tcp(mysql:3306)/infohub?charset=utf8mb4&parseTime=true&loc=Local
```

本地 RSS 配置当前使用以下订阅源：

- [Google Blog RSS](https://blog.google/rss/)
- [OpenAI News RSS](https://openai.com/news/rss.xml)

本地 RSS 配置也默认控制了抓取规模：

- 仅保留最近 `168` 小时内的内容
- 每个订阅源最多保留 `15` 条
- Markdown 报告只渲染前 `12` 条

由于默认去重存储是持久化的，重复调用 `POST /reports/run` 可能返回：

```json
{
  "status": "generated",
  "item_count": 0,
  "display_count": 0
}
```

这通常表示采集器没有发现新的未见条目，并不代表接口失败。

如果你想针对 Compose 栈做一次干净的验证运行，可以在容器内执行带临时去重路径的一次性命令：

```bash
docker exec \
  -e INFOHUB_CONFIG_PATH=/app/configs/config.local.json \
  -e INFOHUB_REPORT_MAX_ITEMS=12 \
  -e INFOHUB_DEDUP_STORE_PATH=/tmp/verify-seen.json \
  -e INFOHUB_MYSQL_DSN="infohub:infohub@tcp(mysql:3306)/infohub?charset=utf8mb4&parseTime=true&loc=Local" \
  -e INFOHUB_MYSQL_TABLE=reports \
  infohub-agent /app/infohub-agent run-once
```

这样可以在生成一份全新验证报告的同时，不影响正常运行时的去重状态。

## 当前限制

- Mock AI 模式使用启发式标签、评分和结构化摘要，适合本地验证，但不能替代真实 LLM 的输出质量。
- 真实 AI 摘要依赖兼容 OpenAI 的端点。
- Redis 去重尚未实现 TTL 或按日期分区。
- 公网部署仍建议置于反向代理之后。
- 报告概览中已经展示部分数据源失败，但还没有更丰富的重试或告警层。

## 建议的下一步

- 增加生产环境反向代理与 HTTPS 示例
- 增加更丰富的数据源健康监控与重试策略
- 结合生产样本继续校准真实 AI 提示词质量与评分
