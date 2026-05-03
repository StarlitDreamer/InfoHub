package processor

import (
	"testing"

	"InfoHub-agent/internal/model"
)

func TestDeduplicateByTitleKeepsFirstItem(t *testing.T) {
	items := []model.NewsItem{
		{ID: 1, Title: "重复标题", Content: "第一条"},
		{ID: 2, Title: "重复标题", Content: "第二条"},
		{ID: 3, Title: "唯一标题", Content: "第三条"},
	}

	result := DeduplicateByTitle(items)

	if len(result) != 2 {
		t.Fatalf("期望 2 条去重结果，实际得到 %d 条", len(result))
	}

	if result[0].ID != 1 {
		t.Fatalf("期望保留第一条重复数据，实际保留 ID %d", result[0].ID)
	}
}
