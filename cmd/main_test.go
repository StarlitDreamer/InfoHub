package main

import (
	"context"
	"testing"
	"time"

	"InfoHub-agent/internal/config"
)

func TestRunRejectsUnknownMode(t *testing.T) {
	err := run(context.Background(), config.Config{}, []string{"unknown"})

	if err == nil {
		t.Fatal("期望未知运行模式返回错误")
	}
}

func TestRunOnceMode(t *testing.T) {
	cfg := config.Config{
		ScheduleInterval: time.Hour,
		StorageDir:       t.TempDir(),
		DedupStorePath:   t.TempDir() + "/seen.json",
	}

	if err := run(context.Background(), cfg, []string{"run-once"}); err != nil {
		t.Fatalf("run-once 执行失败：%v", err)
	}
}

func TestRunReportReturnsItemCount(t *testing.T) {
	cfg := config.Config{
		ScheduleInterval: time.Hour,
		StorageDir:       t.TempDir(),
		DedupStorePath:   t.TempDir() + "/seen.json",
	}

	result, err := runReport(context.Background(), cfg)
	if err != nil {
		t.Fatalf("生成日报失败：%v", err)
	}

	if result.ItemCount != 3 {
		t.Fatalf("期望新增 3 条信息，实际为 %d", result.ItemCount)
	}

	result, err = runReport(context.Background(), cfg)
	if err != nil {
		t.Fatalf("第二次生成日报失败：%v", err)
	}

	if result.ItemCount != 0 {
		t.Fatalf("期望第二次无新增信息，实际为 %d", result.ItemCount)
	}
}
