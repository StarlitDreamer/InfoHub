// Package ai 定义信息摘要、分类和评分能力。
package ai

import "InfoHub-agent/internal/model"

// Processor 表示 AI 信息处理器。
type Processor interface {
	Summarize(item model.NewsItem) (model.NewsItem, error)
}
