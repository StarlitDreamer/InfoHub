package service

import (
	"context"
	"strings"
	"time"

	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/delivery"
	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/processor"
	"InfoHub-agent/internal/repository"
)

// SearchRequest 表示一次搜索请求。
type SearchRequest struct {
	Query      string
	UserID     string
	Preference UserPreference
}

// SearchResult 表示一次搜索执行结果。
type SearchResult struct {
	Query        string
	ItemCount    int
	DisplayCount int
	GeneratedAt  time.Time
	Markdown     string
	Items        []model.NewsItem
	Warnings     []string
}

// SearchAgent 负责搜索主链路。
type SearchAgent struct {
	crawler        crawler.SearchCrawler
	ai             ai.Processor
	repo           repository.SearchRepository
	reportMaxItems int
	now            func() time.Time
}

// NewSearchAgent 创建搜索服务。
func NewSearchAgent(searchCrawler crawler.SearchCrawler, aiProcessor ai.Processor, repo repository.SearchRepository, reportMaxItems int, now func() time.Time) *SearchAgent {
	if now == nil {
		now = time.Now
	}

	return &SearchAgent{
		crawler:        searchCrawler,
		ai:             aiProcessor,
		repo:           repo,
		reportMaxItems: reportMaxItems,
		now:            now,
	}
}

// Run 执行关键词搜索。
func (a *SearchAgent) Run(ctx context.Context, request SearchRequest) (SearchResult, error) {
	query := strings.TrimSpace(request.Query)
	items, err := a.crawler.Search(ctx, query)
	if err != nil {
		return SearchResult{}, err
	}

	items = processor.DeduplicateItems(items)
	result := make([]model.NewsItem, 0, len(items))
	for _, item := range items {
		item.Query = query
		analysis, err := ai.AnalyzeItem(a.ai, item)
		if err != nil {
			return SearchResult{}, err
		}

		item.Tags = analysis.Tags
		item.Content = strings.TrimSpace(analysis.Summary)
		item.Score = analysis.Score
		if item.Content == "" {
			item.Content = item.Title
		}
		result = append(result, item)
	}

	sorted := SortSearchItems(result, request.Preference, a.now())
	displayItems := LimitItemsBalancedBySource(sorted, a.reportMaxItems)
	warnings := searchWarnings(a.crawler)
	markdown := delivery.RenderSearchMarkdown(query, displayItems, warnings)
	generatedAt := a.now()

	if err := a.repo.Save(ctx, repository.SearchRecord{
		Query:       query,
		GeneratedAt: generatedAt,
		Markdown:    markdown,
		Items:       sorted,
		Warnings:    warnings,
	}); err != nil {
		return SearchResult{}, err
	}

	return SearchResult{
		Query:        query,
		ItemCount:    len(sorted),
		DisplayCount: len(displayItems),
		GeneratedAt:  generatedAt,
		Markdown:     markdown,
		Items:        sorted,
		Warnings:     warnings,
	}, nil
}

func searchWarnings(searchCrawler crawler.SearchCrawler) []string {
	provider, ok := searchCrawler.(interface{ Warnings() []string })
	if !ok {
		return nil
	}

	return provider.Warnings()
}
