package service

import (
	"context"
	"fmt"
	"strings"

	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/processor"
)

// Pipeline 聚合采集、处理和 AI 能力。
type Pipeline struct {
	crawler crawler.Crawler
	ai      ai.Processor
	dedup   processor.DedupStore
}

// NewPipeline 创建信息处理流程。
func NewPipeline(crawler crawler.Crawler, ai ai.Processor) *Pipeline {
	return &Pipeline{
		crawler: crawler,
		ai:      ai,
	}
}

// WithDedupStore 配置跨运行去重状态存储。
func (p *Pipeline) WithDedupStore(store processor.DedupStore) *Pipeline {
	p.dedup = store
	return p
}

// Run 执行采集、去重和 AI 摘要处理流程。
func (p *Pipeline) Run() ([]model.NewsItem, error) {
	return p.RunContext(context.Background())
}

// RunContext 执行采集、去重和 AI 摘要处理流程。
func (p *Pipeline) RunContext(ctx context.Context) ([]model.NewsItem, error) {
	items, err := p.crawler.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	items = processor.DeduplicateItems(items)
	result := make([]model.NewsItem, 0, len(items))

	for _, item := range items {
		if p.dedup != nil {
			seen, err := hasSeenAnyDedupKey(ctx, p.dedup, item)
			if err != nil {
				return nil, err
			}
			if seen {
				continue
			}
		}

		analysis, err := ai.AnalyzeItem(p.ai, item)
		if err != nil {
			return nil, err
		}

		summarized := item
		summarized.Tags = analysis.Tags
		summarized.Content = analysis.Summary
		summarized.Score = analysis.Score
		summarized = normalizeSummary(item, summarized)

		if p.dedup != nil {
			if err := markAllDedupKeys(ctx, p.dedup, item); err != nil {
				return nil, err
			}
		}

		result = append(result, summarized)
	}

	return result, nil
}

func hasSeenAnyDedupKey(ctx context.Context, store processor.DedupStore, item model.NewsItem) (bool, error) {
	for _, key := range processor.DedupKeys(item) {
		seen, err := store.Seen(ctx, key)
		if err != nil {
			return false, err
		}
		if seen {
			return true, nil
		}
	}

	return false, nil
}

func markAllDedupKeys(ctx context.Context, store processor.DedupStore, item model.NewsItem) error {
	for _, key := range processor.DedupKeys(item) {
		if err := store.Mark(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func normalizeSummary(original, summarized model.NewsItem) model.NewsItem {
	summarized.Title = strings.TrimSpace(summarized.Title)
	if summarized.Title == "" {
		summarized.Title = original.Title
	}

	summarized.Content = strings.TrimSpace(summarized.Content)
	if summarized.Content == "" {
		summarized.Content = fallbackSummary(original, summarized.Score)
	}

	if summarized.Score < 1 {
		summarized.Score = 1
	}
	if summarized.Score > 5 {
		summarized.Score = 5
	}

	if summarized.Source == "" {
		summarized.Source = original.Source
	}
	if summarized.URL == "" {
		summarized.URL = original.URL
	}
	if summarized.PublishTime.IsZero() {
		summarized.PublishTime = original.PublishTime
	}

	return summarized
}

func fallbackSummary(item model.NewsItem, score float64) string {
	normalizedScore := score
	if normalizedScore < 1 || normalizedScore > 5 {
		normalizedScore = 1
	}

	content := strings.TrimSpace(item.Content)
	if content == "" {
		content = item.Title
	}

	return fmt.Sprintf(
		"【标题】%s\n【发生了什么】%s\n【为什么重要】该信息可能影响后续技术选型或行动判断。\n【影响】建议持续关注相关进展。\n【评分】%.0f",
		item.Title,
		content,
		normalizedScore,
	)
}
