// Package config 负责读取运行时配置。
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultScheduleInterval = time.Hour
const defaultStorageDir = "data/reports"
const defaultHTTPAddr = ":8080"

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
}

// LoadFromEnv 从环境变量加载配置。
func LoadFromEnv() Config {
	rssURL := os.Getenv("INFOHUB_RSS_URL")
	return Config{
		RSSURL:           rssURL,
		RSSURLs:          readList("INFOHUB_RSS_URLS", rssURL),
		AIEndpoint:       os.Getenv("INFOHUB_AI_ENDPOINT"),
		AIAPIKey:         os.Getenv("INFOHUB_AI_API_KEY"),
		AIModel:          os.Getenv("INFOHUB_AI_MODEL"),
		WebhookURL:       os.Getenv("INFOHUB_WEBHOOK_URL"),
		ScheduleInterval: readDuration("INFOHUB_SCHEDULE_INTERVAL_SECONDS", defaultScheduleInterval),
		StorageDir:       readString("INFOHUB_STORAGE_DIR", defaultStorageDir),
		HTTPAddr:         readString("INFOHUB_HTTP_ADDR", defaultHTTPAddr),
	}
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
