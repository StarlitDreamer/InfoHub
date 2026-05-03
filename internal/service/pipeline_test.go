package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestPipelineSkipsSeenItemsAcrossRuns(t *testing.T) {
	store := newMemoryDedupStore()
	item := model.NewsItem{
		Title:   "already seen item",
		URL:     "https://example.com/seen",
		Content: "same content",
	}
	pipeline := NewPipeline(
		staticServiceCrawler{items: []model.NewsItem{item}},
		echoAI{},
	).WithDedupStore(store)

	first, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected 1 item on first run, got %d", len(first))
	}

	second, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("expected 0 items on second run, got %d", len(second))
	}
}

func TestPipelineSkipsItemsWhenAnyStoredDedupKeyMatches(t *testing.T) {
	store := newMemoryDedupStore()
	firstItem := model.NewsItem{
		Title:   "shared title",
		URL:     "https://example.com/a",
		Content: "first content",
	}
	secondItem := model.NewsItem{
		Title:   "shared title",
		URL:     "https://example.com/b",
		Content: "second content",
	}
	pipeline := NewPipeline(
		staticServiceCrawler{items: []model.NewsItem{firstItem}},
		echoAI{},
	).WithDedupStore(store)

	first, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected 1 item on first run, got %d", len(first))
	}

	pipeline = NewPipeline(
		staticServiceCrawler{items: []model.NewsItem{secondItem}},
		echoAI{},
	).WithDedupStore(store)
	second, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("expected title fingerprint match to skip item, got %d", len(second))
	}
}

func TestPipelineDeduplicatesWithinRunUsingCompositeFingerprints(t *testing.T) {
	pipeline := NewPipeline(
		staticServiceCrawler{items: []model.NewsItem{
			{Title: "alpha", URL: "https://example.com/1", Content: "same body"},
			{Title: "beta", URL: "https://example.com/2", Content: " same   body "},
			{Title: "gamma", URL: "https://example.com/3", Content: "different body"},
		}},
		echoAI{},
	)

	result, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("pipeline run failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 items after composite deduplication, got %d", len(result))
	}
	if result[0].Title != "alpha" || result[1].Title != "gamma" {
		t.Fatalf("unexpected within-run deduplication result: %+v", result)
	}
}

func TestPipelineNormalizesEmptySummaryFields(t *testing.T) {
	item := model.NewsItem{
		Title:       "test title",
		Content:     "original content",
		Source:      "test source",
		URL:         "https://example.com/a",
		PublishTime: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	}
	pipeline := NewPipeline(
		staticServiceCrawler{items: []model.NewsItem{item}},
		emptyAI{},
	)

	result, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("pipeline run failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}

	summarized := result[0]
	if summarized.Title != item.Title {
		t.Fatalf("expected title fallback to original title, got %s", summarized.Title)
	}
	if !strings.Contains(summarized.Content, "【标题】test title") {
		t.Fatalf("expected fallback structured summary, got %s", summarized.Content)
	}
	if summarized.Score != 1 {
		t.Fatalf("expected score to be clamped to 1, got %v", summarized.Score)
	}
	if summarized.Source != item.Source || summarized.URL != item.URL || summarized.PublishTime != item.PublishTime {
		t.Fatalf("expected source metadata fallback, got %+v", summarized)
	}
}

type memoryDedupStore struct {
	keys map[string]struct{}
}

func newMemoryDedupStore() *memoryDedupStore {
	return &memoryDedupStore{keys: map[string]struct{}{}}
}

func (s *memoryDedupStore) Seen(ctx context.Context, key string) (bool, error) {
	_, ok := s.keys[key]
	return ok, nil
}

func (s *memoryDedupStore) Mark(ctx context.Context, key string) error {
	s.keys[key] = struct{}{}
	return nil
}

type staticServiceCrawler struct {
	items []model.NewsItem
}

func (c staticServiceCrawler) Fetch() ([]model.NewsItem, error) {
	return c.items, nil
}

type echoAI struct{}

func (a echoAI) Summarize(item model.NewsItem) (model.NewsItem, error) {
	item.Score = 1
	return item, nil
}

type emptyAI struct{}

func (a emptyAI) Summarize(item model.NewsItem) (model.NewsItem, error) {
	return model.NewsItem{
		Score: 0,
	}, nil
}
