package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type SessionStorage interface {
	CreateSession(ctx context.Context, session *RefreshSession) error
	UpdateSession(ctx context.Context, oldToken string, newToken string, newExpiresAt time.Time) (*RefreshSession, error)
	DeleteSession(ctx context.Context, token string) error
	GetSessionByToken(ctx context.Context, refreshToken string) (*RefreshSession, error)
}

func (s *Storage) CreateSession(ctx context.Context, session *RefreshSession) error {
	query := `INSERT INTO auth.sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
				VALUES ($1, $2, $3, $4, $5)`
	tags, err := s.pgx.Exec(ctx, query, session.UserID, session.RefreshToken, session.UserAgent, session.IPAddress, session.ExpiresAt)

	if err != nil {
		log.Printf("Error during creating session, session dont created, %v", err)
		return err
	}

	if tags.RowsAffected() == 0 {
		log.Print("Error during creating session, session dont created, zero rows affected")
		return err
	}
	return nil
}

func (s *Storage) UpdateSession(ctx context.Context, oldTokenHash string, newTokenHash string, newExpiresAt time.Time) (*RefreshSession, error) {
	query := `UPDATE auth.sessions 
        	SET refresh_token = $1, expires_at = $2, updated_at = NOW()
        	WHERE refresh_token = $3
        	RETURNING id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at`

	updatedSession := &RefreshSession{}

	err := s.pgx.QueryRow(ctx, query, newTokenHash, newExpiresAt, oldTokenHash).Scan(
		&updatedSession.ID,
		&updatedSession.UserID,
		&updatedSession.RefreshToken,
		&updatedSession.UserAgent,
		&updatedSession.IPAddress,
		&updatedSession.ExpiresAt,
		&updatedSession.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("error during creating session, session is not created%v", err)
		}
		return nil, err
	}

	return updatedSession, nil
}

func (s *Storage) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM auth.sessions WHERE refresh_token = $1`
	tags, err := s.pgx.Exec(ctx, query, token)

	if err != nil {
		log.Printf("Error during deleteing session, session dont deleted, %v", err)
		return err
	}

	if tags.RowsAffected() == 0 {
		log.Print("Session dont deleted, zero rows affected")
		return err
	}
	return nil
}

func (s *Storage) GetSessionByToken(ctx context.Context, refreshToken string) (*RefreshSession, error) {
	query := `SELECT id, user_id, refresh_token, user_agent, ip_address,
				expires_at, created_at FROM auth.sessions WHERE
				refresh_token = $1`

	refSession := &RefreshSession{}
	err := s.pgx.QueryRow(ctx, query, refreshToken).Scan(
		&refSession.ID,
		&refSession.UserID,
		&refSession.RefreshToken,
		&refSession.UserAgent,
		&refSession.IPAddress,
		&refSession.ExpiresAt,
		&refSession.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("session is not exist, %v", err)
		}
		return nil, err
	}
	return refSession, nil
}
