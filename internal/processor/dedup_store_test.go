package processor

import (
	"context"
	"path/filepath"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestFileDedupStorePersistsKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seen.json")
	store := NewFileDedupStore(path)
	key := DedupKey(model.NewsItem{URL: "https://example.com/a"})

	seen, err := store.Seen(context.Background(), key)
	if err != nil {
		t.Fatalf("读取去重状态失败：%v", err)
	}

	if seen {
		t.Fatal("新 key 不应被判断为已处理")
	}

	if err := store.Mark(context.Background(), key); err != nil {
		t.Fatalf("写入去重状态失败：%v", err)
	}

	nextStore := NewFileDedupStore(path)
	seen, err = nextStore.Seen(context.Background(), key)
	if err != nil {
		t.Fatalf("重新读取去重状态失败：%v", err)
	}

	if !seen {
		t.Fatal("期望跨 store 实例仍能识别已处理 key")
	}
}

func TestDedupKeyFallsBackToTitle(t *testing.T) {
	left := DedupKey(model.NewsItem{Title: "同一标题"})
	right := DedupKey(model.NewsItem{Title: "同一标题"})

	if left == "" || left != right {
		t.Fatal("期望标题生成稳定去重 key")
	}
}
