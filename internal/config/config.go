// Package config 负责读取运行时配置。
package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultScheduleInterval = time.Hour
const defaultStorageDir = "data/reports"
const defaultHTTPAddr = ":8080"
const defaultDedupStorePath = "data/dedup/seen.json"

// Config 保存信息汇总 Agent 的运行配置。
type Config struct {
	RSSURL           string
	RSSURLs          []string
	AIEndpoint       string
	AIAPIKey         string
	AIModel          string
	WebhookURL       string
	ScheduleInterval time.Duration
	StorageDir       string
	HTTPAddr         string
	DedupStorePath   string
	SendEmptyReport  bool
}

// Load 先读取 JSON 配置文件，再使用环境变量覆盖。
func Load() (Config, error) {
	cfg := defaultConfig()
	path := os.Getenv("INFOHUB_CONFIG_PATH")
	if path != "" {
		fileConfig, err := loadFromFile(path)
		if err != nil {
			return Config{}, err
		}

		cfg = mergeConfig(cfg, fileConfig)
	}

	return applyEnv(cfg), nil
}

// LoadFromEnv 从环境变量加载配置。
func LoadFromEnv() Config {
	return applyEnv(defaultConfig())
}

// UseRSS 判断是否启用真实 RSS 数据源。
func (c Config) UseRSS() bool {
	return len(c.RSSURLs) > 0
}

// UseRealAI 判断是否启用真实 AI 客户端。
func (c Config) UseRealAI() bool {
	return c.AIEndpoint != "" && c.AIAPIKey != "" && c.AIModel != ""
}

// UseWebhook 判断是否启用 Webhook 推送。
func (c Config) UseWebhook() bool {
	return c.WebhookURL != ""
}

func defaultConfig() Config {
	return Config{
		ScheduleInterval: defaultScheduleInterval,
		StorageDir:       defaultStorageDir,
		HTTPAddr:         defaultHTTPAddr,
		DedupStorePath:   defaultDedupStorePath,
	}
}

func loadFromFile(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var file fileConfig
	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})
	if err := json.Unmarshal(content, &file); err != nil {
		return Config{}, err
	}

	return file.toConfig(), nil
}

func mergeConfig(base, override Config) Config {
	if override.RSSURL != "" {
		base.RSSURL = override.RSSURL
	}
	if len(override.RSSURLs) > 0 {
		base.RSSURLs = override.RSSURLs
	}
	if override.AIEndpoint != "" {
		base.AIEndpoint = override.AIEndpoint
	}
	if override.AIAPIKey != "" {
		base.AIAPIKey = override.AIAPIKey
	}
	if override.AIModel != "" {
		base.AIModel = override.AIModel
	}
	if override.WebhookURL != "" {
		base.WebhookURL = override.WebhookURL
	}
	if override.ScheduleInterval != 0 {
		base.ScheduleInterval = override.ScheduleInterval
	}
	if override.StorageDir != "" {
		base.StorageDir = override.StorageDir
	}
	if override.HTTPAddr != "" {
		base.HTTPAddr = override.HTTPAddr
	}
	if override.DedupStorePath != "" {
		base.DedupStorePath = override.DedupStorePath
	}
	if override.SendEmptyReport {
		base.SendEmptyReport = true
	}

	return base
}

func applyEnv(cfg Config) Config {
	if value := os.Getenv("INFOHUB_RSS_URL"); value != "" {
		cfg.RSSURL = value
	}

	cfg.RSSURLs = readList("INFOHUB_RSS_URLS", firstNonEmpty(cfg.RSSURL, strings.Join(cfg.RSSURLs, ",")))
	cfg.AIEndpoint = readString("INFOHUB_AI_ENDPOINT", cfg.AIEndpoint)
	cfg.AIAPIKey = readString("INFOHUB_AI_API_KEY", cfg.AIAPIKey)
	cfg.AIModel = readString("INFOHUB_AI_MODEL", cfg.AIModel)
	cfg.WebhookURL = readString("INFOHUB_WEBHOOK_URL", cfg.WebhookURL)
	cfg.ScheduleInterval = readDuration("INFOHUB_SCHEDULE_INTERVAL_SECONDS", cfg.ScheduleInterval)
	cfg.StorageDir = readString("INFOHUB_STORAGE_DIR", cfg.StorageDir)
	cfg.HTTPAddr = readString("INFOHUB_HTTP_ADDR", cfg.HTTPAddr)
	cfg.DedupStorePath = readString("INFOHUB_DEDUP_STORE_PATH", cfg.DedupStorePath)
	cfg.SendEmptyReport = readBool("INFOHUB_SEND_EMPTY_REPORT", cfg.SendEmptyReport)

	return cfg
}

type fileConfig struct {
	RSS struct {
		URL  string   `json:"url"`
		URLs []string `json:"urls"`
	} `json:"rss"`
	AI struct {
		Endpoint string `json:"endpoint"`
		APIKey   string `json:"api_key"`
		Model    string `json:"model"`
	} `json:"ai"`
	Webhook struct {
		URL             string `json:"url"`
		SendEmptyReport bool   `json:"send_empty_report"`
	} `json:"webhook"`
	Storage struct {
		Dir string `json:"dir"`
	} `json:"storage"`
	Dedup struct {
		StorePath string `json:"store_path"`
	} `json:"dedup"`
	HTTP struct {
		Addr string `json:"addr"`
	} `json:"http"`
	Scheduler struct {
		IntervalSeconds int `json:"interval_seconds"`
	} `json:"scheduler"`
}

func (f fileConfig) toConfig() Config {
	cfg := Config{
		RSSURL:          f.RSS.URL,
		RSSURLs:         f.RSS.URLs,
		AIEndpoint:      f.AI.Endpoint,
		AIAPIKey:        f.AI.APIKey,
		AIModel:         f.AI.Model,
		WebhookURL:      f.Webhook.URL,
		SendEmptyReport: f.Webhook.SendEmptyReport,
		StorageDir:      f.Storage.Dir,
		HTTPAddr:        f.HTTP.Addr,
		DedupStorePath:  f.Dedup.StorePath,
	}

	if f.Scheduler.IntervalSeconds > 0 {
		cfg.ScheduleInterval = time.Duration(f.Scheduler.IntervalSeconds) * time.Second
	}

	if len(cfg.RSSURLs) == 0 && cfg.RSSURL != "" {
		cfg.RSSURLs = []string{cfg.RSSURL}
	}

	return cfg
}

func readDuration(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func readString(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

func readList(name, fallback string) []string {
	values := splitList(os.Getenv(name))
	if len(values) > 0 {
		return values
	}

	return splitList(fallback)
}

func readBool(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}

	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
