package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/repository"
)

func TestAgentRunStoresSortedItemsAndTrimmedMarkdown(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 20, 0, 0, 0, time.UTC)
	repo := &agentRepositoryStub{}
	agent := NewAgent(
		staticPipelineRunner{
			items: []model.NewsItem{
				{Title: "alpha", Content: "body a", Score: 1, PublishTime: fixedNow.Add(-2 * time.Hour)},
				{Title: "beta", Content: "body b", Score: 4, PublishTime: fixedNow.Add(-1 * time.Hour)},
				{Title: "gamma", Content: "body c", Score: 3, PublishTime: fixedNow.Add(-30 * time.Minute)},
			},
		},
		repo,
		AgentOptions{
			ReportMaxItems: 2,
			Now:            func() time.Time { return fixedNow },
		},
	)

	result, err := agent.Run(context.Background())
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	if result.ItemCount != 3 || result.DisplayCount != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("expected one save call, got %d", repo.saveCalls)
	}
	if repo.record.GeneratedAt != fixedNow {
		t.Fatalf("expected generated time %s, got %s", fixedNow, repo.record.GeneratedAt)
	}
	if len(repo.record.Items) != 3 {
		t.Fatalf("expected 3 stored items, got %d", len(repo.record.Items))
	}
	if repo.record.Items[0].Title != "beta" || repo.record.Items[1].Title != "gamma" {
		t.Fatalf("unexpected sorted order: %+v", repo.record.Items)
	}
	if strings.Count(repo.record.Markdown, "## ⭐") != 2 {
		t.Fatalf("expected markdown to contain 2 sections, got %s", repo.record.Markdown)
	}
	if strings.Contains(repo.record.Markdown, "alpha") {
		t.Fatalf("expected trimmed markdown to exclude alpha, got %s", repo.record.Markdown)
	}
}

func TestAgentRunWithRequestCarriesExecutionContext(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 20, 0, 0, 0, time.UTC)
	agent := NewAgent(
		staticPipelineRunner{
			items: []model.NewsItem{
				{Title: "beta", Content: "body b", Score: 4, PublishTime: fixedNow.Add(-1 * time.Hour)},
			},
		},
		&agentRepositoryStub{},
		AgentOptions{
			Now: func() time.Time { return fixedNow },
		},
	)

	result, err := agent.RunWithRequest(context.Background(), AgentRequest{
		Context: ExecutionContext{
			Trigger: "http",
			Sources: []Source{
				{Name: "feed-a", Kind: "rss", Location: "https://example.com/a.xml"},
			},
		},
	})
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	if result.Request.Context.Trigger != "http" {
		t.Fatalf("expected trigger http, got %s", result.Request.Context.Trigger)
	}
	if result.Request.Context.RequestedAt != fixedNow {
		t.Fatalf("expected requested time %s, got %s", fixedNow, result.Request.Context.RequestedAt)
	}
	if len(result.Request.Context.Sources) != 1 || result.Request.Context.Sources[0].Kind != "rss" {
		t.Fatalf("unexpected sources: %+v", result.Request.Context.Sources)
	}
}

func TestAgentRunSendsWebhookWhenConfigured(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 20, 0, 0, 0, time.UTC)
	sender := &senderStub{}
	agent := NewAgent(
		staticPipelineRunner{
			items: []model.NewsItem{
				{Title: "beta", Content: "body b", Score: 4, PublishTime: fixedNow.Add(-1 * time.Hour)},
			},
		},
		&agentRepositoryStub{},
		AgentOptions{
			WebhookSender: sender,
			Now:           func() time.Time { return fixedNow },
		},
	)

	if _, err := agent.Run(context.Background()); err != nil {
		t.Fatalf("agent run failed: %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("expected webhook sender to be called once, got %d", sender.calls)
	}
}

