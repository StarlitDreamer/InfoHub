package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/config"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/repository"
)

func TestRunRejectsUnknownMode(t *testing.T) {
	err := run(context.Background(), config.Config{}, []string{"unknown"})

	if err == nil {
		t.Fatal("expected unknown mode to return an error")
	}
}

func TestRunOnceMode(t *testing.T) {
	cfg := config.Config{
		ScheduleInterval: time.Hour,
		StorageDir:       t.TempDir(),
		DedupStorePath:   t.TempDir() + "/seen.json",
	}

	if err := run(context.Background(), cfg, []string{"run-once"}); err != nil {
		t.Fatalf("run-once failed: %v", err)
	}
}

func TestRunReportReturnsItemAndDisplayCount(t *testing.T) {
	cfg := config.Config{
		ScheduleInterval: time.Hour,
		StorageDir:       t.TempDir(),
		DedupStorePath:   t.TempDir() + "/seen.json",
		ReportMaxItems:   2,
	}

	result, err := runReport(context.Background(), cfg, "manual")
	if err != nil {
		t.Fatalf("runReport failed: %v", err)
	}

	if result.ItemCount != 3 {
		t.Fatalf("expected 3 items, got %d", result.ItemCount)
	}
	if result.DisplayCount != 2 {
		t.Fatalf("expected 2 displayed items, got %d", result.DisplayCount)
	}

	result, err = runReport(context.Background(), cfg, "manual")
	if err != nil {
		t.Fatalf("second runReport failed: %v", err)
	}

	if result.ItemCount != 0 {
		t.Fatalf("expected 0 items on second run, got %d", result.ItemCount)
	}
	if result.DisplayCount != 0 {
		t.Fatalf("expected 0 displayed items on second run, got %d", result.DisplayCount)
	}
}

func TestRunReportWithRepositorySavesSortedItemsAndTrimmedMarkdown(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 18, 0, 0, 0, time.UTC)
	previousNow := timeNow
	timeNow = func() time.Time { return fixedNow }
	defer func() { timeNow = previousNow }()

	cfg := config.Config{
		DedupStorePath: t.TempDir() + "/seen.json",
		ReportMaxItems: 2,
	}
	repo := &captureReportRepository{}

	result, err := runReportWithRepository(context.Background(), cfg, repo, "manual")
	if err != nil {
		t.Fatalf("runReportWithRepository failed: %v", err)
	}

	if result.ItemCount != 3 {
		t.Fatalf("expected 3 sorted items, got %d", result.ItemCount)
	}
	if result.DisplayCount != 2 {
		t.Fatalf("expected 2 displayed items, got %d", result.DisplayCount)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("expected repository Save to be called once, got %d", repo.saveCalls)
	}
	if repo.record.GeneratedAt != fixedNow {
		t.Fatalf("expected saved GeneratedAt %s, got %s", fixedNow, repo.record.GeneratedAt)
	}
	if len(repo.record.Items) != 3 {
		t.Fatalf("expected repository to store all sorted items, got %d", len(repo.record.Items))
	}

	expectedOrder := []string{
		"开源模型发布新版本",
		"数据库社区讨论新索引策略",
		"云厂商推出开发者工具更新",
	}
	for index, title := range expectedOrder {
		if repo.record.Items[index].Title != title {
			t.Fatalf("expected sorted item %d to be %s, got %s", index, title, repo.record.Items[index].Title)
		}
	}

	if strings.Count(repo.record.Markdown, "## ") != 2 {
		t.Fatalf("expected markdown to render 2 sections, got %s", repo.record.Markdown)
	}
	if !strings.Contains(repo.record.Markdown, expectedOrder[0]) || !strings.Contains(repo.record.Markdown, expectedOrder[1]) {
		t.Fatalf("expected markdown to contain top 2 items, got %s", repo.record.Markdown)
	}
	if strings.Contains(repo.record.Markdown, expectedOrder[2]) {
		t.Fatalf("expected markdown to exclude trimmed item, got %s", repo.record.Markdown)
	}
}

