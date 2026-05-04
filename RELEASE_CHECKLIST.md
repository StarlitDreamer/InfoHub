# InfoHub Release Checklist

## Release target

Ship a usable information aggregation agent that can:

- fetch real RSS sources
- deduplicate and aggregate near-duplicate items
- generate structured Markdown daily reports
- store reports to files or MySQL
- run by command, scheduler, or HTTP API
- expose partial source failures in the report overview

## What is already ready

- Real RSS crawling with content cleaning and fallback extraction
- In-run and cross-run deduplication
- Similar event merging
- Mock AI summaries, tags, and heuristic scoring for local verification
- OpenAI-compatible AI client for real summarization
- Markdown rendering, grouped rendering, webhook, and email delivery
- File storage and MySQL storage
- Redis-backed dedup store
- Scheduler mode and HTTP API
- Source-balanced Top N display in the final report
- Partial source failure warnings in report overview
- Full test suite passing with `go test ./...`

## Minimum release scope

The release is acceptable if all items below are true:

- `go test ./...` passes
- `go run cmd/main.go run-once` works with mock AI and file storage
- A real RSS verification run produces a readable Markdown report
- The report overview shows source distribution and warning lines when partial fetch failures happen
- The top section of the report is not dominated by a single source when multiple sources succeed
- `POST /reports/run` works locally
- Latest and history report APIs return stored report metadata

## Host-side verification checklist

Use this path when you want the fastest release check on the host machine:

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

Check:

- a new Markdown file appears under `data/reports-verify/reports`
- a new JSON item file appears under `data/reports-verify/items`
- the Markdown overview is readable
- the top items look plausibly ranked
- warnings appear when one source fails and another succeeds

## Docker verification checklist

Use this path when you want MySQL and Redis in the loop:

```bash
docker compose up --build
```

Then verify:

- service starts on `localhost:8080`
- MySQL is reachable on `localhost:3307`
- Redis is reachable on `localhost:6379`
- `POST /reports/run` returns `generated`
- `GET /reports/latest` returns markdown and summary metadata

## Real AI verification checklist

Set these variables:

- `INFOHUB_AI_ENDPOINT`
- `INFOHUB_AI_API_KEY`
- `INFOHUB_AI_MODEL`

Then verify:

- summary format still matches the required labeled structure
- tags are reasonable
- scores are not obviously inflated
- output quality is better than mock mode on the same inputs

## Known non-blocking limitations

These do not have to block a first release:

- Mock AI quality is heuristic and is only for local verification
- Redis dedup has no TTL or partitioning strategy yet
- Source failure handling is visible in reports, but retry and alerting are still basic
- Public deployment still expects a reverse proxy in front

## Release blockers

Do not call the release done if any of these are true:

- tests are failing
- real RSS verification cannot produce a report
- the report content is mostly duplicate items
- a single source silently fails and there is no visible warning
- host-side instructions require Docker-only settings to work
- the top section is obviously dominated by low-value promotional items

## Recommended release order

1. Run `go test ./...`
2. Run host-side mock AI verification
3. Inspect generated Markdown manually
4. Run Docker Compose verification
5. Verify HTTP API endpoints
6. If using real AI in release, run one real AI sample check
7. Freeze config defaults and publish

## Suggested immediate follow-up after release

- Add source health monitoring and retries
- Tighten real AI scoring and summary calibration with production samples
- Add deployment examples with reverse proxy and HTTPS
