package crawler

import (
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

	items, err := NewRSSCrawler(server.URL, server.Client(), RSSOptions{}).Fetch()
	if err != nil {
		t.Fatalf("采集 RSS 失败：%v", err)
	}

	if len(items) != 1 || items[0].Title != "标题一" {
		t.Fatalf("RSS 解析结果不符合预期：%+v", items)
	}
}

func TestRSSCrawlerCleansHTMLDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss><channel><title><![CDATA[ 测试&nbsp;源 ]]></title><item><title><![CDATA[ <b>标题&nbsp;一</b> ]]></title><link>https://example.com/a</link><description><![CDATA[<p>摘要&nbsp;&amp;&nbsp;内容</p>]]></description><pubDate>Sun, 03 May 2026 10:00:00 +0800</pubDate></item></channel></rss>`))
	}))
	defer server.Close()

	items, err := NewRSSCrawler(server.URL, server.Client(), RSSOptions{}).Fetch()
	if err != nil {
		t.Fatalf("采集 RSS 失败：%v", err)
	}

	if items[0].Title != "标题 一" {
		t.Fatalf("标题清洗结果不符合预期：%q", items[0].Title)
	}

	if items[0].Content != "摘要 & 内容" {
		t.Fatalf("正文清洗结果不符合预期：%q", items[0].Content)
	}

	if items[0].Source != "测试 源" {
		t.Fatalf("来源清洗结果不符合预期：%q", items[0].Source)
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

	items, err := NewRSSCrawler(server.URL, server.Client(), RSSOptions{
		MaxItems:     2,
		RecentWithin: 48 * time.Hour,
		Now: func() time.Time {
			return now
		},
	}).Fetch()
	if err != nil {
		t.Fatalf("采集 RSS 失败：%v", err)
	}

	if len(items) != 2 {
		t.Fatalf("期望保留 2 条数据，实际为 %d", len(items))
	}

	if items[0].Title != "最新" || items[1].Title != "次新" {
		t.Fatalf("过滤与排序结果不符合预期：%+v", items)
	}
}
