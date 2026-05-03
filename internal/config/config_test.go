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
	t.Setenv("INFOHUB_RSS_MAX_ITEMS_PER_FEED", "20")
	t.Setenv("INFOHUB_RSS_RECENT_WITHIN_HOURS", "72")
	t.Setenv("INFOHUB_REPORT_MAX_ITEMS", "12")
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
	t.Setenv("INFOHUB_MYSQL_DSN", "user:pass@tcp(localhost:3306)/infohub?parseTime=true")
	t.Setenv("INFOHUB_MYSQL_TABLE", "daily_reports")

	cfg := LoadFromEnv()

	if !cfg.UseRSS() {
		t.Fatal("expected RSS to be enabled")
	}
	if len(cfg.RSSURLs) != 2 {
		t.Fatalf("expected 2 RSS urls, got %d", len(cfg.RSSURLs))
	}
	if cfg.RSSMaxItems != 20 || cfg.RSSRecentWithin != 72*time.Hour || cfg.ReportMaxItems != 12 {
		t.Fatalf("unexpected RSS trimming config: %+v", cfg)
	}
	if !cfg.UseRealAI() {
		t.Fatal("expected real AI client to be enabled")
	}
	if !cfg.UseWebhook() {
		t.Fatal("expected webhook to be enabled")
	}
	if cfg.ScheduleInterval != 120*time.Second {
		t.Fatalf("expected 120s schedule interval, got %s", cfg.ScheduleInterval)
	}
	if cfg.StorageDir != "tmp/reports" {
		t.Fatalf("expected storage dir tmp/reports, got %s", cfg.StorageDir)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected http addr :9090, got %s", cfg.HTTPAddr)
	}
	if cfg.DedupStorePath != "tmp/dedup/seen.json" {
		t.Fatalf("expected dedup path tmp/dedup/seen.json, got %s", cfg.DedupStorePath)
	}
	if !cfg.SendEmptyReport {
		t.Fatal("expected empty report sending to be enabled")
	}
	if cfg.AuthToken != "env-token" {
		t.Fatalf("expected auth token env-token, got %s", cfg.AuthToken)
	}
	if cfg.RedisAddr != "localhost:6379" || cfg.RedisPassword != "secret" || cfg.RedisDB != 2 || cfg.RedisDedupKey != "custom:dedup" {
		t.Fatalf("unexpected Redis config: %+v", cfg)
	}
	if !cfg.UseMySQL() {
		t.Fatal("expected MySQL storage to be enabled")
	}
	if cfg.MySQLDSN != "user:pass@tcp(localhost:3306)/infohub?parseTime=true" || cfg.MySQLTable != "daily_reports" {
		t.Fatalf("unexpected MySQL config: %+v", cfg)
	}
}

func TestLoadFromEnvFallsBackToSingleRSSURL(t *testing.T) {
	t.Setenv("INFOHUB_RSS_URL", "https://example.com/rss.xml")

	cfg := LoadFromEnv()

	if len(cfg.RSSURLs) != 1 || cfg.RSSURLs[0] != "https://example.com/rss.xml" {
		t.Fatalf("expected single RSS URL fallback, got %+v", cfg.RSSURLs)
	}
	if len(cfg.SourcesOrDefault()) != 1 || cfg.SourcesOrDefault()[0].Kind != "rss" {
		t.Fatalf("expected rss source fallback, got %+v", cfg.SourcesOrDefault())
	}
}

func TestLoadFromEnvUsesFallbackInterval(t *testing.T) {
	t.Setenv("INFOHUB_SCHEDULE_INTERVAL_SECONDS", "invalid")

	cfg := LoadFromEnv()

	if cfg.ScheduleInterval != defaultScheduleInterval {
		t.Fatalf("expected default interval, got %s", cfg.ScheduleInterval)
	}
	if cfg.StorageDir != defaultStorageDir {
		t.Fatalf("expected default storage dir, got %s", cfg.StorageDir)
	}
	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("expected default http addr, got %s", cfg.HTTPAddr)
	}
	if cfg.DedupStorePath != defaultDedupStorePath {
		t.Fatalf("expected default dedup path, got %s", cfg.DedupStorePath)
	}
	if cfg.MySQLTable != defaultMySQLTable {
		t.Fatalf("expected default MySQL table, got %s", cfg.MySQLTable)
	}
	if cfg.SendEmptyReport {
		t.Fatal("did not expect empty report sending by default")
	}
}

func TestLoadFromJSONFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "sources": [{"name": "primary", "kind": "rss", "location": "https://example.com/source.xml"}],
  "rss": {"urls": ["https://example.com/a.xml", "https://example.com/b.xml"], "max_items_per_feed": 10, "recent_within_hours": 48},
  "report": {"max_items": 15},
  "ai": {"endpoint": "https://api.example.com/v1/chat/completions", "api_key": "file-key", "model": "file-model"},
  "webhook": {"url": "https://example.com/webhook", "send_empty_report": true},
  "storage": {"dir": "file/reports"},
  "dedup": {"store_path": "file/dedup/seen.json"},
  "http": {"addr": ":7070"},
  "auth": {"token": "file-token"},
  "redis": {"addr": "redis:6379", "password": "file-secret", "db": 1, "dedup_key": "file:dedup"},
  "mysql": {"dsn": "user:pass@tcp(mysql:3306)/infohub?parseTime=true", "table": "report_records"},
  "scheduler": {"interval_seconds": 300}
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	t.Setenv("INFOHUB_CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.RSSURLs) != 2 {
		t.Fatalf("expected 2 RSS urls, got %d", len(cfg.RSSURLs))
	}
	if len(cfg.Sources) != 1 || cfg.Sources[0].Name != "primary" {
		t.Fatalf("expected explicit sources config, got %+v", cfg.Sources)
	}
	if cfg.RSSMaxItems != 10 || cfg.RSSRecentWithin != 48*time.Hour || cfg.ReportMaxItems != 15 {
		t.Fatalf("unexpected RSS trimming config: %+v", cfg)
	}
	if cfg.AIAPIKey != "file-key" {
		t.Fatalf("expected file AI key, got %s", cfg.AIAPIKey)
	}
	if cfg.ScheduleInterval != 300*time.Second {
		t.Fatalf("expected 300s interval, got %s", cfg.ScheduleInterval)
	}
	if cfg.AuthToken != "file-token" {
		t.Fatalf("expected file auth token, got %s", cfg.AuthToken)
	}
	if cfg.RedisAddr != "redis:6379" || cfg.RedisPassword != "file-secret" || cfg.RedisDB != 1 || cfg.RedisDedupKey != "file:dedup" {
		t.Fatalf("unexpected Redis config: %+v", cfg)
	}
	if cfg.MySQLDSN != "user:pass@tcp(mysql:3306)/infohub?parseTime=true" || cfg.MySQLTable != "report_records" {
		t.Fatalf("unexpected MySQL config: %+v", cfg)
	}
}

func TestSourcesOrDefaultPrefersExplicitSources(t *testing.T) {
	cfg := Config{
		Sources: []SourceConfig{
			{Name: "demo-source", Kind: "demo", Location: "in-memory"},
		},
		RSSURLs: []string{"https://example.com/a.xml"},
	}

	sources := cfg.SourcesOrDefault()

	if len(sources) != 1 || sources[0].Kind != "demo" {
		t.Fatalf("expected explicit sources to win, got %+v", sources)
	}
}

func TestSourcesOrDefaultFallsBackToDemo(t *testing.T) {
	sources := (Config{}).SourcesOrDefault()

	if len(sources) != 1 || sources[0].Kind != "demo" {
		t.Fatalf("expected demo fallback, got %+v", sources)
	}
}

func TestLoadEnvOverridesJSONFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "rss": {"urls": ["https://example.com/file.xml"]},
  "ai": {"endpoint": "https://api.example.com/file", "api_key": "file-key", "model": "file-model"},
  "http": {"addr": ":7070"},
  "mysql": {"dsn": "user:pass@tcp(mysql:3306)/infohub?parseTime=true", "table": "file_reports"}
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	t.Setenv("INFOHUB_CONFIG_PATH", path)
	t.Setenv("INFOHUB_AI_MODEL", "env-model")
	t.Setenv("INFOHUB_HTTP_ADDR", ":9090")
	t.Setenv("INFOHUB_MYSQL_TABLE", "env_reports")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.AIModel != "env-model" {
		t.Fatalf("expected env AI model, got %s", cfg.AIModel)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected env http addr, got %s", cfg.HTTPAddr)
	}
	if cfg.MySQLTable != "env_reports" {
		t.Fatalf("expected env MySQL table, got %s", cfg.MySQLTable)
	}
}

func TestLoadJSONFileWithBOM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"http":{"addr":":6060"}}`)...)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	t.Setenv("INFOHUB_CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load BOM config: %v", err)
	}

	if cfg.HTTPAddr != ":6060" {
		t.Fatalf("expected http addr :6060, got %s", cfg.HTTPAddr)
	}
}
