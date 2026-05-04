# InfoHub Agent

InfoHub Agent is an information aggregation and decision-support service. It collects content from RSS and demo sources, deduplicates items, runs AI summarization and scoring, and outputs a decision-ready daily report.

## Current capabilities

- Demo crawler and multi-RSS aggregation
- Mock AI processor and OpenAI-compatible Chat Completions client
- In-run and cross-run deduplication
- Markdown report rendering
- File-backed or MySQL-backed report storage
- Webhook and email delivery
- Manual run, scheduler mode, and Gin HTTP API
- JSON config plus environment variable overrides
- Redis-backed dedup store

## Project layout

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

## Quick start

Run tests:

```bash
go test ./...
```

Run the MySQL integration test against the local Compose database:

```bash
INFOHUB_TEST_MYSQL_DSN="infohub:infohub@tcp(localhost:3307)/infohub?charset=utf8mb4&parseTime=true&loc=Local" go test ./internal/repository -run TestMySQLReportRepositoryIntegration
```

Run the Redis integration test against the local Compose cache:

```bash
INFOHUB_TEST_REDIS_ADDR="localhost:6379" go test ./internal/processor -run TestRedisDedupStoreIntegration
```

Generate one report:

```bash
go run cmd/main.go run-once
```

Start scheduler mode:

```bash
go run cmd/main.go schedule
```

The scheduler supports either a fixed interval via `INFOHUB_SCHEDULE_INTERVAL_SECONDS` or a 5-field cron expression via `INFOHUB_SCHEDULE_CRON`, for example `0 9,18 * * 1-5`.

Start HTTP server:

```bash
go run cmd/main.go serve
```

If no mode is passed, the app defaults to `run-once`.

## Configuration

Use the bundled JSON example:

```bash
INFOHUB_CONFIG_PATH=configs/config.example.json go run cmd/main.go run-once
```

PowerShell example:

```powershell
$env:INFOHUB_CONFIG_PATH="D:\code\GolandProjects\InfoHub\configs\config.example.json"
go run cmd\main.go run-once
```

The example files are:

- [configs/config.example.json](D:/code/GolandProjects/InfoHub/configs/config.example.json)
- [configs/config.example.env](D:/code/GolandProjects/InfoHub/configs/config.example.env)

For local RSS verification without touching the example config, use:

- `configs/config.local.json`

Common environment variables:

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
- `INFOHUB_SCHEDULE_INTERVAL_SECONDS`
- `INFOHUB_SCHEDULE_CRON`

Source config can also use an explicit `sources` array in JSON:

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

`enabled`, `timeout_seconds`, and `headers` are optional per-source runtime settings.

`http_json` currently accepts either a top-level array or an object with an `items` field. Each item may include `title`, `content`, `source`, `url`, `publish_time`, `tags`, and `score`.

Set `report.group_by_source` to `true` if you want the generated Markdown report to be grouped by item source.

Set personalization preferences if you want decision scoring to favor selected tags, sources, and keywords. You can also tune the relative impact of tag, source, and keyword matches through preference weight settings.

## HTTP API

Health check:

```http
GET /health
```

Run report generation:

```http
POST /reports/run
```

Optional request body:

```json
{
  "preference": {
    "tags": ["AI", "Agent"],
    "sources": ["openai-news"],
    "keywords": ["workflow"]
  }
}
```

Example response:

```json
{
  "status": "generated",
  "item_count": 26,
  "display_count": 12
}
```

Read latest report:

```http
GET /reports/latest
```

The latest report response includes `display_count`, which reflects how many items are actually rendered in the Markdown report.

List historical reports:

```http
GET /reports
```

Each history entry includes `item_count` and `display_count` so clients can show report size without opening the full report.

If `INFOHUB_AUTH_TOKEN` is configured, all endpoints except `/health` require:

```http
Authorization: Bearer <token>
```

## Report storage

### File storage

Default output paths:

```text
data/reports/reports/<timestamp>.md
data/reports/items/<timestamp>.json
data/dedup/seen.json
```

### MySQL storage

Set `INFOHUB_MYSQL_DSN` to enable MySQL-backed report storage. If it is empty, the app uses file storage.

Default table name:

```text
reports
```

The MySQL initialization SQL is stored at:

```text
scripts/mysql/init/001_create_reports.sql
```

## Docker

Build the image:

```bash
docker build -t infohub-agent:local .
```

Run the full local stack:

```bash
docker compose up --build
```

By default, Docker Compose now uses:

```text
/app/configs/config.local.json
```

This keeps the local stack on RSS + mock AI by default, which is safer for verification.

If you want to override the config explicitly:

```bash
INFOHUB_CONFIG_PATH=/app/configs/config.local.json docker compose up --build
```

PowerShell example:

```powershell
$env:INFOHUB_CONFIG_PATH="/app/configs/config.local.json"
docker compose up --build
```

If you want to override the Markdown display cap explicitly at startup:

```bash
INFOHUB_REPORT_MAX_ITEMS=12 docker compose up --build
```

This starts:

- `mysql:8.4` on `localhost:3307`
- `redis:7` on `localhost:6379`
- `infohub-agent` on `localhost:8080`

Start only dependencies:

```bash
docker compose up -d mysql redis
```

Default DSN on the host machine:

```text
infohub:infohub@tcp(localhost:3307)/infohub?charset=utf8mb4&parseTime=true&loc=Local
```

Default DSN inside Docker Compose:

```text
infohub:infohub@tcp(mysql:3306)/infohub?charset=utf8mb4&parseTime=true&loc=Local
```

The local RSS config currently uses these feeds:

- [Google Blog RSS](https://blog.google/rss/)
- [OpenAI News RSS](https://openai.com/news/rss.xml)

The local RSS config also trims feed volume by default:

- keep only items from the last `168` hours
- keep at most `15` items per feed
- render only the top `12` items in the Markdown report

Because the default dedup store is persistent, repeated `POST /reports/run` calls may return:

```json
{
  "status": "generated",
  "item_count": 0,
  "display_count": 0
}
```

That usually means the crawler did not find any new unseen items, not that the endpoint failed.

For a clean verification run against the Compose stack, execute a one-off command with a temporary dedup store path inside the container:

```bash
docker exec \
  -e INFOHUB_CONFIG_PATH=/app/configs/config.local.json \
  -e INFOHUB_REPORT_MAX_ITEMS=12 \
  -e INFOHUB_DEDUP_STORE_PATH=/tmp/verify-seen.json \
  -e INFOHUB_MYSQL_DSN="infohub:infohub@tcp(mysql:3306)/infohub?charset=utf8mb4&parseTime=true&loc=Local" \
  -e INFOHUB_MYSQL_TABLE=reports \
  infohub-agent /app/infohub-agent run-once
```

This keeps the normal runtime dedup state untouched while generating a fresh report for verification.

## Current limitations

- RSS parsing focuses on common fields and does not yet do deeper HTML extraction.
- Real AI summaries depend on an OpenAI-compatible endpoint.
- Redis dedup does not yet implement TTL or date partitioning.
- Public deployment should still sit behind a reverse proxy.

## Suggested next steps

- Improve RSS HTML cleaning and content extraction
- Add production reverse proxy and HTTPS examples
- Add MySQL integration tests against a real container
