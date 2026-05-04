package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileUserPreferenceRepository 将用户偏好保存到本地 JSON 文件。
type FileUserPreferenceRepository struct {
	path string
}

// NewFileUserPreferenceRepository 创建文件型用户偏好存储。
func NewFileUserPreferenceRepository(path string) *FileUserPreferenceRepository {
	return &FileUserPreferenceRepository{path: path}
}

// Save 保存指定用户偏好。
func (r *FileUserPreferenceRepository) Save(ctx context.Context, record UserPreferenceRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	record.UserID = strings.TrimSpace(record.UserID)
	if record.UserID == "" {
		return ErrUserPreferenceNotFound
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = time.Now()
	}

	store, err := r.readAll()
	if err != nil {
		return err
	}
	store[record.UserID] = normalizeUserPreferenceRecord(record)

	return r.writeAll(store)
}

// Get 读取指定用户偏好。
func (r *FileUserPreferenceRepository) Get(ctx context.Context, userID string) (UserPreferenceRecord, error) {
	if err := ctx.Err(); err != nil {
		return UserPreferenceRecord{}, err
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return UserPreferenceRecord{}, ErrUserPreferenceNotFound
	}

	store, err := r.readAll()
	if err != nil {
		return UserPreferenceRecord{}, err
	}

	record, ok := store[userID]
	if !ok {
		return UserPreferenceRecord{}, ErrUserPreferenceNotFound
	}

	return normalizeUserPreferenceRecord(record), nil
}

func (r *FileUserPreferenceRepository) readAll() (map[string]UserPreferenceRecord, error) {
	content, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]UserPreferenceRecord{}, nil
		}
		return nil, err
	}

	var store map[string]UserPreferenceRecord
	if err := json.Unmarshal(content, &store); err != nil {
		return nil, err
	}
	if store == nil {
		store = map[string]UserPreferenceRecord{}
	}

	return store, nil
}

func (r *FileUserPreferenceRepository) writeAll(store map[string]UserPreferenceRecord) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.path, content, 0644)
}

func normalizeUserPreferenceRecord(record UserPreferenceRecord) UserPreferenceRecord {
	record.UserID = strings.TrimSpace(record.UserID)
	record.Tags = cloneStringSlice(record.Tags)
	record.Sources = cloneStringSlice(record.Sources)
	record.Keywords = cloneStringSlice(record.Keywords)
	return record
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}

	return result
}
