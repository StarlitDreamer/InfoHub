package crawler

import (
	"testing"
	"time"

	"InfoHub-agent/internal/config"
)

func TestBuildFromSourcesReturnsSingleRSSCrawler(t *testing.T) {
	crawler, err := BuildFromSources([]config.SourceConfig{
		{
			Enabled:        true,
			Name:           "feed-a",
			Kind:           "rss",
			Location:       "https://example.com/a.xml",
			TimeoutSeconds: 10,
			Headers: map[string]string{
				"Authorization": "Bearer token",
			},
		},
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
	if rssCrawler.client == nil || rssCrawler.client.Timeout != 10*time.Second {
		t.Fatalf("expected source timeout to be applied, got %+v", rssCrawler.client)
	}
	if rssCrawler.headers["Authorization"] != "Bearer token" {
		t.Fatalf("expected source headers to be preserved, got %+v", rssCrawler.headers)
	}
}

func TestBuildFromSourcesReturnsSingleHTTPJSONCrawler(t *testing.T) {
	crawler, err := BuildFromSources([]config.SourceConfig{
		{Enabled: true, Name: "api-a", Kind: "http_json", Location: "https://example.com/api.json"},
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
		{Enabled: true, Name: "feed-a", Kind: "rss", Location: "https://example.com/a.xml"},
		{Enabled: true, Name: "demo", Kind: "demo", Location: "in-memory"},
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

func TestBuildFromSourcesSkipsDisabledSources(t *testing.T) {
	crawler, err := BuildFromSources([]config.SourceConfig{
		{Enabled: false, Name: "feed-a", Kind: "rss", Location: "https://example.com/a.xml"},
	}, FactoryOptions{})
	if err != nil {
		t.Fatalf("build crawler failed: %v", err)
	}

	if _, ok := crawler.(*DemoCrawler); !ok {
		t.Fatalf("expected demo fallback when all sources disabled, got %T", crawler)
	}
}

func TestBuildFromSourcesRejectsUnsupportedKind(t *testing.T) {
	_, err := BuildFromSources([]config.SourceConfig{
		{Enabled: true, Name: "github", Kind: "github", Location: "octocat/hello-world"},
	}, FactoryOptions{})
	if err == nil {
		t.Fatal("expected unsupported source kind error")
	}
}
