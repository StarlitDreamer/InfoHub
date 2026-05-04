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

func TestHTTPJSONCrawlerFetchesItemsFromObjectEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
  "items": [
    {
      "title": "Test item",
      "content": "<p>Summary</p>",
      "source": "API Feed",
      "url": "https://example.com/a",
      "publish_time": "2026-05-04T09:00:00Z",
      "tags": ["AI", " News "],
      "score": 4
    }
  ]
}`))
	}))
	defer server.Close()

	items, err := NewHTTPJSONCrawler(server.URL, server.Client(), nil).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Test item" || items[0].Content != "Summary" {
		t.Fatalf("unexpected item cleanup result: %+v", items[0])
	}
	if items[0].Source != "API Feed" || items[0].URL != "https://example.com/a" {
		t.Fatalf("unexpected item metadata: %+v", items[0])
	}
	if items[0].PublishTime != time.Date(2026, 5, 4, 9, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected publish time: %s", items[0].PublishTime)
	}
	if len(items[0].Tags) != 2 || items[0].Tags[1] != "News" {
		t.Fatalf("unexpected tags: %+v", items[0].Tags)
	}
	if items[0].Score != 4 {
		t.Fatalf("unexpected score: %v", items[0].Score)
	}
}

func TestHTTPJSONCrawlerFetchesItemsFromArrayEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
  {
    "title": "Array item",
    "content": "Array summary",
    "source": "Array Feed",
    "url": "https://example.com/b",
    "publish_time": "2026-05-04 09:30:00"
  }
]`))
	}))
	defer server.Close()

	items, err := NewHTTPJSONCrawler(server.URL, server.Client(), nil).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Array item" {
		t.Fatalf("unexpected array envelope result: %+v", items)
	}
}

func TestHTTPJSONCrawlerIncludesSourceURLInStatusErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := NewHTTPJSONCrawler(server.URL, server.Client(), nil).Fetch(context.Background())
	if err == nil {
		t.Fatal("expected fetch to fail")
	}

	message := err.Error()
	if !strings.Contains(message, server.URL) || !strings.Contains(message, "status code 502") {
		t.Fatalf("expected url and status code in error, got %s", message)
	}
}

func TestHTTPJSONCrawlerSendsConfiguredHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != "secret" {
			t.Fatalf("expected header X-Token=secret, got %s", r.Header.Get("X-Token"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	_, err := NewHTTPJSONCrawler(server.URL, server.Client(), map[string]string{
		"X-Token": "secret",
	}).Fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
}

func TestHTTPJSONCrawlerRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewHTTPJSONCrawler("https://example.com/data.json", http.DefaultClient, nil).Fetch(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
