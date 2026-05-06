package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MySQLSearchRepository 将搜索结果保存到 MySQL。
type MySQLSearchRepository struct {
	db    *sql.DB
	table string
}

// NewMySQLSearchRepository 创建 MySQL 搜索结果仓储。
func NewMySQLSearchRepository(db *sql.DB, table string) (*MySQLSearchRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("mysql db is nil")
	}
	if !mysqlTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf("invalid mysql table name: %s", table)
	}

	repo := &MySQLSearchRepository{db: db, table: table}
	if err := repo.ensureTable(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

// Save 保存搜索结果。
func (r *MySQLSearchRepository) Save(ctx context.Context, record SearchRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	itemsJSON, err := json.Marshal(record.Items)
	if err != nil {
		return err
	}
	warningsJSON, err := json.Marshal(record.Warnings)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (query_text, generated_at, markdown, items_json, warnings_json) VALUES (?, ?, ?, ?, ?)",
		r.table,
	)
	_, err = r.db.ExecContext(ctx, query, record.Query, record.GeneratedAt.UTC(), record.Markdown, string(itemsJSON), string(warningsJSON))
	return err
}

// Latest 读取最近一次搜索结果。
func (r *MySQLSearchRepository) Latest(ctx context.Context) (SearchRecord, error) {
	query := fmt.Sprintf(
		"SELECT query_text, generated_at, markdown, items_json, warnings_json FROM `%s` ORDER BY generated_at DESC, id DESC LIMIT 1",
		r.table,
	)

	return r.readSingle(ctx, query)
}

// Get 按名称读取指定搜索结果。
func (r *MySQLSearchRepository) Get(ctx context.Context, name string) (SearchRecord, error) {
	generatedAt, err := time.Parse("20060102-150405", strings.TrimSpace(name))
	if err != nil {
		return SearchRecord{}, ErrReportNotFound
	}
	query := fmt.Sprintf(
		"SELECT query_text, generated_at, markdown, items_json, warnings_json FROM `%s` WHERE generated_at = ? ORDER BY id DESC LIMIT 1",
		r.table,
	)

	return r.readSingle(ctx, query, generatedAt.UTC())
}

// List 返回搜索历史索引。
func (r *MySQLSearchRepository) List(ctx context.Context) ([]SearchMetadata, error) {
	query := fmt.Sprintf(
		"SELECT query_text, generated_at, markdown, items_json FROM `%s` ORDER BY generated_at DESC, id DESC",
		r.table,
	)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]SearchMetadata, 0)
	for rows.Next() {
		var (
			queryText   string
			generatedAt time.Time
			markdown    string
			itemsJSON   string
		)
		if err := rows.Scan(&queryText, &generatedAt, &markdown, &itemsJSON); err != nil {
			return nil, err
		}
		items, err := decodeItemsJSON(itemsJSON)
		if err != nil {
			return nil, err
		}
		result = append(result, BuildSearchMetadata(generatedAt.UTC().Format("20060102-150405"), queryText, markdown, items, generatedAt, 2))
	}

	return result, rows.Err()
}

// Close 关闭数据库连接。
func (r *MySQLSearchRepository) Close() error {
	return r.db.Close()
}

func (r *MySQLSearchRepository) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS `%s` ("+
			"id BIGINT NOT NULL AUTO_INCREMENT,"+
			"query_text VARCHAR(255) NOT NULL,"+
			"generated_at DATETIME(6) NOT NULL,"+
			"markdown LONGTEXT NOT NULL,"+
			"items_json LONGTEXT NOT NULL,"+
			"warnings_json LONGTEXT NOT NULL,"+
			"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,"+
			"PRIMARY KEY (id),"+
			"KEY idx_generated_at (generated_at)"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		r.table,
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *MySQLSearchRepository) readSingle(ctx context.Context, query string, args ...any) (SearchRecord, error) {
	var (
		queryText    string
		generatedAt  time.Time
		markdown     string
		itemsJSON    string
		warningsJSON string
	)
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&queryText, &generatedAt, &markdown, &itemsJSON, &warningsJSON); err != nil {
		if err == sql.ErrNoRows {
			return SearchRecord{}, ErrReportNotFound
		}
		return SearchRecord{}, err
	}

	items, err := decodeItemsJSON(itemsJSON)
	if err != nil {
		return SearchRecord{}, err
	}
	var warnings []string
	if err := json.Unmarshal([]byte(warningsJSON), &warnings); err != nil {
		return SearchRecord{}, err
	}

	return SearchRecord{
		Query:       queryText,
		GeneratedAt: generatedAt,
		Markdown:    markdown,
		Items:       items,
		Warnings:    warnings,
	}, nil
}
