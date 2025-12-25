package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) GetCredentialsByLogin(ctx context.Context, login string) (*Credential, error) {
	cred := &Credential{}
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, login, password_hash, role, is_active, created_at, updated_at
		FROM identity.credentials
		WHERE login = $1
	`, login).Scan(&cred.UserID, &cred.Login, &cred.PasswordHash, &cred.Role, &cred.IsActive, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func (r *AuthRepo) GetCredentialsByUserID(ctx context.Context, userID int64) (*Credential, error) {
	cred := &Credential{}
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, login, password_hash, role, is_active, created_at, updated_at
		FROM identity.credentials
		WHERE user_id = $1
	`, userID).Scan(&cred.UserID, &cred.Login, &cred.PasswordHash, &cred.Role, &cred.IsActive, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func (r *AuthRepo) CreateCredentials(ctx context.Context, tx pgx.Tx, cred *Credential) (int64, error) {
	var userID int64
	err := tx.QueryRow(ctx, `
		INSERT INTO identity.credentials (login, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_id
	`, cred.Login, cred.PasswordHash, cred.Role, cred.IsActive, cred.CreatedAt, cred.UpdatedAt).Scan(&userID)
	return userID, err
}

func (r *AuthRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE identity.credentials SET password_hash = $1, updated_at = $2 WHERE user_id = $3
	`, passwordHash, time.Now(), userID)
	return err
}

func (r *AuthRepo) UpdateStatus(ctx context.Context, userID int64, isActive bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE identity.credentials SET is_active = $1, updated_at = $2 WHERE user_id = $3
	`, isActive, time.Now(), userID)
	return err
}

// Session methods
func (r *AuthRepo) CreateSession(ctx context.Context, session *RefreshSession) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO identity.sessions (user_id, refresh_token, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, session.UserID, session.RefreshToken, session.UserAgent, session.IPAddress, session.ExpiresAt, session.CreatedAt)
	return err
}

func (r *AuthRepo) GetSessionByToken(ctx context.Context, token string) (*RefreshSession, error) {
	session := &RefreshSession{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM identity.sessions
		WHERE refresh_token = $1
	`, token).Scan(&session.ID, &session.UserID, &session.RefreshToken, &session.UserAgent, &session.IPAddress, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *AuthRepo) DeleteSession(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identity.sessions WHERE refresh_token = $1`, token)
	return err
}

func (r *AuthRepo) DeleteUserSessions(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
	return err
}

func (r *AuthRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}
