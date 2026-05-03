package main

import (
	"fmt"
	"log"

	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/config"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/delivery"
	"InfoHub-agent/internal/service"
)

func main() {
	// 根据运行配置选择真实链路或本地演示链路。
	cfg := config.LoadFromEnv()
	pipeline := service.NewPipeline(newCrawler(cfg), newAIProcessor(cfg))

	items, err := pipeline.Run()
	if err != nil {
		log.Fatal(err)
	}

	report := delivery.RenderMarkdown(items)
	fmt.Print(report)

	if cfg.UseWebhook() {
		if err := delivery.NewWebhookSender(cfg.WebhookURL, nil).Send(report); err != nil {
			log.Fatal(err)
		}
	}
}

func newCrawler(cfg config.Config) crawler.Crawler {
	if cfg.UseRSS() {
		return crawler.NewRSSCrawler(cfg.RSSURL, nil)
	}

	return crawler.NewDemoCrawler()
}

func newAIProcessor(cfg config.Config) ai.Processor {
	if cfg.UseRealAI() {
		return ai.NewHTTPClient(cfg.AIEndpoint, cfg.AIAPIKey, cfg.AIModel, nil)
	}

	return ai.NewMockProcessor()
}
