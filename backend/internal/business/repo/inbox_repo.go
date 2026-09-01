package repo

import (
	"context"
)

type InboxRepo struct {
	db DBExecutor
}

func NewInboxRepo(db DBExecutor) *InboxRepo {
	return &InboxRepo{db: db}
}

func (r *InboxRepo) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM business.processed_events WHERE event_id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, eventID).Scan(&exists)
	return exists, err
}

func (r *InboxRepo) MarkProcessed(ctx context.Context, eventID, eventType string) error {
	query := `
		INSERT INTO business.processed_events (event_id, event_type, processed_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (event_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, eventID, eventType)
	return err
}

func (r *InboxRepo) UpsertUserMeta(ctx context.Context, userID int64, fullName string, photoURL *string) error {
	query := `
		INSERT INTO business.users_meta (user_id, full_name, photo_url, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			photo_url = EXCLUDED.photo_url,
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, userID, fullName, photoURL)
	return err
}

func (r *InboxRepo) DeleteUserMeta(ctx context.Context, userID int64) error {
	query := `DELETE FROM business.users_meta WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
