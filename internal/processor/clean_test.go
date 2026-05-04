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

func TestCleanTextRemovesNoisyBlocksAndURLs(t *testing.T) {
	input := `<article><p>正文第一段</p><script>alert('x')</script><p>正文第二段 https://example.com/readmore</p><footer>Copyright 2026</footer></article>`

	result := CleanText(input, 200)

	if result != "正文第一段 正文第二段" {
		t.Fatalf("噪声清洗结果不符合预期：%q", result)
	}
}

func TestCleanTextRemovesBoilerplateAndDuplicateLines(t *testing.T) {
	input := "标题一\n阅读更多\n标题一\n关注我们\n核心内容"

	result := CleanText(input, 200)

	if result != "标题一 核心内容" {
		t.Fatalf("样板行和重复行清洗结果不符合预期：%q", result)
	}
}
