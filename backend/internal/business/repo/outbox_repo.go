package repo

import (
	"context"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
)

type OutboxRepo struct {
	db DBExecutor
}

func NewOutboxRepo(db DBExecutor) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) Insert(ctx context.Context, eventType string, payload []byte) (string, error) {
	query := `
		INSERT INTO business.outbox (event_type, payload, status, retry_count, created_at, updated_at)
		VALUES ($1, $2, 'PENDING', 0, NOW(), NOW())
		RETURNING id::text
	`
	var id string
	err := r.db.QueryRow(ctx, query, eventType, payload).Scan(&id)
	return id, err
}

func (r *OutboxRepo) FetchPendingBatch(ctx context.Context, limit int) ([]dto.OutboxRecord, error) {
	query := `
		SELECT id::text, event_type, payload, status, retry_count, created_at, updated_at
		FROM business.outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []dto.OutboxRecord
	for rows.Next() {
		var rec dto.OutboxRecord
		if err := rows.Scan(
			&rec.ID, &rec.EventType, &rec.Payload, &rec.Status, &rec.RetryCount, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id string) error {
	query := `
		UPDATE business.outbox
		SET status = 'PUBLISHED', processed_at = NOW(), updated_at = NOW()
		WHERE id = $1::uuid
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id string, errMsg string) error {
	query := `
		UPDATE business.outbox
		SET status = CASE WHEN retry_count >= 5 THEN 'FAILED' ELSE 'PENDING' END,
		    retry_count = retry_count + 1,
		    error_message = $1,
		    updated_at = NOW()
		WHERE id = $2::uuid
	`
	_, err := r.db.Exec(ctx, query, errMsg, id)
	return err
}
