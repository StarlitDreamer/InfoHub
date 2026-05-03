package delivery

import (
	"strings"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestRenderMarkdownIncludesTitleAndSummary(t *testing.T) {
	report := RenderMarkdown([]model.NewsItem{
		{Title: "测试标题", Content: "测试摘要", Score: 5},
	})

	if !strings.Contains(report, "# 今日信息") {
		t.Fatal("期望 Markdown 包含日报标题")
	}

	if !strings.Contains(report, "- 标题：测试标题") {
		t.Fatal("期望 Markdown 包含信息标题")
	}
}

func TestRenderMarkdownEmptyItems(t *testing.T) {
	report := RenderMarkdown(nil)

	if !strings.Contains(report, "今日暂无新增信息。") {
		t.Fatal("期望空日报包含无新增信息提示")
	}
}
