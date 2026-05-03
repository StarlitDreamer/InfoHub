package processor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"InfoHub-agent/internal/model"
)

// DedupStore 保存跨运行的去重状态。
type DedupStore interface {
	Seen(ctx context.Context, key string) (bool, error)
	Mark(ctx context.Context, key string) error
}

// FileDedupStore 使用本地 JSON 文件保存已处理内容指纹。
type FileDedupStore struct {
	path string
}

// NewFileDedupStore 创建文件版去重状态存储。
func NewFileDedupStore(path string) *FileDedupStore {
	return &FileDedupStore{path: path}
}

// Seen 判断指定 key 是否已经处理过。
func (s *FileDedupStore) Seen(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	keys, err := s.read()
	if err != nil {
		return false, err
	}

	_, ok := keys[key]
	return ok, nil
}

// Mark 记录指定 key 已经处理过。
func (s *FileDedupStore) Mark(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	keys, err := s.read()
	if err != nil {
		return err
	}

	keys[key] = struct{}{}
	return s.write(keys)
}

// DedupKeys 生成多维度去重 key，覆盖标题、URL 和正文指纹。
func DedupKeys(item model.NewsItem) []string {
	keys := make([]string, 0, 3)

	if value := normalizeDedupText(item.Title); value != "" {
		keys = append(keys, hashDedupValue("title", value))
	}
	if value := normalizeDedupURL(item.URL); value != "" {
		keys = append(keys, hashDedupValue("url", value))
	}
	if value := normalizeDedupText(item.Content); value != "" {
		keys = append(keys, hashDedupValue("content", value))
	}

	return keys
}

// DedupKey 保留单 key 接口，优先返回 URL，其次标题，最后正文指纹。
func DedupKey(item model.NewsItem) string {
	keys := DedupKeys(item)
	for _, prefix := range []string{"url:", "title:", "content:"} {
		for _, key := range keys {
			if strings.HasPrefix(key, prefix) {
				return key
			}
		}
	}

	return ""
}

func normalizeDedupURL(value string) string {
	value = normalizeDedupText(value)
	return strings.TrimRight(value, "/")
}

func normalizeDedupText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	return strings.Join(strings.Fields(value), " ")
}

func hashDedupValue(kind, value string) string {
	sum := sha256.Sum256([]byte(value))
	return kind + ":" + hex.EncodeToString(sum[:])
}

func (s *FileDedupStore) read() (map[string]struct{}, error) {
	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}

		return nil, err
	}

	if len(content) == 0 {
		return map[string]struct{}{}, nil
	}

	var keys []string
	if err := json.Unmarshal(content, &keys); err != nil {
		return nil, err
	}

	result := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		result[key] = struct{}{}
	}

	return result, nil
}

func (s *FileDedupStore) write(keys map[string]struct{}) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	values := make([]string, 0, len(keys))
	for key := range keys {
		values = append(values, key)
	}

	content, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, content, 0o644)
}
