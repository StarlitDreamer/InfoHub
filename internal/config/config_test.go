package config

import (
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("INFOHUB_RSS_URL", "https://example.com/rss.xml")
	t.Setenv("INFOHUB_AI_ENDPOINT", "https://api.example.com/v1/chat/completions")
	t.Setenv("INFOHUB_AI_API_KEY", "test-key")
	t.Setenv("INFOHUB_AI_MODEL", "test-model")
	t.Setenv("INFOHUB_WEBHOOK_URL", "https://example.com/webhook")
	t.Setenv("INFOHUB_SCHEDULE_INTERVAL_SECONDS", "120")
	t.Setenv("INFOHUB_STORAGE_DIR", "tmp/reports")

	cfg := LoadFromEnv()

	if !cfg.UseRSS() {
		t.Fatal("期望启用 RSS 数据源")
	}

	if !cfg.UseRealAI() {
		t.Fatal("期望启用真实 AI 客户端")
	}

	if !cfg.UseWebhook() {
		t.Fatal("期望启用 Webhook 推送")
	}

	if cfg.ScheduleInterval != 120*time.Second {
		t.Fatalf("期望调度间隔为 120 秒，实际为 %s", cfg.ScheduleInterval)
	}

	if cfg.StorageDir != "tmp/reports" {
		t.Fatalf("期望存储目录为 tmp/reports，实际为 %s", cfg.StorageDir)
	}
}

func TestLoadFromEnvUsesFallbackInterval(t *testing.T) {
	t.Setenv("INFOHUB_SCHEDULE_INTERVAL_SECONDS", "invalid")

	cfg := LoadFromEnv()

	if cfg.ScheduleInterval != defaultScheduleInterval {
		t.Fatalf("期望使用默认调度间隔，实际为 %s", cfg.ScheduleInterval)
	}

	if cfg.StorageDir != defaultStorageDir {
		t.Fatalf("期望使用默认存储目录，实际为 %s", cfg.StorageDir)
	}
}
