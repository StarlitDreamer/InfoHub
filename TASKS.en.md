# TASKS.md - InfoHub Agent Development Checklist

[简体中文](./TASKS.md) | [English](./TASKS.en.md)

## 0. Global execution rules

### Required execution flow

1. Read `AGENTS.md`
2. Clarify the scope of the current task
3. Output a 3-6 step execution plan
4. Start implementation only after that

### General constraints

- Complete only one task module at a time
- Do not modify code across unrelated modules
- All code must compile
- Tests must be run when they exist
- Keep changes minimal

## 1. Project bootstrap

### Goal

Initialize a Go project skeleton that matches the structure defined in `AGENTS.md`.

### Execution steps

1. Create the project directory structure:

```bash
mkdir -p cmd internal/{crawler,processor,ai,service,repository,scheduler,delivery} pkg configs tests
```

2. Initialize the Go module:

```bash
go mod init InfoHub-agent
```

3. Create the basic entry file:

- `cmd/main.go`

### Verification criteria

- `go run cmd/main.go` runs successfully
- The project structure matches `AGENTS.md`

## 2. Define the core data model

### Goal

Implement the core `NewsItem` data structure.

### Execution steps

1. Create the file:

```text
internal/model/news.go
```

2. Define the structure fields:

- ID
- Title
- Content
- Source
- URL
- PublishTime
- Tags
- Score

### Verification criteria

- Code compiles
- The structure can be imported by other modules

## 3. Implement the crawler module

### Goal

Implement a basic data collector, using mock or simple data sources first.

### Execution steps

1. Define the interface:

```go
type Crawler interface {
    Fetch() ([]NewsItem, error)
}
```

2. Implement a demo crawler:

- Return mock data
- Do not connect to a real API yet

### Verification criteria

- Can return at least 3 items
- The service layer can call it

## 4. Implement the processor module

### Goal

Implement basic deduplication logic.

### Execution steps

1. Create:

```text
internal/processor/deduplicate.go
```

2. Implement:

- Title-based deduplication using a map
- Keep the first item

### Verification criteria

- Duplicate input becomes unique output
- Unit tests pass

## 5. Implement the AI module

### Goal

Wrap the AI interface, starting with a mock implementation.

### Execution steps

1. Define the interface:

```go
type AIProcessor interface {
    Summarize(item NewsItem) (NewsItem, error)
}
```

2. Implement a mock version:

- Add a generated summary based on `Content`
- Generate a score from 1 to 5

### Verification criteria

- Every item has a `Score`
- Output structure matches `AGENTS.md`

## 6. Implement the business workflow

### Goal

Connect the whole pipeline together.

### Execution steps

Create:

```text
internal/service/pipeline.go
```

Implement the workflow:

```text
Fetch -> Deduplicate -> AI -> Return
```

### Verification criteria

- Returns a fully processed item list
- No panic

## 7. Implement the delivery module

### Goal

Output a Markdown report.

### Execution steps

1. Create:

```text
internal/delivery/markdown.go
```

2. Implement the output format:

```md
# Today's Information

## [Category]
- Title
- Summary
```

### Verification criteria

- Can generate a Markdown string
- Content structure is correct

## 8. Integrate the main flow

### Goal

Connect the full end-to-end path.

### Execution steps

1. In `main.go`:

- Initialize crawler
- Initialize processor
- Initialize AI
- Call the pipeline
- Output Markdown

### Verification criteria

```bash
go run cmd/main.go
```

Output:

- One complete daily report

## 9. Add scheduled tasks

### Goal

Implement automatic execution.

### Execution steps

1. Use cron:

```go
cron.New().AddFunc("@every 1h", job)
```

2. Call the pipeline

### Verification criteria

- Runs once per hour
- No repeated crashes

## 10. Introduce real data sources

### Goal

Connect to real data such as RSS or GitHub.

### Execution steps

- Replace the mock crawler
- Support at least one real source

### Verification criteria

- The source data is real
- Parsing works normally

## 11. Connect a real AI model

### Goal

Connect to a real LLM.

### Execution steps

- Wrap an API client
- Implement a prompt such as:

```text
Summarize the following content using this format:
[What happened]
[Why it matters]
[Impact]
[Score]
```

### Verification criteria

- Produces real summaries
- Output format is correct

## 12. Add delivery channels

### Goal

Support notifications.

### Execution steps

Implement at least one of:

- Email
- Webhook
- Bot

### Verification criteria

- Messages are delivered successfully
- Content is correct

## 13. Upgrade deduplication

### Goal

Improve deduplication accuracy.

### Execution steps

- Introduce embeddings
- Calculate similarity

### Verification criteria

- Similar content is merged
- Accuracy improves

## 14. Ranking strategy

### Goal

Implement intelligent ranking.

### Execution steps

Example scoring model:

```text
score = popularity + recency + AI score
```

### Verification criteria

- Higher-value content appears first

## 15. Personalization

### Goal

Support user preferences.

### Execution steps

- Tag filtering
- Subscription mechanism

### Verification criteria

- Users only see content they care about
