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
		t.Fatal("expected unknown mode to return an error")
	}
}

func TestRunOnceMode(t *testing.T) {
	cfg := config.Config{
		ScheduleInterval: time.Hour,
		StorageDir:       t.TempDir(),
		DedupStorePath:   t.TempDir() + "/seen.json",
	}

	if err := run(context.Background(), cfg, []string{"run-once"}); err != nil {
		t.Fatalf("run-once failed: %v", err)
	}
}

func TestRunReportReturnsItemAndDisplayCount(t *testing.T) {
	cfg := config.Config{
		ScheduleInterval: time.Hour,
		StorageDir:       t.TempDir(),
		DedupStorePath:   t.TempDir() + "/seen.json",
		ReportMaxItems:   2,
	}

	result, err := runReport(context.Background(), cfg)
	if err != nil {
		t.Fatalf("runReport failed: %v", err)
	}

	if result.ItemCount != 3 {
		t.Fatalf("expected 3 items, got %d", result.ItemCount)
	}
	if result.DisplayCount != 2 {
		t.Fatalf("expected 2 displayed items, got %d", result.DisplayCount)
	}

	result, err = runReport(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second runReport failed: %v", err)
	}

	if result.ItemCount != 0 {
		t.Fatalf("expected 0 items on second run, got %d", result.ItemCount)
	}
	if result.DisplayCount != 0 {
		t.Fatalf("expected 0 displayed items on second run, got %d", result.DisplayCount)
	}
}
