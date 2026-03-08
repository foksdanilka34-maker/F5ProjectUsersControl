package repo

import (
	"context"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type LogRepo struct {
	db *pgxpool.Pool
}

func NewLogRepo(db *pgxpool.Pool) *LogRepo {
	return &LogRepo{db: db}
}

func (r *LogRepo) Create(ctx context.Context, entry *LogEntry) error {
	metadataJSON, _ := json.Marshal(entry.Metadata)
	query := `
		INSERT INTO logs (id, service, level, message, user_id, request_id, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		entry.ID, entry.Service, entry.Level, entry.Message,
		entry.UserID, entry.RequestID, metadataJSON, entry.Timestamp,
	)
	return err
}

func (r *LogRepo) List(ctx context.Context, filter *LogFilter) ([]*LogEntry, int, error) {

	countBuilder := psql.Select("COUNT(*)").From("logs")

	if filter.Service != "" {
		countBuilder = countBuilder.Where(sq.Eq{"service": filter.Service})
	}
	if filter.Level != "" {
		countBuilder = countBuilder.Where(sq.Eq{"level": filter.Level})
	}
	if filter.UserID != "" {
		countBuilder = countBuilder.Where(sq.Eq{"user_id": filter.UserID})
	}
	if filter.RequestID != "" {
		countBuilder = countBuilder.Where(sq.Eq{"request_id": filter.RequestID})
	}
	if filter.StartTime != nil {
		countBuilder = countBuilder.Where(sq.GtOrEq{"timestamp": *filter.StartTime})
	}
	if filter.EndTime != nil {
		countBuilder = countBuilder.Where(sq.LtOrEq{"timestamp": *filter.EndTime})
	}

	countQuery, countArgs, _ := countBuilder.ToSql()
	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryBuilder := psql.Select("id", "service", "level", "message", "user_id", "request_id", "metadata", "timestamp").
		From("logs")

	if filter.Service != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"service": filter.Service})
	}
	if filter.Level != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"level": filter.Level})
	}
	if filter.UserID != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"user_id": filter.UserID})
	}
	if filter.RequestID != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"request_id": filter.RequestID})
	}
	if filter.StartTime != nil {
		queryBuilder = queryBuilder.Where(sq.GtOrEq{"timestamp": *filter.StartTime})
	}
	if filter.EndTime != nil {
		queryBuilder = queryBuilder.Where(sq.LtOrEq{"timestamp": *filter.EndTime})
	}

	queryBuilder = queryBuilder.OrderBy("timestamp DESC").Limit(uint64(filter.PageSize)).Offset(uint64(filter.Offset))

	query, args, _ := queryBuilder.ToSql()
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*LogEntry
	for rows.Next() {
		var entry LogEntry
		var metadataJSON []byte
		if err := rows.Scan(
			&entry.ID, &entry.Service, &entry.Level, &entry.Message,
			&entry.UserID, &entry.RequestID, &metadataJSON, &entry.Timestamp,
		); err != nil {
			return nil, 0, err
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &entry.Metadata)
		}
		entries = append(entries, &entry)
	}

	return entries, total, nil
}

func (r *LogRepo) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	query, args, _ := psql.Delete("logs").
		Where("timestamp < NOW() - INTERVAL '1 day' * ?", days).
		ToSql()

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}


