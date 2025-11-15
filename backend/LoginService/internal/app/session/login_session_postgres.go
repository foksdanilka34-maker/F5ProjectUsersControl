package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pgx *pgxpool.Pool
}

func NewStorage(p *pgxpool.Pool) *Storage {
	return &Storage{
		pgx: p,
	}
}

type SessionStorage interface {
	CreateSession(ctx context.Context, session *auth.RefreshSession) error
	UpdateSession(ctx context.Context, oldToken string, newToken string, newExpiresAt time.Time) (*auth.RefreshSession, error)
	DeleteSession(ctx context.Context, token string) error
	GetSessionByToken(ctx context.Context, refreshToken string) (*auth.RefreshSession, error)
}

func (s *Storage) CreateSession(ctx context.Context, session *auth.RefreshSession) error {
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

func (s *Storage) UpdateSession(ctx context.Context, oldTokenHash string, newTokenHash string, newExpiresAt time.Time) (*auth.RefreshSession, error) {
	query := `UPDATE auth.sessions 
        	SET refresh_token = $1, expires_at = $2
        	WHERE refresh_token = $3
        	RETURNING id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at`

	updatedSession := &auth.RefreshSession{}

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
			log.Printf("error during updating session, session is not updated%v", err)
			return nil, err
		}
		log.Printf("system error during updating session, session is not updated%v", err)
		return nil, err
	}

	return updatedSession, nil
}

func (s *Storage) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM auth.sessions WHERE refresh_token = $1`
	tags, err := s.pgx.Exec(ctx, query, token)

	log.Printf("Delete session method postgres activated")
	if err != nil {
		log.Printf("Error during deleting session, session not deleted, %v", err)
		return err
	}

	if tags.RowsAffected() == 0 {
		log.Print("Session not deleted, zero rows affected")
		return fmt.Errorf("session not found")
	}

	log.Printf("Session deleted successfully, rows affected: %d", tags.RowsAffected())
	return nil
}

func (s *Storage) GetSessionByToken(ctx context.Context, refreshToken string) (*auth.RefreshSession, error) {
	query := `SELECT id, user_id, refresh_token, user_agent, ip_address,
				expires_at, created_at FROM auth.sessions WHERE
				refresh_token = $1`

	refSession := &auth.RefreshSession{}
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
