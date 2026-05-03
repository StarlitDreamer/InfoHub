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
const defaultRedisDedupKey = "infohub:dedup:seen"
const defaultMySQLTable = "reports"

// Config 保存信息汇总 Agent 的运行配置。
type Config struct {
	Sources             []SourceConfig
	RSSURL              string
	RSSURLs             []string
	RSSMaxItems         int
	RSSRecentWithin     time.Duration
	ReportMaxItems      int
	ReportGroupBySource bool
	AIEndpoint          string
	AIAPIKey            string
	AIModel             string
	WebhookURL          string
	ScheduleInterval    time.Duration
	StorageDir          string
	HTTPAddr            string
	DedupStorePath      string
	SendEmptyReport     bool
	AuthToken           string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	RedisDedupKey       string
	MySQLDSN            string
	MySQLTable          string
}

// SourceConfig 定义一个可执行的数据源配置。
type SourceConfig struct {
	Enabled         bool
	Name            string
	Kind            string
	Location        string
	Priority        int
	IncludeInReport bool
	TimeoutSeconds  int
	Headers         map[string]string
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

// SourcesOrDefault 返回显式 source 配置；如果未提供，则兼容旧 RSS 配置。
func (c Config) SourcesOrDefault() []SourceConfig {
	if len(c.Sources) > 0 {
		result := make([]SourceConfig, 0, len(c.Sources))
		for _, source := range c.Sources {
			if source.Kind == "" || source.Location == "" {
				continue
			}
			result = append(result, source)
		}
		if len(result) > 0 {
			return result
		}
	}

	if len(c.RSSURLs) > 0 {
		result := make([]SourceConfig, 0, len(c.RSSURLs))
		for _, url := range c.RSSURLs {
			result = append(result, SourceConfig{
				Enabled:         true,
				Name:            url,
				Kind:            "rss",
				Location:        url,
				IncludeInReport: true,
			})
		}
		return result
	}

	return []SourceConfig{{
		Enabled:         true,
		Name:            "demo",
		Kind:            "demo",
		Location:        "in-memory",
		IncludeInReport: true,
	}}
}

// UseRealAI 判断是否启用真实 AI 客户端。
func (c Config) UseRealAI() bool {
	return c.AIEndpoint != "" && c.AIAPIKey != "" && c.AIModel != ""
}

// UseWebhook 判断是否启用 Webhook 推送。
func (c Config) UseWebhook() bool {
	return c.WebhookURL != ""
}

// UseMySQL 判断是否启用 MySQL 日报存储。
func (c Config) UseMySQL() bool {
	return c.MySQLDSN != ""
}

func defaultConfig() Config {
	return Config{
		ScheduleInterval: defaultScheduleInterval,
		StorageDir:       defaultStorageDir,
		HTTPAddr:         defaultHTTPAddr,
		DedupStorePath:   defaultDedupStorePath,
		RedisDedupKey:    defaultRedisDedupKey,
		MySQLTable:       defaultMySQLTable,
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
	if len(override.Sources) > 0 {
		base.Sources = override.Sources
	}
	if override.RSSURL != "" {
		base.RSSURL = override.RSSURL
	}
	if len(override.RSSURLs) > 0 {
		base.RSSURLs = override.RSSURLs
	}
	if override.RSSMaxItems > 0 {
		base.RSSMaxItems = override.RSSMaxItems
	}
	if override.RSSRecentWithin > 0 {
		base.RSSRecentWithin = override.RSSRecentWithin
	}
	if override.ReportMaxItems > 0 {
		base.ReportMaxItems = override.ReportMaxItems
	}
	if override.ReportGroupBySource {
		base.ReportGroupBySource = true
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
	if override.AuthToken != "" {
		base.AuthToken = override.AuthToken
	}
	if override.RedisAddr != "" {
		base.RedisAddr = override.RedisAddr
	}
	if override.RedisPassword != "" {
		base.RedisPassword = override.RedisPassword
	}
	if override.RedisDB != 0 {
		base.RedisDB = override.RedisDB
	}
	if override.RedisDedupKey != "" {
		base.RedisDedupKey = override.RedisDedupKey
	}
	if override.MySQLDSN != "" {
		base.MySQLDSN = override.MySQLDSN
	}
	if override.MySQLTable != "" {
		base.MySQLTable = override.MySQLTable
	}

	return base
}

func applyEnv(cfg Config) Config {
	if value := os.Getenv("INFOHUB_RSS_URL"); value != "" {
		cfg.RSSURL = value
	}

	cfg.RSSURLs = readList("INFOHUB_RSS_URLS", firstNonEmpty(cfg.RSSURL, strings.Join(cfg.RSSURLs, ",")))
	cfg.RSSMaxItems = readInt("INFOHUB_RSS_MAX_ITEMS_PER_FEED", cfg.RSSMaxItems)
	cfg.RSSRecentWithin = readHours("INFOHUB_RSS_RECENT_WITHIN_HOURS", cfg.RSSRecentWithin)
	cfg.ReportMaxItems = readInt("INFOHUB_REPORT_MAX_ITEMS", cfg.ReportMaxItems)
	cfg.AIEndpoint = readString("INFOHUB_AI_ENDPOINT", cfg.AIEndpoint)
	cfg.AIAPIKey = readString("INFOHUB_AI_API_KEY", cfg.AIAPIKey)
	cfg.AIModel = readString("INFOHUB_AI_MODEL", cfg.AIModel)
	cfg.WebhookURL = readString("INFOHUB_WEBHOOK_URL", cfg.WebhookURL)
	cfg.ScheduleInterval = readDuration("INFOHUB_SCHEDULE_INTERVAL_SECONDS", cfg.ScheduleInterval)
	cfg.StorageDir = readString("INFOHUB_STORAGE_DIR", cfg.StorageDir)
	cfg.HTTPAddr = readString("INFOHUB_HTTP_ADDR", cfg.HTTPAddr)
	cfg.DedupStorePath = readString("INFOHUB_DEDUP_STORE_PATH", cfg.DedupStorePath)
	cfg.SendEmptyReport = readBool("INFOHUB_SEND_EMPTY_REPORT", cfg.SendEmptyReport)
	cfg.AuthToken = readString("INFOHUB_AUTH_TOKEN", cfg.AuthToken)
	cfg.RedisAddr = readString("INFOHUB_REDIS_ADDR", cfg.RedisAddr)
	cfg.RedisPassword = readString("INFOHUB_REDIS_PASSWORD", cfg.RedisPassword)
	cfg.RedisDB = readInt("INFOHUB_REDIS_DB", cfg.RedisDB)
	cfg.RedisDedupKey = readString("INFOHUB_REDIS_DEDUP_KEY", cfg.RedisDedupKey)
	cfg.MySQLDSN = readString("INFOHUB_MYSQL_DSN", cfg.MySQLDSN)
	cfg.MySQLTable = readString("INFOHUB_MYSQL_TABLE", cfg.MySQLTable)

	return cfg
}

type fileConfig struct {
	Sources []sourceFileConfig `json:"sources"`
	RSS     struct {
		URL               string   `json:"url"`
		URLs              []string `json:"urls"`
		MaxItemsPerFeed   int      `json:"max_items_per_feed"`
		RecentWithinHours int      `json:"recent_within_hours"`
	} `json:"rss"`
	Report struct {
		MaxItems      int  `json:"max_items"`
		GroupBySource bool `json:"group_by_source"`
	} `json:"report"`
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
	Auth struct {
		Token string `json:"token"`
	} `json:"auth"`
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
		DedupKey string `json:"dedup_key"`
	} `json:"redis"`
	MySQL struct {
		DSN   string `json:"dsn"`
		Table string `json:"table"`
	} `json:"mysql"`
	Scheduler struct {
		IntervalSeconds int `json:"interval_seconds"`
	} `json:"scheduler"`
}

type sourceFileConfig struct {
	Enabled         *bool             `json:"enabled"`
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Location        string            `json:"location"`
	Priority        int               `json:"priority"`
	IncludeInReport *bool             `json:"include_in_report"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	Headers         map[string]string `json:"headers"`
}

func (f fileConfig) toConfig() Config {
	cfg := Config{
		Sources:             make([]SourceConfig, 0, len(f.Sources)),
		RSSURL:              f.RSS.URL,
		RSSURLs:             f.RSS.URLs,
		RSSMaxItems:         f.RSS.MaxItemsPerFeed,
		AIEndpoint:          f.AI.Endpoint,
		AIAPIKey:            f.AI.APIKey,
		AIModel:             f.AI.Model,
		WebhookURL:          f.Webhook.URL,
		SendEmptyReport:     f.Webhook.SendEmptyReport,
		StorageDir:          f.Storage.Dir,
		HTTPAddr:            f.HTTP.Addr,
		DedupStorePath:      f.Dedup.StorePath,
		AuthToken:           f.Auth.Token,
		RedisAddr:           f.Redis.Addr,
		RedisPassword:       f.Redis.Password,
		RedisDB:             f.Redis.DB,
		RedisDedupKey:       f.Redis.DedupKey,
		MySQLDSN:            f.MySQL.DSN,
		MySQLTable:          firstNonEmpty(f.MySQL.Table, defaultMySQLTable),
		ReportMaxItems:      f.Report.MaxItems,
		ReportGroupBySource: f.Report.GroupBySource,
	}
	for _, source := range f.Sources {
		enabled := true
		if source.Enabled != nil {
			enabled = *source.Enabled
		}
		includeInReport := true
		if source.IncludeInReport != nil {
			includeInReport = *source.IncludeInReport
		}
		cfg.Sources = append(cfg.Sources, SourceConfig{
			Enabled:         enabled,
			Name:            source.Name,
			Kind:            source.Kind,
			Location:        source.Location,
			Priority:        source.Priority,
			IncludeInReport: includeInReport,
			TimeoutSeconds:  source.TimeoutSeconds,
			Headers:         cloneStringMap(source.Headers),
		})
	}

	if f.Scheduler.IntervalSeconds > 0 {
		cfg.ScheduleInterval = time.Duration(f.Scheduler.IntervalSeconds) * time.Second
	}
	if f.RSS.RecentWithinHours > 0 {
		cfg.RSSRecentWithin = time.Duration(f.RSS.RecentWithinHours) * time.Hour
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

func readHours(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	hours, err := strconv.Atoi(value)
	if err != nil || hours <= 0 {
		return fallback
	}

	return time.Duration(hours) * time.Hour
}

func readString(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

func readInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
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

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}

	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}

	return result
}
