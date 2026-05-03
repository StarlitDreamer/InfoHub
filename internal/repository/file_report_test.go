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
		t.Fatalf("保存日报失败：%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "reports", "20260503-153000.md")); err != nil {
		t.Fatalf("期望 Markdown 日报文件存在：%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "items", "20260503-153000.json")); err != nil {
		t.Fatalf("期望 NewsItem JSON 文件存在：%v", err)
	}
}

func TestFileReportRepositoryLatestAndList(t *testing.T) {
	root := t.TempDir()
	repo := NewFileReportRepository(root)

	records := []ReportRecord{
		{
			GeneratedAt: time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC),
			Markdown:    "# 第一份日报",
			Items:       []model.NewsItem{{Title: "第一条"}},
		},
		{
			GeneratedAt: time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC),
			Markdown:    "# 第二份日报",
			Items:       []model.NewsItem{{Title: "第二条"}},
		},
	}

	for _, record := range records {
		if err := repo.Save(context.Background(), record); err != nil {
			t.Fatalf("保存日报失败：%v", err)
		}
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("读取日报列表失败：%v", err)
	}

	if len(list) != 2 || list[0].Name != "20260503-163000" {
		t.Fatalf("日报列表排序不符合预期：%+v", list)
	}

	latest, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("读取最新日报失败：%v", err)
	}

	if latest.Markdown != "# 第二份日报" {
		t.Fatalf("期望读取第二份日报，实际为 %s", latest.Markdown)
	}
}
