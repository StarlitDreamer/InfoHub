// Package config 负责读取运行时配置。
package config

import (
	"os"
	"strconv"
	"time"
)

const defaultScheduleInterval = time.Hour

// Config 保存信息汇总 Agent 的运行配置。
type Config struct {
	RSSURL           string
	AIEndpoint       string
	AIAPIKey         string
	AIModel          string
	WebhookURL       string
	ScheduleInterval time.Duration
}

// LoadFromEnv 从环境变量加载配置。
func LoadFromEnv() Config {
	return Config{
		RSSURL:           os.Getenv("INFOHUB_RSS_URL"),
		AIEndpoint:       os.Getenv("INFOHUB_AI_ENDPOINT"),
		AIAPIKey:         os.Getenv("INFOHUB_AI_API_KEY"),
		AIModel:          os.Getenv("INFOHUB_AI_MODEL"),
		WebhookURL:       os.Getenv("INFOHUB_WEBHOOK_URL"),
		ScheduleInterval: readDuration("INFOHUB_SCHEDULE_INTERVAL_SECONDS", defaultScheduleInterval),
	}
}

// UseRSS 判断是否启用真实 RSS 数据源。
func (c Config) UseRSS() bool {
	return c.RSSURL != ""
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
