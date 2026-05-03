// Package service 串联信息处理主业务流程。
package service

import (
	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/processor"
)

// Pipeline 聚合采集、处理和 AI 能力。
type Pipeline struct {
	crawler crawler.Crawler
	ai      ai.Processor
}

// NewPipeline 创建信息处理流程。
func NewPipeline(crawler crawler.Crawler, ai ai.Processor) *Pipeline {
	return &Pipeline{
		crawler: crawler,
		ai:      ai,
	}
}

// Run 执行采集、去重和 AI 摘要处理流程。
func (p *Pipeline) Run() ([]model.NewsItem, error) {
	items, err := p.crawler.Fetch()
	if err != nil {
		return nil, err
	}

	items = processor.DeduplicateByTitle(items)
	result := make([]model.NewsItem, 0, len(items))

	for _, item := range items {
		summarized, err := p.ai.Summarize(item)
		if err != nil {
			return nil, err
		}

		result = append(result, summarized)
	}

	return result, nil
}
