package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/config"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/delivery"
	"InfoHub-agent/internal/scheduler"
	"InfoHub-agent/internal/service"
)

func main() {
	cfg := config.LoadFromEnv()
	if err := run(context.Background(), cfg, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg config.Config, args []string) error {
	mode := "run-once"
	if len(args) > 0 {
		mode = args[0]
	}

	switch mode {
	case "run-once":
		return runReport(cfg)
	case "schedule":
		return runSchedule(ctx, cfg)
	default:
		return fmt.Errorf("未知运行模式：%s，可用模式：run-once、schedule", mode)
	}
}

func runReport(cfg config.Config) error {
	// 根据运行配置选择真实链路或本地演示链路。
	pipeline := service.NewPipeline(newCrawler(cfg), newAIProcessor(cfg))
	items, err := pipeline.Run()
	if err != nil {
		return err
	}

	report := delivery.RenderMarkdown(items)
	fmt.Print(report)

	if cfg.UseWebhook() {
		if err := delivery.NewWebhookSender(cfg.WebhookURL, nil).Send(report); err != nil {
			return err
		}
	}

	return nil
}

func runSchedule(ctx context.Context, cfg config.Config) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	job := func(context.Context) error {
		return runReport(cfg)
	}
	task := scheduler.New(cfg.ScheduleInterval, job)

	if err := task.RunOnce(ctx); err != nil {
		return err
	}

	task.Start(ctx, func(err error) {
		log.Printf("定时任务执行失败：%v", err)
	})

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
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
