package crawler

import (
	"testing"
	"time"

	"InfoHub-agent/internal/config"
)

func TestBuildFromSourcesReturnsSingleRSSCrawler(t *testing.T) {
	crawler, err := BuildFromSources([]config.SourceConfig{
		{Name: "feed-a", Kind: "rss", Location: "https://example.com/a.xml"},
	}, FactoryOptions{
		RSSMaxItems:     10,
		RSSRecentWithin: 48 * time.Hour,
	})
	if err != nil {
		t.Fatalf("build crawler failed: %v", err)
	}

	rssCrawler, ok := crawler.(*RSSCrawler)
	if !ok {
		t.Fatalf("expected rss crawler, got %T", crawler)
	}
	if rssCrawler.url != "https://example.com/a.xml" {
		t.Fatalf("expected rss location to be preserved, got %s", rssCrawler.url)
	}
}

func TestBuildFromSourcesReturnsSingleHTTPJSONCrawler(t *testing.T) {
	crawler, err := BuildFromSources([]config.SourceConfig{
		{Name: "api-a", Kind: "http_json", Location: "https://example.com/api.json"},
	}, FactoryOptions{})
	if err != nil {
		t.Fatalf("build crawler failed: %v", err)
	}

	httpJSONCrawler, ok := crawler.(*HTTPJSONCrawler)
	if !ok {
		t.Fatalf("expected http json crawler, got %T", crawler)
	}
	if httpJSONCrawler.url != "https://example.com/api.json" {
		t.Fatalf("expected http json location to be preserved, got %s", httpJSONCrawler.url)
	}
}

func TestBuildFromSourcesReturnsMultiCrawlerForMultipleSources(t *testing.T) {
	crawler, err := BuildFromSources([]config.SourceConfig{
		{Name: "feed-a", Kind: "rss", Location: "https://example.com/a.xml"},
		{Name: "demo", Kind: "demo", Location: "in-memory"},
	}, FactoryOptions{})
	if err != nil {
		t.Fatalf("build crawler failed: %v", err)
	}

	multiCrawler, ok := crawler.(*MultiCrawler)
	if !ok {
		t.Fatalf("expected multi crawler, got %T", crawler)
	}
	if len(multiCrawler.crawlers) != 2 {
		t.Fatalf("expected 2 child crawlers, got %d", len(multiCrawler.crawlers))
	}
}

func TestBuildFromSourcesFallsBackToDemoWhenEmpty(t *testing.T) {
	crawler, err := BuildFromSources(nil, FactoryOptions{})
	if err != nil {
		t.Fatalf("build crawler failed: %v", err)
	}

	if _, ok := crawler.(*DemoCrawler); !ok {
		t.Fatalf("expected demo crawler, got %T", crawler)
	}
}

func TestBuildFromSourcesRejectsUnsupportedKind(t *testing.T) {
	_, err := BuildFromSources([]config.SourceConfig{
		{Name: "github", Kind: "github", Location: "octocat/hello-world"},
	}, FactoryOptions{})
	if err == nil {
		t.Fatal("expected unsupported source kind error")
	}
}
