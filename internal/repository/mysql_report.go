package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
)

var mysqlTableNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type MySQLReportRepository struct {
	db    *sql.DB
	table string
}

func NewMySQLReportRepository(db *sql.DB, table string) (*MySQLReportRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("mysql db is nil")
	}

	if !mysqlTableNamePattern.MatchString(table) {
		return nil, fmt.Errorf("invalid mysql table name: %s", table)
	}

	repo := &MySQLReportRepository{db: db, table: table}
	if err := repo.ensureTable(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *MySQLReportRepository) Save(ctx context.Context, record ReportRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	items, err := json.Marshal(record.Items)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (generated_at, markdown, items_json) VALUES (?, ?, ?)",
		r.table,
	)
	_, err = r.db.ExecContext(ctx, query, record.GeneratedAt.UTC(), record.Markdown, string(items))
	return err
}

func (r *MySQLReportRepository) Latest(ctx context.Context) (ReportRecord, error) {
	if err := ctx.Err(); err != nil {
		return ReportRecord{}, err
	}

	query := fmt.Sprintf(
		"SELECT generated_at, markdown, items_json FROM `%s` ORDER BY generated_at DESC, id DESC LIMIT 1",
		r.table,
	)

	var (
		generatedAt time.Time
		markdown    string
		itemsJSON   string
	)
	if err := r.db.QueryRowContext(ctx, query).Scan(&generatedAt, &markdown, &itemsJSON); err != nil {
		if err == sql.ErrNoRows {
			return ReportRecord{}, ErrReportNotFound
		}

		return ReportRecord{}, err
	}

	items, err := decodeItemsJSON(itemsJSON)
	if err != nil {
		return ReportRecord{}, err
	}

	return ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    markdown,
		Items:       items,
	}, nil
}

// Get 按名称读取指定日报详情。
func (r *MySQLReportRepository) Get(ctx context.Context, name string) (ReportRecord, error) {
	if err := ctx.Err(); err != nil {
		return ReportRecord{}, err
	}

	generatedAt, err := time.Parse("20060102-150405", strings.TrimSpace(name))
	if err != nil {
		return ReportRecord{}, ErrReportNotFound
	}

	query := fmt.Sprintf(
		"SELECT generated_at, markdown, items_json FROM `%s` WHERE generated_at = ? ORDER BY id DESC LIMIT 1",
		r.table,
	)

	var (
		markdown  string
		itemsJSON string
	)
	if err := r.db.QueryRowContext(ctx, query, generatedAt.UTC()).Scan(&generatedAt, &markdown, &itemsJSON); err != nil {
		if err == sql.ErrNoRows {
			return ReportRecord{}, ErrReportNotFound
		}

		return ReportRecord{}, err
	}

	items, err := decodeItemsJSON(itemsJSON)
	if err != nil {
		return ReportRecord{}, err
	}

	return ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    markdown,
		Items:       items,
	}, nil
}

func (r *MySQLReportRepository) List(ctx context.Context) ([]ReportMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT generated_at, markdown, items_json FROM `%s` ORDER BY generated_at DESC, id DESC",
		r.table,
	)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]ReportMetadata, 0)
	for rows.Next() {
		var (
			generatedAt time.Time
			markdown    string
			itemsJSON   string
		)
		if err := rows.Scan(&generatedAt, &markdown, &itemsJSON); err != nil {
			return nil, err
		}

		items, err := decodeItemsJSON(itemsJSON)
		if err != nil {
			return nil, err
		}

		records = append(records, BuildReportMetadata(
			generatedAt.UTC().Format("20060102-150405"),
			"",
			"",
			markdown,
			items,
			generatedAt,
			2,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *MySQLReportRepository) Close() error {
	return r.db.Close()
}

func (r *MySQLReportRepository) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS `%s` ("+
			"id BIGINT NOT NULL AUTO_INCREMENT,"+
			"generated_at DATETIME(6) NOT NULL,"+
			"markdown LONGTEXT NOT NULL,"+
			"items_json LONGTEXT NOT NULL,"+
			"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,"+
			"PRIMARY KEY (id),"+
			"KEY idx_generated_at (generated_at)"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		r.table,
	)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func decodeItemsJSON(value string) ([]model.NewsItem, error) {
	var items []model.NewsItem
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return nil, err
	}

	return items, nil
}
