package service

import (
	"context"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestPipelineSkipsSeenItems(t *testing.T) {
	store := newMemoryDedupStore()
	item := model.NewsItem{Title: "已处理信息", URL: "https://example.com/seen"}
	pipeline := NewPipeline(
		staticServiceCrawler{items: []model.NewsItem{item}},
		echoAI{},
	).WithDedupStore(store)

	first, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("第一次运行失败：%v", err)
	}

	if len(first) != 1 {
		t.Fatalf("期望第一次输出 1 条，实际为 %d", len(first))
	}

	second, err := pipeline.RunContext(context.Background())
	if err != nil {
		t.Fatalf("第二次运行失败：%v", err)
	}

	if len(second) != 0 {
		t.Fatalf("期望第二次跳过已处理信息，实际输出 %d 条", len(second))
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
