package config

import (
	"os"
	"path/filepath"
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
	t.Setenv("INFOHUB_AUTH_TOKEN", "env-token")
	t.Setenv("INFOHUB_REDIS_ADDR", "localhost:6379")
	t.Setenv("INFOHUB_REDIS_PASSWORD", "secret")
	t.Setenv("INFOHUB_REDIS_DB", "2")
	t.Setenv("INFOHUB_REDIS_DEDUP_KEY", "custom:dedup")

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

	if cfg.AuthToken != "env-token" {
		t.Fatalf("期望读取鉴权 token，实际为 %s", cfg.AuthToken)
	}

	if cfg.RedisAddr != "localhost:6379" || cfg.RedisPassword != "secret" || cfg.RedisDB != 2 || cfg.RedisDedupKey != "custom:dedup" {
		t.Fatalf("Redis 配置不符合预期：%+v", cfg)
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

func TestLoadFromJSONFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "rss": {"urls": ["https://example.com/a.xml", "https://example.com/b.xml"]},
  "ai": {"endpoint": "https://api.example.com/v1/chat/completions", "api_key": "file-key", "model": "file-model"},
  "webhook": {"url": "https://example.com/webhook", "send_empty_report": true},
  "storage": {"dir": "file/reports"},
  "dedup": {"store_path": "file/dedup/seen.json"},
  "http": {"addr": ":7070"},
  "auth": {"token": "file-token"},
  "redis": {"addr": "redis:6379", "password": "file-secret", "db": 1, "dedup_key": "file:dedup"},
  "scheduler": {"interval_seconds": 300}
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试配置失败：%v", err)
	}
	t.Setenv("INFOHUB_CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("读取 JSON 配置失败：%v", err)
	}

	if len(cfg.RSSURLs) != 2 {
		t.Fatalf("期望读取 2 个 RSS 源，实际为 %d", len(cfg.RSSURLs))
	}

	if cfg.AIAPIKey != "file-key" {
		t.Fatalf("期望读取文件中的 AI key，实际为 %s", cfg.AIAPIKey)
	}

	if cfg.ScheduleInterval != 300*time.Second {
		t.Fatalf("期望调度间隔为 300 秒，实际为 %s", cfg.ScheduleInterval)
	}

	if cfg.AuthToken != "file-token" {
		t.Fatalf("期望读取文件中的鉴权 token，实际为 %s", cfg.AuthToken)
	}

	if cfg.RedisAddr != "redis:6379" || cfg.RedisPassword != "file-secret" || cfg.RedisDB != 1 || cfg.RedisDedupKey != "file:dedup" {
		t.Fatalf("期望读取文件中的 Redis 配置，实际为 %+v", cfg)
	}
}

func TestLoadEnvOverridesJSONFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "rss": {"urls": ["https://example.com/file.xml"]},
  "ai": {"endpoint": "https://api.example.com/file", "api_key": "file-key", "model": "file-model"},
  "http": {"addr": ":7070"}
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试配置失败：%v", err)
	}
	t.Setenv("INFOHUB_CONFIG_PATH", path)
	t.Setenv("INFOHUB_AI_MODEL", "env-model")
	t.Setenv("INFOHUB_HTTP_ADDR", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("读取配置失败：%v", err)
	}

	if cfg.AIModel != "env-model" {
		t.Fatalf("期望环境变量覆盖 AI model，实际为 %s", cfg.AIModel)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("期望环境变量覆盖 HTTP 地址，实际为 %s", cfg.HTTPAddr)
	}
}

func TestLoadJSONFileWithBOM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"http":{"addr":":6060"}}`)...)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("写入测试配置失败：%v", err)
	}
	t.Setenv("INFOHUB_CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("读取带 BOM 的 JSON 配置失败：%v", err)
	}

	if cfg.HTTPAddr != ":6060" {
		t.Fatalf("期望读取 HTTP 地址 :6060，实际为 %s", cfg.HTTPAddr)
	}
}
