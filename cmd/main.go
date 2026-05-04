package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
		_, err := runReport(ctx, cfg, "manual")
		return err
	case "schedule":
		return runSchedule(ctx, cfg)
	case "serve":
		return runServer(cfg)
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
}

func runReport(ctx context.Context, cfg config.Config, trigger string) (server.ReportResult, error) {
	repo, closeRepo, err := newReportRepository(cfg)
	if err != nil {
		return server.ReportResult{}, err
	}
	defer closeRepo()

	return runReportWithRepository(ctx, cfg, repo, buildAgentRequest(cfg, trigger))
}

func runReportWithRepository(ctx context.Context, cfg config.Config, repo repository.ReportRepository, request service.AgentRequest) (server.ReportResult, error) {
	pipeline := service.NewPipeline(
		newCrawler(cfg),
		newAIProcessor(cfg),
	).WithDedupStore(newDedupStore(cfg))
	options := service.AgentOptions{
		SendEmptyReport: cfg.SendEmptyReport,
		GroupBySource:   cfg.ReportGroupBySource,
		ReportMaxItems:  cfg.ReportMaxItems,
		Now:             timeNow,
		Senders:         buildSenders(cfg),
	}

	agent := service.NewAgent(pipeline, repo, options)
	result, err := agent.RunWithRequest(ctx, request)
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
		_, err := runReport(ctx, cfg, "schedule")
		return err
	}
	task, err := buildScheduler(cfg, job)
	if err != nil {
		return err
	}

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

	preferenceRepo, closePreferenceRepo, err := newUserPreferenceRepository(cfg)
	if err != nil {
		return err
	}
	defer closePreferenceRepo()
	router := server.NewRouter(repo, func(ctx context.Context, request server.RunReportRequest) (server.ReportResult, error) {
		agentRequest := buildAgentRequest(cfg, "http")
		if request.UserID != "" {
			preference, err := resolveUserPreference(ctx, preferenceRepo, request.UserID)
			if err != nil {
				return server.ReportResult{}, err
			}
			agentRequest.Context.Preference = mergePreference(agentRequest.Context.Preference, preference)
		}
		agentRequest.Context.Preference = mergePreference(agentRequest.Context.Preference, request.Preference.ToUserPreference())
		return runReportWithRepository(ctx, cfg, repo, agentRequest)
	}, server.Options{
		AuthToken:          cfg.AuthToken,
		UserPreferenceRepo: preferenceRepo,
	})

	return router.Run(cfg.HTTPAddr)
}

var timeNow = func() time.Time {
	return time.Now()
}

func newCrawler(cfg config.Config) crawler.Crawler {
	built, err := crawler.BuildFromSources(cfg.SourcesOrDefault(), crawler.FactoryOptions{
		RSSMaxItems:     cfg.RSSMaxItems,
		RSSRecentWithin: cfg.RSSRecentWithin,
	})
	if err != nil {
		log.Printf("build crawler failed, fallback to demo source: %v", err)
		return crawler.NewDemoCrawler()
	}

	return built
}

func newAIProcessor(cfg config.Config) ai.Processor {
	if cfg.UseRealAI() {
		return ai.NewHTTPClient(cfg.AIEndpoint, cfg.AIAPIKey, cfg.AIModel, nil)
	}

	return ai.NewMockProcessor()
}

func buildAgentRequest(cfg config.Config, trigger string) service.AgentRequest {
	configuredSources := cfg.SourcesOrDefault()
	sources := make([]service.Source, 0, len(configuredSources))
	for _, source := range configuredSources {
		sources = append(sources, service.Source{
			Name:            source.Name,
			Kind:            source.Kind,
			Location:        source.Location,
			Priority:        source.Priority,
			IncludeInReport: source.IncludeInReport,
		})
	}

	return service.AgentRequest{
		Context: service.ExecutionContext{
			Trigger: trigger,
			Sources: sources,
			Preference: service.UserPreference{
				Tags:     append([]string(nil), cfg.PreferenceTags...),
				Sources:  append([]string(nil), cfg.PreferenceSources...),
				Keywords: append([]string(nil), cfg.PreferenceKeywords...),
				Weights: service.PreferenceWeights{
					TagMatch:     cfg.PreferenceTagWeight,
					SourceMatch:  cfg.PreferenceSourceWeight,
					KeywordMatch: cfg.PreferenceKeywordWeight,
				},
			},
		},
	}
}

func mergePreference(base, override service.UserPreference) service.UserPreference {
	merged := base
	if len(override.Tags) > 0 {
		merged.Tags = append([]string(nil), override.Tags...)
	}
	if len(override.Sources) > 0 {
		merged.Sources = append([]string(nil), override.Sources...)
	}
	if len(override.Keywords) > 0 {
		merged.Keywords = append([]string(nil), override.Keywords...)
	}
	if override.Weights.TagMatch > 0 {
		merged.Weights.TagMatch = override.Weights.TagMatch
	}
	if override.Weights.SourceMatch > 0 {
		merged.Weights.SourceMatch = override.Weights.SourceMatch
	}
	if override.Weights.KeywordMatch > 0 {
		merged.Weights.KeywordMatch = override.Weights.KeywordMatch
	}

	return merged
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

func buildScheduler(cfg config.Config, job scheduler.Job) (*scheduler.Scheduler, error) {
	if cfg.ScheduleCron != "" {
		return scheduler.NewCron(cfg.ScheduleCron, job)
	}

	return scheduler.New(cfg.ScheduleInterval, job), nil
}

func buildSenders(cfg config.Config) []service.MarkdownSender {
	senders := make([]service.MarkdownSender, 0, 2)
	if cfg.UseWebhook() {
		senders = append(senders, delivery.NewWebhookSender(cfg.WebhookURL, nil))
	}
	if cfg.UseEmail() {
		senders = append(senders, delivery.NewEmailSender(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.EmailFrom,
			cfg.EmailTo,
			cfg.EmailSubject,
		))
	}

	return senders
}

func newUserPreferenceRepository(cfg config.Config) (repository.UserPreferenceRepository, func() error, error) {
	if cfg.UseMySQL() {
		db, err := sql.Open("mysql", cfg.MySQLDSN)
		if err != nil {
			return nil, nil, err
		}

		repo, err := repository.NewMySQLUserPreferenceRepository(db, cfg.MySQLPreferenceTable)
		if err != nil {
			_ = db.Close()
			return nil, nil, err
		}

		return repo, repo.Close, nil
	}

	repo := repository.NewFileUserPreferenceRepository(filepath.Join(cfg.StorageDir, "preferences", "users.json"))
	return repo, func() error { return nil }, nil
}

func resolveUserPreference(ctx context.Context, repo repository.UserPreferenceRepository, userID string) (service.UserPreference, error) {
	if repo == nil || userID == "" {
		return service.UserPreference{}, nil
	}

	stored, err := repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserPreferenceNotFound) {
			return service.UserPreference{}, nil
		}
		return service.UserPreference{}, err
	}

	return service.UserPreference{
		Tags:     append([]string(nil), stored.Tags...),
		Sources:  append([]string(nil), stored.Sources...),
		Keywords: append([]string(nil), stored.Keywords...),
		Weights: service.PreferenceWeights{
			TagMatch:     stored.Weights.Tag,
			SourceMatch:  stored.Weights.Source,
			KeywordMatch: stored.Weights.Keyword,
		},
	}, nil
}
