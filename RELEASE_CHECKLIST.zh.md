# InfoHub 发布检查清单

[English](./RELEASE_CHECKLIST.md) | [简体中文](./RELEASE_CHECKLIST.zh.md)

## 发布目标

交付一套可用的信息汇总 Agent，能够：

- 拉取真实 RSS 源
- 去重并聚合近似重复条目
- 生成结构化的 Markdown 日报
- 将报告写入文件或 MySQL
- 通过命令行、定时任务或 HTTP API 运行
- 在报告概览中体现部分数据源失败的情况

## 已就绪能力

- 真实 RSS 采集、正文清洗与兜底抽取
- 单次运行内与跨次运行的去重
- 相似事件合并
- Mock AI 摘要、标签与启发式打分，便于本地验证
- 兼容 OpenAI 的 AI 客户端，用于真实摘要
- Markdown 渲染、分组渲染、Webhook 与邮件投递
- 文件存储与 MySQL 存储
- 基于 Redis 的去重存储
- 定时调度模式与 HTTP API
- 最终报告中来源均衡的 Top N 展示
- 报告概览中的部分数据源失败告警
- `go test ./...` 全量测试通过

## 最低可发布范围

若以下**全部**满足，即可认为达到发布标准：

- `go test ./...` 通过
- `go run cmd/main.go run-once` 在 Mock AI + 文件存储下可正常运行
- 使用真实 RSS 做一次验证运行，能产出可读的 Markdown 报告
- 报告概览展示来源分布，且在部分拉取失败时出现告警行
- 多个源均成功时，报告靠前部分不应被单一来源完全占据
- 本地 `POST /reports/run` 可用
- 最新报告与历史报告相关 API 能返回已存储的报告元数据

## 宿主机侧验证清单

在宿主机上做**最快**发布检查时，可按此路径执行：

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

检查项：

- `data/reports-verify/reports` 下出现新的 Markdown 文件
- `data/reports-verify/items` 下出现新的 JSON 条目文件
- Markdown 概览可读
- 靠前条目排序/优先级看起来合理
- 某一源失败而其他源成功时，出现相应告警

## Docker 验证清单

若需将 MySQL、Redis 一并纳入验证，使用此路径：

```bash
docker compose up --build
```

随后确认：

- 服务在 `localhost:8080` 启动
- MySQL 可在 `localhost:3307` 访问
- Redis 可在 `localhost:6379` 访问
- `POST /reports/run` 返回 `generated`
- `GET /reports/latest` 返回 Markdown 及摘要相关元数据

## 真实 AI 验证清单

配置以下环境变量：

- `INFOHUB_AI_ENDPOINT`
- `INFOHUB_AI_API_KEY`
- `INFOHUB_AI_MODEL`

随后确认：

- 摘要格式仍符合要求的标签化结构
- 标签合理
- 分数没有明显虚高
- 在相同输入下，输出质量优于 Mock 模式

## 已知非阻塞限制

以下项**不必**阻碍首次发布：

- Mock AI 质量为启发式，仅用于本地验证
- Redis 去重尚未配置 TTL 或分片策略
- 数据源失败已在报告中可见，但重试与告警仍较基础
- 公网部署仍建议在前面加反向代理

## 发布阻塞项

若存在以下**任一**情况，**不得**宣告发布完成：

- 测试失败
- 真实 RSS 验证无法产出报告
- 报告内容大多为重复条目
- 某一数据源静默失败且报告中无任何可见告警
- 宿主机侧操作说明必须依赖「仅 Docker 可用」的配置才能跑通
- 报告靠前部分明显被低价值推广类内容占据

## 建议发布顺序

1. 执行 `go test ./...`
2. 执行宿主机侧 Mock AI 验证
3. 人工检查生成的 Markdown
4. 执行 Docker Compose 验证
5. 验证 HTTP API 端点
6. 若发布环境使用真实 AI，再跑一轮真实 AI 抽样检查
7. 冻结默认配置并对外发布

## 发布后建议的近期跟进

- 增加数据源健康监控与重试
- 结合生产样本收紧真实 AI 的打分与摘要校准
- 补充反向代理与 HTTPS 的部署示例
