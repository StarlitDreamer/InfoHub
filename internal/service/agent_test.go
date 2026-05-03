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
	if strings.Count(repo.record.Markdown, "## ") != 2 {
		t.Fatalf("expected markdown to contain 2 sections, got %s", repo.record.Markdown)
	}
	if strings.Contains(repo.record.Markdown, "alpha") {
		t.Fatalf("expected trimmed markdown to exclude alpha, got %s", repo.record.Markdown)
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
