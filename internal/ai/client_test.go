package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestHTTPClientSummarize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("缺少 Authorization 请求头")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"【标题】测试\n【发生了什么】已摘要\n【为什么重要】重要\n【影响】关注\n【评分】4"}}]}`))
	}))
	defer server.Close()

	item, err := NewHTTPClient(server.URL, "test-key", "test-model", server.Client()).Summarize(model.NewsItem{Title: "测试", Content: "内容"})
	if err != nil {
		t.Fatalf("AI 摘要失败：%v", err)
	}

	if item.Score != 4 {
		t.Fatalf("期望评分 4，实际得到 %.0f", item.Score)
	}
}
