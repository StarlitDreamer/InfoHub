package crawler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRSSCrawlerFetchesItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss><channel><title>测试源</title><item><title>标题一</title><link>https://example.com/a</link><description>摘要一</description><pubDate>Sun, 03 May 2026 10:00:00 +0800</pubDate></item></channel></rss>`))
	}))
	defer server.Close()

	items, err := NewRSSCrawler(server.URL, server.Client(), nil, RSSOptions{}).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(items) != 1 || items[0].Title != "标题一" {
		t.Fatalf("unexpected fetch result: %+v", items)
	}
}

func TestRSSCrawlerCleansHTMLDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss><channel><title><![CDATA[ 测试&nbsp;源 ]]></title><item><title><![CDATA[ <b>标题&nbsp;一</b> ]]></title><link>https://example.com/a</link><description><![CDATA[<p>摘要&nbsp;&amp;&nbsp;内容</p>]]></description><pubDate>Sun, 03 May 2026 10:00:00 +0800</pubDate></item></channel></rss>`))
	}))
	defer server.Close()

	items, err := NewRSSCrawler(server.URL, server.Client(), nil, RSSOptions{}).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if items[0].Title != "标题 一" {
		t.Fatalf("unexpected cleaned title: %q", items[0].Title)
	}
	if items[0].Content != "摘要 & 内容" {
		t.Fatalf("unexpected cleaned content: %q", items[0].Content)
	}
	if items[0].Source != "测试 源" {
		t.Fatalf("unexpected cleaned source: %q", items[0].Source)
	}
}

func TestRSSCrawlerFiltersByRecentWindowAndMaxItems(t *testing.T) {
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.TrimSpace(`<?xml version="1.0"?>
<rss><channel><title>测试源</title>
<item><title>最新</title><link>https://example.com/a</link><description>最新摘要</description><pubDate>Sun, 03 May 2026 11:00:00 +0000</pubDate></item>
<item><title>次新</title><link>https://example.com/b</link><description>次新摘要</description><pubDate>Sun, 03 May 2026 10:00:00 +0000</pubDate></item>
<item><title>太旧</title><link>https://example.com/c</link><description>太旧摘要</description><pubDate>Thu, 30 Apr 2026 10:00:00 +0000</pubDate></item>
</channel></rss>`)))
	}))
	defer server.Close()

	items, err := NewRSSCrawler(server.URL, server.Client(), nil, RSSOptions{
		MaxItems:     2,
		RecentWithin: 48 * time.Hour,
		Now: func() time.Time {
			return now
		},
	}).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items after trimming, got %d", len(items))
	}
	if items[0].Title != "最新" || items[1].Title != "次新" {
		t.Fatalf("unexpected filtered result: %+v", items)
	}
}

func TestRSSCrawlerIncludesFeedURLInStatusErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := NewRSSCrawler(server.URL, server.Client(), nil, RSSOptions{}).Fetch(context.Background())
	if err == nil {
		t.Fatal("expected fetch to fail")
	}

	message := err.Error()
	if !strings.Contains(message, server.URL) || !strings.Contains(message, "status code 404") {
		t.Fatalf("expected url and status code in error, got %s", message)
	}
}

func TestRSSCrawlerPrefersEncodedContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss xmlns:content="http://purl.org/rss/1.0/modules/content/"><channel><title>测试源</title><item><title>标题一</title><link>https://example.com/a</link><description>短摘要</description><content:encoded><![CDATA[<div><p>更完整的正文内容</p></div>]]></content:encoded><pubDate>Sun, 03 May 2026 10:00:00 +0800</pubDate></item></channel></rss>`))
	}))
	defer server.Close()

	items, err := NewRSSCrawler(server.URL, server.Client(), nil, RSSOptions{}).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(items) != 1 || items[0].Content != "更完整的正文内容" {
		t.Fatalf("expected encoded content to win, got %+v", items)
	}
}

func TestRSSCrawlerRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRSSCrawler("https://example.com/rss.xml", http.DefaultClient, nil, RSSOptions{}).Fetch(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