func TestRunReportWithRepositoryStoresEmptyReportWhenNoNewItems(t *testing.T) {
	fixedNow := time.Date(2026, 5, 3, 18, 30, 0, 0, time.UTC)
	previousNow := timeNow
	timeNow = func() time.Time { return fixedNow }
	defer func() { timeNow = previousNow }()

	cfg := config.Config{
		DedupStorePath: t.TempDir() + "/seen.json",
		ReportMaxItems: 2,
	}

	firstRepo := &captureReportRepository{}
	if _, err := runReportWithRepository(context.Background(), cfg, firstRepo, "manual"); err != nil {
		t.Fatalf("first runReportWithRepository failed: %v", err)
	}

	secondRepo := &captureReportRepository{}
	result, err := runReportWithRepository(context.Background(), cfg, secondRepo, "manual")
	if err != nil {
		t.Fatalf("second runReportWithRepository failed: %v", err)
	}

	if result.ItemCount != 0 || result.DisplayCount != 0 {
		t.Fatalf("expected second run to generate empty result, got %+v", result)
	}
	if secondRepo.saveCalls != 1 {
		t.Fatalf("expected repository Save on empty report, got %d calls", secondRepo.saveCalls)
	}
	if len(secondRepo.record.Items) != 0 {
		t.Fatalf("expected empty stored items, got %+v", secondRepo.record.Items)
	}
	if !strings.Contains(secondRepo.record.Markdown, "今日暂无新增信息") {
		t.Fatalf("expected empty markdown message, got %s", secondRepo.record.Markdown)
	}
}

func TestBuildAgentRequestUsesRSSSources(t *testing.T) {
	cfg := config.Config{
		RSSURLs: []string{
			"https://example.com/a.xml",
			"https://example.com/b.xml",
		},
	}

	request := buildAgentRequest(cfg, "http")

	if request.Context.Trigger != "http" {
		t.Fatalf("expected trigger http, got %s", request.Context.Trigger)
	}
	if len(request.Context.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(request.Context.Sources))
	}
	if request.Context.Sources[0].Kind != "rss" || request.Context.Sources[0].Location != "https://example.com/a.xml" {
		t.Fatalf("unexpected first source: %+v", request.Context.Sources[0])
	}
}

func TestBuildAgentRequestFallsBackToDemoSource(t *testing.T) {
	request := buildAgentRequest(config.Config{}, "manual")

	if len(request.Context.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(request.Context.Sources))
	}
	if request.Context.Sources[0].Kind != "demo" {
		t.Fatalf("expected demo source, got %+v", request.Context.Sources[0])
	}
}

func TestBuildAgentRequestPrefersExplicitSources(t *testing.T) {
	request := buildAgentRequest(config.Config{
		Sources: []config.SourceConfig{
			{Name: "custom-demo", Kind: "demo", Location: "memory://custom", Priority: 9, IncludeInReport: false},
		},
		RSSURLs: []string{"https://example.com/a.xml"},
	}, "manual")

	if len(request.Context.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(request.Context.Sources))
	}
	if request.Context.Sources[0].Name != "custom-demo" || request.Context.Sources[0].Location != "memory://custom" {
		t.Fatalf("expected explicit source to be used, got %+v", request.Context.Sources[0])
	}
	if request.Context.Sources[0].Priority != 9 || request.Context.Sources[0].IncludeInReport {
		t.Fatalf("expected source strategy fields to be carried, got %+v", request.Context.Sources[0])
	}
}

func TestNewCrawlerFallsBackToDemoWhenSourceBuildFails(t *testing.T) {
	built := newCrawler(config.Config{
		Sources: []config.SourceConfig{
			{Name: "broken", Kind: "rss", Location: ""},
		},
	})

	wrapped, ok := built.(crawler.Crawler)
	if !ok {
		t.Fatalf("expected crawler, got %T", built)
	}
	items, err := wrapped.Fetch()
	if err != nil {
		t.Fatalf("expected demo fallback fetch to succeed, got %v", err)
	}
	if len(items) == 0 || items[0].SourceName != "demo" {
		t.Fatalf("expected demo fallback items, got %+v", items)
	}
}

type captureReportRepository struct {
	saveCalls int
	record    repository.ReportRecord
}

func (r *captureReportRepository) Save(ctx context.Context, record repository.ReportRecord) error {
	r.saveCalls++
	r.record = record
	return nil
}

func (r *captureReportRepository) Latest(ctx context.Context) (repository.ReportRecord, error) {
	return repository.ReportRecord{}, repository.ErrReportNotFound
}

func (r *captureReportRepository) List(ctx context.Context) ([]repository.ReportMetadata, error) {
	return nil, nil
}
