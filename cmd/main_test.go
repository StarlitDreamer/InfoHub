package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/config"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/repository"
	"InfoHub-agent/internal/service"
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
	if len(result.TopPriorityItems) == 0 {
		t.Fatalf("expected run summary titles, got %+v", result)
	}
	if len(result.DecisionSummary) == 0 {
		t.Fatalf("expected run decision summary, got %+v", result)
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

	result, err := runReportWithRepository(context.Background(), cfg, repo, buildAgentRequest(cfg, "manual"))
	if err != nil {
		t.Fatalf("runReportWithRepository failed: %v", err)
	}

	if result.ItemCount != 3 {
		t.Fatalf("expected 3 sorted items, got %d", result.ItemCount)
	}
	if result.DisplayCount != 2 {
		t.Fatalf("expected 2 displayed items, got %d", result.DisplayCount)
	}
	expectedOrder := []string{
		"开源模型发布新版本",
		"数据库社区讨论新索引策略",
		"云厂商推出开发者工具更新",
	}
	if result.GeneratedAt != fixedNow {
		t.Fatalf("expected generated time %s, got %s", fixedNow, result.GeneratedAt)
	}
	if len(result.TopPriorityItems) != 3 || result.TopPriorityItems[0] != expectedOrder[0] {
		t.Fatalf("expected top priority titles in result, got %+v", result.TopPriorityItems)
	}
	if len(result.DecisionSummary) != 3 || result.DecisionSummary[0].Title != expectedOrder[0] {
		t.Fatalf("expected decision summary in result, got %+v", result.DecisionSummary)
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

	for index, title := range expectedOrder {
		if repo.record.Items[index].Title != title {
			t.Fatalf("expected sorted item %d to be %s, got %s", index, title, repo.record.Items[index].Title)
		}
	}

	if strings.Count(repo.record.Markdown, "## ⭐") != 2 {
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
	if _, err := runReportWithRepository(context.Background(), cfg, firstRepo, buildAgentRequest(cfg, "manual")); err != nil {
		t.Fatalf("first runReportWithRepository failed: %v", err)
	}

	secondRepo := &captureReportRepository{}
	result, err := runReportWithRepository(context.Background(), cfg, secondRepo, buildAgentRequest(cfg, "manual"))
	if err != nil {
		t.Fatalf("second runReportWithRepository failed: %v", err)
	}

	if result.ItemCount != 0 || result.DisplayCount != 0 {
		t.Fatalf("expected second run to generate empty result, got %+v", result)
	}
	if len(result.TopPriorityItems) != 0 || len(result.DecisionSummary) != 0 {
		t.Fatalf("expected empty run summary, got %+v", result)
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

func TestBuildAgentRequestCarriesPreference(t *testing.T) {
	request := buildAgentRequest(config.Config{
		PreferenceTags:     []string{"AI"},
		PreferenceSources:  []string{"openai-news"},
		PreferenceKeywords: []string{"agent"},
	}, "manual")

	if len(request.Context.Preference.Tags) != 1 || request.Context.Preference.Tags[0] != "AI" {
		t.Fatalf("expected preference tags to be carried, got %+v", request.Context.Preference)
	}
	if len(request.Context.Preference.Sources) != 1 || request.Context.Preference.Sources[0] != "openai-news" {
		t.Fatalf("expected preference sources to be carried, got %+v", request.Context.Preference)
	}
	if len(request.Context.Preference.Keywords) != 1 || request.Context.Preference.Keywords[0] != "agent" {
		t.Fatalf("expected preference keywords to be carried, got %+v", request.Context.Preference)
	}
}

func TestBuildAgentRequestCarriesConfiguredPreferenceWeights(t *testing.T) {
	request := buildAgentRequest(config.Config{
		PreferenceTagWeight:     2.0,
		PreferenceSourceWeight:  1.5,
		PreferenceKeywordWeight: 0.9,
	}, "manual")

	if request.Context.Preference.Weights.TagMatch != 2.0 || request.Context.Preference.Weights.SourceMatch != 1.5 || request.Context.Preference.Weights.KeywordMatch != 0.9 {
		t.Fatalf("expected configured preference weights, got %+v", request.Context.Preference.Weights)
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
	items, err := wrapped.Fetch(context.Background())
	if err != nil {
		t.Fatalf("expected demo fallback fetch to succeed, got %v", err)
	}
	if len(items) == 0 || items[0].SourceName != "demo" {
		t.Fatalf("expected demo fallback items, got %+v", items)
	}
}

func TestBuildSchedulerUsesCronWhenConfigured(t *testing.T) {
	task, err := buildScheduler(config.Config{
		ScheduleCron:     "*/10 * * * *",
		ScheduleInterval: time.Hour,
	}, func(context.Context) error { return nil })
	if err != nil {
		t.Fatalf("expected cron scheduler to build, got %v", err)
	}
	if task == nil {
		t.Fatal("expected scheduler instance")
	}
}

func TestBuildSchedulerRejectsInvalidCron(t *testing.T) {
	_, err := buildScheduler(config.Config{
		ScheduleCron: "invalid-cron",
	}, func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected invalid cron to fail")
	}
}

func TestMergePreferenceOverridesConfiguredValuesWhenRequestProvided(t *testing.T) {
	merged := mergePreference(
		service.UserPreference{
			Tags:     []string{"AI"},
			Sources:  []string{"config-source"},
			Keywords: []string{"workflow"},
			Weights: service.PreferenceWeights{
				TagMatch:     1.2,
				SourceMatch:  1.0,
				KeywordMatch: 0.6,
			},
		},
		service.UserPreference{
			Tags:    []string{"数据库"},
			Sources: []string{"request-source"},
			Weights: service.PreferenceWeights{
				SourceMatch: 1.5,
			},
		},
	)

	if len(merged.Tags) != 1 || merged.Tags[0] != "数据库" {
		t.Fatalf("expected request tags to override config, got %+v", merged)
	}
	if len(merged.Sources) != 1 || merged.Sources[0] != "request-source" {
		t.Fatalf("expected request sources to override config, got %+v", merged)
	}
	if len(merged.Keywords) != 1 || merged.Keywords[0] != "workflow" {
		t.Fatalf("expected missing request keywords to keep config value, got %+v", merged)
	}
	if merged.Weights.TagMatch != 1.2 || merged.Weights.SourceMatch != 1.5 || merged.Weights.KeywordMatch != 0.6 {
		t.Fatalf("expected weights to merge correctly, got %+v", merged.Weights)
	}
}

func TestResolveUserPreferenceReturnsStoredRecord(t *testing.T) {
	repo := &preferenceRepoStub{
		record: repository.UserPreferenceRecord{
			UserID:   "alice",
			Tags:     []string{"AI"},
			Sources:  []string{"openai-news"},
			Keywords: []string{"agent"},
			Weights: repository.PreferenceWeightValue{
				Tag:     1.7,
				Source:  1.2,
				Keyword: 0.9,
			},
		},
	}

	preference, err := resolveUserPreference(context.Background(), repo, "alice")
	if err != nil {
		t.Fatalf("expected stored preference, got %v", err)
	}
	if len(preference.Tags) != 1 || preference.Weights.TagMatch != 1.7 {
		t.Fatalf("unexpected resolved preference: %+v", preference)
	}
}

func TestResolveUserPreferenceIgnoresMissingUser(t *testing.T) {
	preference, err := resolveUserPreference(context.Background(), &preferenceRepoStub{err: repository.ErrUserPreferenceNotFound}, "missing")
	if err != nil {
		t.Fatalf("expected missing user to be ignored, got %v", err)
	}
	if !preference.IsZero() {
		t.Fatalf("expected empty preference, got %+v", preference)
	}
}

func TestResolveUserPreferenceReturnsUnexpectedError(t *testing.T) {
	expectedErr := errors.New("boom")
	_, err := resolveUserPreference(context.Background(), &preferenceRepoStub{err: expectedErr}, "alice")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestNewUserPreferenceRepositoryFallsBackToFile(t *testing.T) {
	repo, closeFn, err := newUserPreferenceRepository(config.Config{
		StorageDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("expected file preference repository, got %v", err)
	}
	defer closeFn()

	if _, ok := repo.(*repository.FileUserPreferenceRepository); !ok {
		t.Fatalf("expected file preference repository, got %T", repo)
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

func (r *captureReportRepository) Get(ctx context.Context, name string) (repository.ReportRecord, error) {
	return repository.ReportRecord{}, repository.ErrReportNotFound
}

func (r *captureReportRepository) List(ctx context.Context) ([]repository.ReportMetadata, error) {
	return nil, nil
}

type preferenceRepoStub struct {
	record repository.UserPreferenceRecord
	err    error
}

func (r *preferenceRepoStub) Save(ctx context.Context, record repository.UserPreferenceRecord) error {
	r.record = record
	return r.err
}

func (r *preferenceRepoStub) Get(ctx context.Context, userID string) (repository.UserPreferenceRecord, error) {
	if r.err != nil {
		return repository.UserPreferenceRecord{}, r.err
	}
	return r.record, nil
}
