package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/config"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/delivery"
	"InfoHub-agent/internal/processor"
	"InfoHub-agent/internal/repository"
	"InfoHub-agent/internal/scheduler"
	"InfoHub-agent/internal/server"
	"InfoHub-agent/internal/service"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

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
		_, err := runReport(ctx, cfg)
		return err
	case "schedule":
		return runSchedule(ctx, cfg)
	case "serve":
		return runServer(cfg)
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
}

func runReport(ctx context.Context, cfg config.Config) (server.ReportResult, error) {
	repo, closeRepo, err := newReportRepository(cfg)
	if err != nil {
		return server.ReportResult{}, err
	}
	defer closeRepo()

	return runReportWithRepository(ctx, cfg, repo)
}

func runReportWithRepository(ctx context.Context, cfg config.Config, repo repository.ReportRepository) (server.ReportResult, error) {
	pipeline := service.NewPipeline(
		newCrawler(cfg),
		newAIProcessor(cfg),
	).WithDedupStore(newDedupStore(cfg))
	options := service.AgentOptions{
		SendEmptyReport: cfg.SendEmptyReport,
		ReportMaxItems:  cfg.ReportMaxItems,
		Now:             timeNow,
	}
	if cfg.UseWebhook() {
		options.WebhookSender = delivery.NewWebhookSender(cfg.WebhookURL, nil)
	}

	agent := service.NewAgent(pipeline, repo, options)
	result, err := agent.Run(ctx)
	if err != nil {
		return server.ReportResult{}, err
	}

	fmt.Print(result.Markdown)
	return server.ReportResult{
		ItemCount:    result.ItemCount,
		DisplayCount: result.DisplayCount,
	}, nil
}

func runSchedule(ctx context.Context, cfg config.Config) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	job := func(context.Context) error {
		_, err := runReport(ctx, cfg)
		return err
	}
	task := scheduler.New(cfg.ScheduleInterval, job)

	if err := task.RunOnce(ctx); err != nil {
		return err
	}

	task.Start(ctx, func(err error) {
		log.Printf("schedule run failed: %v", err)
	})

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func runServer(cfg config.Config) error {
	repo, closeRepo, err := newReportRepository(cfg)
	if err != nil {
		return err
	}
	defer closeRepo()

	router := server.NewRouter(repo, func(ctx context.Context) (server.ReportResult, error) {
		return runReportWithRepository(ctx, cfg, repo)
	}, server.Options{AuthToken: cfg.AuthToken})

	return router.Run(cfg.HTTPAddr)
}

var timeNow = func() time.Time {
	return time.Now()
}

func newCrawler(cfg config.Config) crawler.Crawler {
	if cfg.UseRSS() {
		crawlers := make([]crawler.Crawler, 0, len(cfg.RSSURLs))
		for _, url := range cfg.RSSURLs {
			crawlers = append(crawlers, crawler.NewRSSCrawler(url, nil, crawler.RSSOptions{
				MaxItems:     cfg.RSSMaxItems,
				RecentWithin: cfg.RSSRecentWithin,
			}))
		}

		if len(crawlers) == 1 {
			return crawlers[0]
		}

		return crawler.NewMultiCrawler(crawlers)
	}

	return crawler.NewDemoCrawler()
}

func newAIProcessor(cfg config.Config) ai.Processor {
	if cfg.UseRealAI() {
		return ai.NewHTTPClient(cfg.AIEndpoint, cfg.AIAPIKey, cfg.AIModel, nil)
	}

	return ai.NewMockProcessor()
}

func newDedupStore(cfg config.Config) processor.DedupStore {
	if cfg.RedisAddr != "" {
		client := processor.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		return processor.NewRedisDedupStore(client, cfg.RedisDedupKey)
	}

	return processor.NewFileDedupStore(cfg.DedupStorePath)
}

func newReportRepository(cfg config.Config) (repository.ReportRepository, func() error, error) {
	if cfg.UseMySQL() {
		db, err := sql.Open("mysql", cfg.MySQLDSN)
		if err != nil {
			return nil, nil, err
		}

		repo, err := repository.NewMySQLReportRepository(db, cfg.MySQLTable)
		if err != nil {
			_ = db.Close()
			return nil, nil, err
		}

		return repo, repo.Close, nil
	}

	repo := repository.NewFileReportRepository(cfg.StorageDir)
	return repo, func() error { return nil }, nil
}
