// Package ai 定义信息分类、摘要和评分能力。
package ai

import "InfoHub-agent/internal/model"

// Analysis 表示一次 AI 分析结果。
type Analysis struct {
	Tags    []string
	Summary string
	Score   float64
}

// Classifier 表示分类能力。
type Classifier interface {
	Classify(item model.NewsItem) ([]string, error)
}

// Summarizer 表示摘要能力。
type Summarizer interface {
	Summarize(item model.NewsItem) (string, error)
}

// Scorer 表示评分能力。
type Scorer interface {
	Score(item model.NewsItem) (float64, error)
}

// Processor 表示 Agent 使用的 AI 处理器。
type Processor interface {
	Classifier
	Summarizer
	Scorer
}

// Analyzer 表示支持一次性返回完整分析结果的处理器。
type Analyzer interface {
	Analyze(item model.NewsItem) (Analysis, error)
}

// AnalyzeItem 优先使用聚合分析能力，否则按分类、摘要、评分顺序执行。
func AnalyzeItem(processor Processor, item model.NewsItem) (Analysis, error) {
	if analyzer, ok := processor.(Analyzer); ok {
		return analyzer.Analyze(item)
	}

	tags, err := processor.Classify(item)
	if err != nil {
		return Analysis{}, err
	}

	summary, err := processor.Summarize(item)
	if err != nil {
		return Analysis{}, err
	}

	score, err := processor.Score(item)
	if err != nil {
		return Analysis{}, err
	}

	return Analysis{
		Tags:    tags,
		Summary: summary,
		Score:   score,
	}, nil
}
