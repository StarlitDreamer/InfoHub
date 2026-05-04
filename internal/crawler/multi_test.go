package crawler

import (
	"context"
	"errors"
	"strings"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestMultiCrawlerFetchesAllSuccessfulSources(t *testing.T) {
	crawler := NewMultiCrawler([]Crawler{
		staticCrawler{items: []model.NewsItem{{Title: "第一条"}}},
		staticCrawler{items: []model.NewsItem{{Title: "第二条"}}},
	})

	items, err := crawler.Fetch(context.Background())
	if err != nil {
		t.Fatalf("multi source fetch failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if len(crawler.Warnings()) != 0 {
		t.Fatalf("expected no warnings, got %+v", crawler.Warnings())
	}
}

func TestMultiCrawlerKeepsSuccessfulSourcesWhenSomeFail(t *testing.T) {
	crawler := NewMultiCrawler([]Crawler{
		errorCrawler{label: "feed-a", err: errors.New("fetch failed")},
		staticCrawler{items: []model.NewsItem{{Title: "成功数据"}}},
	})

	items, err := crawler.Fetch(context.Background())
	if err != nil {
		t.Fatalf("expected partial failure to keep successful items: %v", err)
	}
	if len(items) != 1 || items[0].Title != "成功数据" {
		t.Fatalf("unexpected partial fetch result: %+v", items)
	}
	warnings := crawler.Warnings()
	if len(warnings) != 1 || !strings.Contains(warnings[0], "feed-a: fetch failed") {
		t.Fatalf("expected partial failure warning, got %+v", warnings)
	}
}

func TestMultiCrawlerReturnsSourceLabelsWhenAllSourcesFail(t *testing.T) {
	crawler := NewMultiCrawler([]Crawler{
		errorCrawler{label: "feed-a", err: errors.New("fetch failed")},
		errorCrawler{label: "feed-b", err: errors.New("timed out")},
	})

	_, err := crawler.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error when all sources fail")
	}
	message := err.Error()
	if !strings.Contains(message, "feed-a: fetch failed") || !strings.Contains(message, "feed-b: timed out") {
		t.Fatalf("expected source labels in error, got %s", message)
	}
}

type staticCrawler struct {
	items []model.NewsItem
}

func (c staticCrawler) Fetch(ctx context.Context) ([]model.NewsItem, error) {
	return c.items, nil
}

type errorCrawler struct {
	label string
	err   error
}

func (c errorCrawler) Fetch(ctx context.Context) ([]model.NewsItem, error) {
	return nil, c.err
}

func (c errorCrawler) String() string {
	return c.label
}
