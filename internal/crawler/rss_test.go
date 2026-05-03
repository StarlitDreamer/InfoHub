package crawler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRSSCrawlerFetchesItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss><channel><title>测试源</title><item><title>标题一</title><link>https://example.com/a</link><description>摘要一</description><pubDate>Sun, 03 May 2026 10:00:00 +0800</pubDate></item></channel></rss>`))
	}))
	defer server.Close()

	items, err := NewRSSCrawler(server.URL, server.Client()).Fetch()
	if err != nil {
		t.Fatalf("采集 RSS 失败：%v", err)
	}

	if len(items) != 1 || items[0].Title != "标题一" {
		t.Fatalf("RSS 解析结果不符合预期：%+v", items)
	}
}
