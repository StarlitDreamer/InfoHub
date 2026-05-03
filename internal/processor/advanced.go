package processor

import (
	"math"
	"strings"

	"InfoHub-agent/internal/model"
)

// EmbeddingProvider 为内容生成向量表示。
type EmbeddingProvider interface {
	Embed(text string) []float64
}

// KeywordEmbeddingProvider 使用关键词频次生成轻量向量。
type KeywordEmbeddingProvider struct {
	vocabulary []string
}

// NewKeywordEmbeddingProvider 创建关键词向量生成器。
func NewKeywordEmbeddingProvider(vocabulary []string) *KeywordEmbeddingProvider {
	return &KeywordEmbeddingProvider{vocabulary: vocabulary}
}

// Embed 根据词表统计文本中的关键词出现次数。
func (p *KeywordEmbeddingProvider) Embed(text string) []float64 {
	vector := make([]float64, len(p.vocabulary))
	lower := strings.ToLower(text)

	for index, word := range p.vocabulary {
		vector[index] = float64(strings.Count(lower, strings.ToLower(word)))
	}

	return vector
}

// DeduplicateByEmbedding 按向量相似度合并相似内容，保留第一条。
func DeduplicateByEmbedding(items []model.NewsItem, provider EmbeddingProvider, threshold float64) []model.NewsItem {
	result := make([]model.NewsItem, 0, len(items))
	vectors := make([][]float64, 0, len(items))

	for _, item := range items {
		vector := provider.Embed(item.Title + " " + item.Content)
		if hasSimilarVector(vector, vectors, threshold) {
			continue
		}

		result = append(result, item)
		vectors = append(vectors, vector)
	}

	return result
}

func hasSimilarVector(vector []float64, vectors [][]float64, threshold float64) bool {
	for _, existing := range vectors {
		if cosineSimilarity(vector, existing) >= threshold {
			return true
		}
	}

	return false
}

func cosineSimilarity(left, right []float64) float64 {
	var dot, leftNorm, rightNorm float64
	for index := range left {
		dot += left[index] * right[index]
		leftNorm += left[index] * left[index]
		rightNorm += right[index] * right[index]
	}

	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}

	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}
