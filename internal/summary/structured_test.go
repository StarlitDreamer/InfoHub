package summary

import (
	"strings"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestParseStructuredSummary(t *testing.T) {
	item := model.NewsItem{
		Title:   "原始标题",
		Content: "【标题】测试标题\n【发生了什么】发生了摘要\n【为什么重要】原因摘要\n【影响】影响摘要\n【评分】4",
	}

	result := Parse(item)

	if result.Title != "测试标题" || result.WhatHappened != "发生了摘要" {
		t.Fatalf("expected structured fields to be parsed, got %+v", result)
	}
	if result.WhyImportant != "原因摘要" || result.Impact != "影响摘要" {
		t.Fatalf("expected importance and impact to be parsed, got %+v", result)
	}
}

func TestParseFallsBackForPlainSummary(t *testing.T) {
	item := model.NewsItem{
		Title:   "原始标题",
		Content: "普通摘要",
	}

	result := Parse(item)

	if result.Title != "原始标题" || result.WhatHappened != "普通摘要" {
		t.Fatalf("expected plain text fallback, got %+v", result)
	}
	if result.WhyImportant == "" || result.Impact == "" {
		t.Fatalf("expected default importance and impact, got %+v", result)
	}
}

func TestRecommendActionReturnsUnifiedLabelAndDescription(t *testing.T) {
	action := RecommendAction(
		model.NewsItem{Score: 3, Tags: []string{"Database"}},
		Structured{WhyImportant: "数据库性能值得关注"},
	)

	if action.Label != "专项验证" {
		t.Fatalf("expected database label, got %+v", action)
	}
	if !strings.Contains(action.Description, "专项验证") {
		t.Fatalf("expected database description, got %+v", action)
	}
}
