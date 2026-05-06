package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStackExchangeSearchCrawlerSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "agent" {
			t.Fatalf("expected query agent, got %s", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"title":"Agent orchestration","body":"How to build an agent workflow","link":"https://stackoverflow.com/q/1","score":12,"creation_date":1770000000}]}`))
	}))
	defer server.Close()

	oldBaseURL := stackExchangeSearchBaseURL
	stackExchangeSearchBaseURL = server.URL
	defer func() { stackExchangeSearchBaseURL = oldBaseURL }()

	items, err := NewStackExchangeSearchCrawler("stackoverflow", 5, server.Client(), nil).Search(context.Background(), "agent")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Channel != "stack_overflow" || items[0].SourceScore != 12 {
		t.Fatalf("unexpected item mapping: %+v", items[0])
	}
}

func TestRedditSearchCrawlerSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/r/programming/search.json") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"children":[{"data":{"title":"Agent frameworks","selftext":"Useful thread","permalink":"/r/programming/comments/1","created_utc":1770000100,"score":42}}]}}`))
	}))
	defer server.Close()

	oldBaseURL := redditSearchBaseURL
	redditSearchBaseURL = server.URL
	defer func() { redditSearchBaseURL = oldBaseURL }()

	items, err := NewRedditSearchCrawler("programming", 5, server.Client(), nil).Search(context.Background(), "agent")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Channel != "reddit" || items[0].SourceScore != 42 {
		t.Fatalf("unexpected item mapping: %+v", items[0])
	}
}

func TestRSSSearchCrawlerFiltersByKeyword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<rss><channel><title>Feed</title><item><title>Agent release</title><link>https://example.com/a</link><description>Agent notes</description><pubDate>Sun, 03 May 2026 10:00:00 +0800</pubDate></item><item><title>Other</title><link>https://example.com/b</link><description>Misc</description><pubDate>Sun, 03 May 2026 11:00:00 +0800</pubDate></item></channel></rss>`))
	}))
	defer server.Close()

	items, err := NewRSSSearchCrawler(server.URL, 5, server.Client(), nil, RSSOptions{Now: func() time.Time { return time.Now() }}).Search(context.Background(), "agent")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(items) != 1 || items[0].Channel != "rss_search" {
		t.Fatalf("expected rss search result, got %+v", items)
	}
}