func TestAgentRunCanGroupMarkdownBySource(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 20, 0, 0, 0, time.UTC)
	repo := &agentRepositoryStub{}
	agent := NewAgent(
		staticPipelineRunner{
			items: []model.NewsItem{
				{Title: "a", Content: "summary a", Source: "OpenAI News", Score: 5, PublishTime: fixedNow.Add(-1 * time.Hour)},
				{Title: "b", Content: "summary b", Source: "Google Blog", Score: 4, PublishTime: fixedNow.Add(-2 * time.Hour)},
			},
		},
		repo,
		AgentOptions{
			GroupBySource: true,
			Now:           func() time.Time { return fixedNow },
		},
	)

	result, err := agent.Run(context.Background())
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	if !strings.Contains(result.Markdown, "### 来源：Google Blog") || !strings.Contains(result.Markdown, "### 来源：OpenAI News") {
		t.Fatalf("expected grouped markdown, got %s", result.Markdown)
	}
}

func TestAgentRunAppliesSourcePriorityAndReportFilter(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 20, 0, 0, 0, time.UTC)
	repo := &agentRepositoryStub{}
	agent := NewAgent(
		staticPipelineRunner{
			items: []model.NewsItem{
				{SourceName: "normal", Title: "normal item", Content: "summary a", Source: "OpenAI News", Score: 5, PublishTime: fixedNow.Add(-1 * time.Hour)},
				{SourceName: "priority", Title: "priority item", Content: "summary b", Source: "Google Blog", Score: 1, PublishTime: fixedNow.Add(-2 * time.Hour)},
				{SourceName: "hidden", Title: "hidden item", Content: "summary c", Source: "Internal Feed", Score: 5, PublishTime: fixedNow.Add(-30 * time.Minute)},
			},
		},
		repo,
		AgentOptions{
			Now: func() time.Time { return fixedNow },
		},
	)

	result, err := agent.RunWithRequest(context.Background(), AgentRequest{
		Context: ExecutionContext{
			Sources: []Source{
				{Name: "normal", IncludeInReport: true},
				{Name: "priority", Priority: 10, IncludeInReport: true},
				{Name: "hidden", IncludeInReport: false},
			},
		},
	})
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	if len(result.Items) != 3 || result.Items[0].SourceName != "priority" {
		t.Fatalf("expected priority source to sort first, got %+v", result.Items)
	}
	if strings.Contains(result.Markdown, "hidden item") {
		t.Fatalf("expected hidden source to be excluded from report, got %s", result.Markdown)
	}
	if result.DisplayCount != 2 {
		t.Fatalf("expected display count 2 after source filter, got %d", result.DisplayCount)
	}
}

func TestAgentRunSkipsEmptyWebhookByDefault(t *testing.T) {
	agent := NewAgent(
		staticPipelineRunner{},
		&agentRepositoryStub{},
		AgentOptions{
			WebhookSender: &senderStub{},
		},
	)

	result, err := agent.Run(context.Background())
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}
	if result.ItemCount != 0 || result.DisplayCount != 0 {
		t.Fatalf("unexpected empty result: %+v", result)
	}
}

func TestAgentRunPropagatesPipelineError(t *testing.T) {
	expectedErr := errors.New("pipeline failed")
	agent := NewAgent(
		errorPipelineRunner{err: expectedErr},
		&agentRepositoryStub{},
		AgentOptions{},
	)

	_, err := agent.Run(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected pipeline error, got %v", err)
	}
}

type staticPipelineRunner struct {
	items []model.NewsItem
}

func (r staticPipelineRunner) RunContext(ctx context.Context) ([]model.NewsItem, error) {
	return r.items, nil
}

type errorPipelineRunner struct {
	err error
}

func (r errorPipelineRunner) RunContext(ctx context.Context) ([]model.NewsItem, error) {
	return nil, r.err
}

type agentRepositoryStub struct {
	saveCalls int
	record    repository.ReportRecord
}

func (r *agentRepositoryStub) Save(ctx context.Context, record repository.ReportRecord) error {
	r.saveCalls++
	r.record = record
	return nil
}

func (r *agentRepositoryStub) Latest(ctx context.Context) (repository.ReportRecord, error) {
	return repository.ReportRecord{}, repository.ErrReportNotFound
}

func (r *agentRepositoryStub) List(ctx context.Context) ([]repository.ReportMetadata, error) {
	return nil, nil
}

type senderStub struct {
	calls int
}

func (s *senderStub) Send(markdown string) error {
	s.calls++
	return nil
}
