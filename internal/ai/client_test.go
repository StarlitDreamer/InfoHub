package ai

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestHTTPClientAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("expected authorization header")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"{\\\"tags\\\":[\\\"AI\\\",\\\"OpenAI\\\"],\\\"summary\\\":\\\"【标题】测试\\\\n【发生了什么】已摘要\\\\n【为什么重要】重要\\\\n【影响】关注\\\\n【评分】4\\\",\\\"score\\\":4}\"}}]}"))
	}))
	defer server.Close()

	analysis, err := NewHTTPClient(server.URL, "test-key", "test-model", server.Client()).Analyze(model.NewsItem{
		Title:   "测试",
		Content: "内容",
	})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	if analysis.Score != 4 {
		t.Fatalf("expected score 4, got %.0f", analysis.Score)
	}
	if len(analysis.Tags) != 2 || analysis.Tags[0] != "AI" {
		t.Fatalf("unexpected tags: %+v", analysis.Tags)
	}
	if !strings.Contains(analysis.Summary, "【标题】测试") {
		t.Fatalf("unexpected summary: %s", analysis.Summary)
	}
}

func TestHTTPClientAnalyzeFallsBackWhenSummaryIsNotStructured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"tags\":[\"AI\"],\"summary\":\"plain text summary\",\"score\":7}"}}]}`))
	}))
	defer server.Close()

	item := model.NewsItem{
		Title:   "测试标题",
		Content: "测试内容",
	}
	analysis, err := NewHTTPClient(server.URL, "test-key", "test-model", server.Client()).Analyze(item)
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	if analysis.Score != 5 {
		t.Fatalf("expected score to be clamped to 5, got %.0f", analysis.Score)
	}
	for _, label := range []string{"【标题】", "【发生了什么】", "【为什么重要】", "【影响】", "【评分】"} {
		if !strings.Contains(analysis.Summary, label) {
			t.Fatalf("expected fallback summary to contain %s, got %s", label, analysis.Summary)
		}
	}
	if !strings.Contains(analysis.Summary, item.Title) || !strings.Contains(analysis.Summary, item.Content) {
		t.Fatalf("expected fallback summary to use item fields, got %s", analysis.Summary)
	}
}
