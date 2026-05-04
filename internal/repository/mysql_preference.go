package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// MySQLUserPreferenceRepository 使用 MySQL 保存用户偏好。
type MySQLUserPreferenceRepository struct {
	db    *sql.DB
	table string
}

// NewMySQLUserPreferenceRepository 创建 MySQL 用户偏好仓库。
func NewMySQLUserPreferenceRepository(db *sql.DB, table string) (*MySQLUserPreferenceRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("mysql db is nil")
	}
	if !mysqlTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf("invalid mysql table name: %s", table)
	}

	repo := &MySQLUserPreferenceRepository{db: db, table: table}
	if err := repo.ensureTable(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

// Save 保存用户偏好。
func (r *MySQLUserPreferenceRepository) Save(ctx context.Context, record UserPreferenceRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	record = normalizeUserPreferenceRecord(record)
	if record.UserID == "" {
		return ErrUserPreferenceNotFound
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = time.Now().UTC()
	}

	tagsJSON, err := json.Marshal(record.Tags)
	if err != nil {
		return err
	}
	sourcesJSON, err := json.Marshal(record.Sources)
	if err != nil {
		return err
	}
	keywordsJSON, err := json.Marshal(record.Keywords)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (user_id, tags_json, sources_json, keywords_json, tag_weight, source_weight, keyword_weight, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE tags_json = VALUES(tags_json), sources_json = VALUES(sources_json), keywords_json = VALUES(keywords_json), "+
			"tag_weight = VALUES(tag_weight), source_weight = VALUES(source_weight), keyword_weight = VALUES(keyword_weight), updated_at = VALUES(updated_at)",
		r.table,
	)
	_, err = r.db.ExecContext(
		ctx,
		query,
		record.UserID,
		string(tagsJSON),
		string(sourcesJSON),
		string(keywordsJSON),
		record.Weights.Tag,
		record.Weights.Source,
		record.Weights.Keyword,
		record.UpdatedAt.UTC(),
	)
	return err
}

// Get 读取用户偏好。
func (r *MySQLUserPreferenceRepository) Get(ctx context.Context, userID string) (UserPreferenceRecord, error) {
	if err := ctx.Err(); err != nil {
		return UserPreferenceRecord{}, err
	}

	query := fmt.Sprintf(
		"SELECT user_id, tags_json, sources_json, keywords_json, tag_weight, source_weight, keyword_weight, updated_at FROM `%s` WHERE user_id = ? LIMIT 1",
		r.table,
	)

	var (
		record       UserPreferenceRecord
		tagsJSON     string
		sourcesJSON  string
		keywordsJSON string
	)
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&record.UserID,
		&tagsJSON,
		&sourcesJSON,
		&keywordsJSON,
		&record.Weights.Tag,
		&record.Weights.Source,
		&record.Weights.Keyword,
		&record.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserPreferenceRecord{}, ErrUserPreferenceNotFound
		}
		return UserPreferenceRecord{}, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &record.Tags); err != nil {
		return UserPreferenceRecord{}, err
	}
	if err := json.Unmarshal([]byte(sourcesJSON), &record.Sources); err != nil {
		return UserPreferenceRecord{}, err
	}
	if err := json.Unmarshal([]byte(keywordsJSON), &record.Keywords); err != nil {
		return UserPreferenceRecord{}, err
	}

	return normalizeUserPreferenceRecord(record), nil
}

// Close 关闭底层数据库连接。
func (r *MySQLUserPreferenceRepository) Close() error {
	return r.db.Close()
}

func (r *MySQLUserPreferenceRepository) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS `%s` ("+
			"user_id VARCHAR(191) NOT NULL,"+
			"tags_json LONGTEXT NOT NULL,"+
			"sources_json LONGTEXT NOT NULL,"+
			"keywords_json LONGTEXT NOT NULL,"+
			"tag_weight DOUBLE NOT NULL,"+
			"source_weight DOUBLE NOT NULL,"+
			"keyword_weight DOUBLE NOT NULL,"+
			"updated_at DATETIME(6) NOT NULL,"+
			"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,"+
			"PRIMARY KEY (user_id),"+
			"KEY idx_updated_at (updated_at)"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		r.table,
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}
