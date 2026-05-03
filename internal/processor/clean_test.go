package processor

import "testing"

func TestCleanTextRemovesHTMLAndDecodesEntities(t *testing.T) {
	result := CleanText("<p>AI&nbsp;&amp;&nbsp;Cloud</p>", 100)

	if result != "AI & Cloud" {
		t.Fatalf("清洗结果不符合预期：%q", result)
	}
}

func TestCleanTextMergesWhitespace(t *testing.T) {
	result := CleanText("标题\n\n\t 内容   更多", 100)

	if result != "标题 内容 更多" {
		t.Fatalf("空白合并结果不符合预期：%q", result)
	}
}

func TestCleanTextTruncatesByRune(t *testing.T) {
	result := CleanText("一二三四五", 3)

	if result != "一二三" {
		t.Fatalf("截断结果不符合预期：%q", result)
	}
}
