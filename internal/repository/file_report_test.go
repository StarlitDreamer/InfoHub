package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestFileReportRepositorySave(t *testing.T) {
	root := t.TempDir()
	generatedAt := time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC)
	repo := NewFileReportRepository(root)

	err := repo.Save(context.Background(), ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    "# 今日信息",
		Items:       []model.NewsItem{{Title: "测试标题"}},
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "reports", "20260503-153000.md")); err != nil {
		t.Fatalf("expected markdown file to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "items", "20260503-153000.json")); err != nil {
		t.Fatalf("expected items file to exist: %v", err)
	}
}

func TestFileReportRepositoryLatestAndList(t *testing.T) {
	root := t.TempDir()
	repo := NewFileReportRepository(root)

	records := []ReportRecord{
		{
			GeneratedAt: time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC),
			Markdown:    "# 今日信息\n\n## ⭐⭐⭐\n- 标题：第一条\n- 摘要：摘要一\n",
			Items:       []model.NewsItem{{Title: "第一条"}},
		},
		{
			GeneratedAt: time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC),
			Markdown:    "# 今日信息\n\n## ⭐⭐⭐\n- 标题：第二条\n- 摘要：摘要二\n\n## ⭐⭐\n- 标题：第三条\n- 摘要：摘要三\n",
			Items:       []model.NewsItem{{Title: "第二条"}, {Title: "第三条"}, {Title: "库存但未展示"}},
		},
	}

	for _, record := range records {
		if err := repo.Save(context.Background(), record); err != nil {
			t.Fatalf("save failed: %v", err)
		}
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(list) != 2 || list[0].Name != "20260503-163000" {
		t.Fatalf("unexpected list order: %+v", list)
	}
	if list[0].ItemCount != 3 || list[0].DisplayCount != 2 {
		t.Fatalf("unexpected list summary: %+v", list[0])
	}

	latest, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("latest failed: %v", err)
	}
	if latest.Markdown != records[1].Markdown {
		t.Fatalf("expected latest markdown to match second record, got %s", latest.Markdown)
	}
}
