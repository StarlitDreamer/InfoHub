# AGENTS.md - InfoHub Agent Project Rules

[简体中文](./AGENTS.md) | [English](./AGENTS.en.md)

## 1. Project overview

This project is an "information aggregation and decision-support agent." It automatically collects information from multiple sources, deduplicates and classifies it, generates summaries and scores, and outputs a high-quality daily brief that supports decision-making.

Core goals:

- Reduce the time users spend filtering information
- Provide decision value that directly supports judgment and action
- Support automatic delivery to email, webhook, or bots

## 2. Tech stack

- Language: Go >= 1.22
- Web framework: Gin
- Primary storage: MySQL
- Cache and deduplication: Redis
- AI capability: OpenAI, Claude, or other LLM APIs
- Search capability: Elasticsearch, optional
- Scheduling: cron or a custom scheduler

## 3. Recommended project structure

```text
project-root/
├── cmd/
│   └── main.go
├── internal/
│   ├── crawler/
│   ├── processor/
│   ├── ai/
│   ├── service/
│   ├── repository/
│   ├── scheduler/
│   └── delivery/
├── pkg/
├── configs/
├── scripts/
├── tests/
└── AGENTS.md
```

## 4. Core data structure

```go
type NewsItem struct {
    ID          int64
    Title       string
    Content     string
    Source      string
    URL         string
    PublishTime time.Time
    Tags        []string
    Score       float64
}
```

## 5. Core module responsibilities

### crawler

- Collect information from multiple data sources
- Must guarantee idempotent collection
- Must support plugin-style source extension

### processor

- Deduplicate by title, URL, content fingerprint, or embedding
- Clean article body content
- Aggregate similar content

### ai

Must provide the following capabilities:

- Classification
- Structured summaries
- Importance scoring

AI output format:

```text
[Title]
[What happened]
[Why it matters]
[Impact]
[Score] 1-5
```

### service

Responsible for orchestrating the core business flow:

```text
collect -> deduplicate -> AI processing -> storage -> output
```

### scheduler

- Support cron-based scheduled triggers
- Support manual triggers

### delivery

Must support at least:

- Markdown daily reports
- Email delivery
- Webhook or bot delivery

## 6. Agent execution flow

When the agent processes information, it must follow this order:

1. Get the source list
2. Run collection
3. Run deduplication
4. Call AI processing
5. Store the results
6. Sort by score
7. Output the daily report or push result

## 7. Code standards

- Follow idiomatic Go style
- Keep functions short, ideally under 50 lines
- Keep module responsibilities single-purpose
- Avoid giant functions
- Business capabilities must have interface abstractions
- Code must include comments, and comments must use Simplified Chinese
- New features must include tests
- Do not introduce undeclared dependencies
- Do not modify unrelated modules

## 7.1 Documentation standards

- Markdown documents in this repository must always be provided in both Chinese and English
- Keep the default entry filename and add the paired language version with a suffix, for example `README.md` with `README.zh.md`, and `AGENTS.md` with `AGENTS.en.md`
- Whenever any `.md` file is added or updated, the corresponding language version must be updated as well so the two versions do not drift apart

## 8. Codex working rules

When Codex executes a task, it must:

1. Read the project structure first
2. Identify the target module clearly
3. Output a 3-6 step execution plan
4. List the expected affected files
5. Start coding only after that
6. Add necessary comments while coding, and comments must use Simplified Chinese
7. Make small, incremental changes and avoid broad edits in one pass
8. Run related tests after completion, or explain why tests were not run
9. Prioritize the agent's core delivery path when core and peripheral work conflict

Forbidden behavior:

- Skipping analysis and starting to code immediately
- Coding without a plan
- Modifying many files at once
- Modifying modules unrelated to the task
- Introducing undeclared dependencies

## 9. Core-path-first principle

Future development in this project must default to advancing the agent's core path first. Unless the user explicitly requests otherwise, or the core path is already complete enough for stable delivery, peripheral enhancements should not be prioritized ahead of core capabilities.

### Core agent capability scope

The following are core-path capabilities and have the highest priority:

1. Data source collection quality
2. Body cleaning and content extraction
3. Deduplication accuracy and similar-content aggregation
4. AI classification, structured summaries, and importance scoring
5. Stability of core workflow orchestration
6. Daily brief quality and decision usefulness
7. Core-path test coverage and regression verification

### Non-core capability scope

The following are considered peripheral enhancements and are lower priority by default:

- Expanding push channels
- Deeper personalization
- Multi-user support
- Preference storage
- Admin panels or auxiliary APIs
- Other productization work that does not directly improve the core pipeline quality

### Execution requirement

When the user request does not specify a direction clearly, Codex should first judge whether the work directly improves this pipeline:

```text
collect -> clean -> deduplicate -> AI processing -> ranking -> storage -> daily report output
```

If the task does not directly improve that pipeline, Codex should first explain that it is peripheral work and then propose core-path gap fixes before prioritizing it.

## 10. Development priorities

### Phase 1: MVP

- Data collection
- AI summaries
- Markdown output

### Phase 2

- Deduplication
- Score-based ranking

### Phase 3

- Push system
- Personalization

## 11. Future expansion

- User-interest recommendation
- GitHub project analysis
- Automatic report generation
- Automatic actions such as starring or saving projects

## 12. Success criteria

- Generate information summaries automatically every day
- No duplicate content
- Structured output
- User reading time under 5 minutes

## 13. Common commands

```bash
# Start the service
go run cmd/main.go

# Run tests
go test ./...

# Build
go build -o app cmd/main.go

# Format
go fmt ./...
```
