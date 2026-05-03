package config

import (
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("INFOHUB_RSS_URL", "https://example.com/rss.xml")
	t.Setenv("INFOHUB_RSS_URLS", "https://example.com/a.xml, https://example.com/b.xml")
	t.Setenv("INFOHUB_AI_ENDPOINT", "https://api.example.com/v1/chat/completions")
	t.Setenv("INFOHUB_AI_API_KEY", "test-key")
	t.Setenv("INFOHUB_AI_MODEL", "test-model")
	t.Setenv("INFOHUB_WEBHOOK_URL", "https://example.com/webhook")
	t.Setenv("INFOHUB_SCHEDULE_INTERVAL_SECONDS", "120")
	t.Setenv("INFOHUB_STORAGE_DIR", "tmp/reports")
	t.Setenv("INFOHUB_HTTP_ADDR", ":9090")
	t.Setenv("INFOHUB_DEDUP_STORE_PATH", "tmp/dedup/seen.json")
	t.Setenv("INFOHUB_SEND_EMPTY_REPORT", "true")

	cfg := LoadFromEnv()

	if !cfg.UseRSS() {
		t.Fatal("期望启用 RSS 数据源")
	}

	if len(cfg.RSSURLs) != 2 {
		t.Fatalf("期望读取 2 个 RSS 数据源，实际为 %d", len(cfg.RSSURLs))
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

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("期望 HTTP 地址为 :9090，实际为 %s", cfg.HTTPAddr)
	}

	if cfg.DedupStorePath != "tmp/dedup/seen.json" {
		t.Fatalf("期望去重状态路径为 tmp/dedup/seen.json，实际为 %s", cfg.DedupStorePath)
	}

	if !cfg.SendEmptyReport {
		t.Fatal("期望启用空日报推送")
	}
}

func TestLoadFromEnvFallsBackToSingleRSSURL(t *testing.T) {
	t.Setenv("INFOHUB_RSS_URL", "https://example.com/rss.xml")

	cfg := LoadFromEnv()

	if len(cfg.RSSURLs) != 1 || cfg.RSSURLs[0] != "https://example.com/rss.xml" {
		t.Fatalf("期望兼容单个 RSS URL，实际为 %+v", cfg.RSSURLs)
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

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("期望使用默认 HTTP 地址，实际为 %s", cfg.HTTPAddr)
	}

	if cfg.DedupStorePath != defaultDedupStorePath {
		t.Fatalf("期望使用默认去重状态路径，实际为 %s", cfg.DedupStorePath)
	}

	if cfg.SendEmptyReport {
		t.Fatal("默认不应推送空日报")
	}
}
