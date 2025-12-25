package main

import (
	"context"
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/nats"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmployeeEventHandler struct {
	pool *pgxpool.Pool
}

func NewEmployeeEventHandler(pool *pgxpool.Pool) *EmployeeEventHandler {
	return &EmployeeEventHandler{pool: pool}
}

func (h *EmployeeEventHandler) HandleEmployeeEvent(ctx context.Context, event *nats.EmployeeEvent) error {
	// Sync employee data to local database for joins
	query := `
		INSERT INTO users_meta (user_id, full_name, photo_url, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			photo_url = EXCLUDED.photo_url,
			updated_at = NOW()
	`
	_, err := h.pool.Exec(ctx, query, event.UserID, event.FullName, event.PhotoURL)
	if err != nil {
		log.Printf("failed to upsert user meta for %d: %v", event.UserID, err)
		return err
	}
	log.Printf("User meta upserted: %d (%s)", event.UserID, event.FullName)
	return nil
}
