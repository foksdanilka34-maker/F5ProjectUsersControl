package repo

import (
	"context"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
)

type AuthRepo struct {
	db DBExecutor
}

func NewAuthRepo(db DBExecutor) *AuthRepo {
	return &AuthRepo{db: db}
}

func (r *AuthRepo) GetByLogin(ctx context.Context, login string) (*dto.UserInfo, string, bool, error) {
	query := `
		SELECT c.user_id, c.login, c.password_hash, c.role, c.is_active,
		       COALESCE(p.first_name || ' ' || p.last_name, c.login) as full_name,
		       p.avatar_url
		FROM identity.credentials c
		LEFT JOIN identity.profiles p ON c.user_id = p.id
		WHERE c.login = $1
	`
	var user dto.UserInfo
	var passHash string
	var isActive bool
	var avatarURL *string

	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&passHash,
		&user.Role,
		&isActive,
		&user.FullName,
		&avatarURL,
	)
	if err != nil {
		return nil, "", false, err
	}
	if avatarURL != nil {
		user.AvatarURL = *avatarURL
	}
	return &user, passHash, isActive, nil
}

func (r *AuthRepo) GetByID(ctx context.Context, userID int64) (*dto.UserInfo, bool, error) {
	query := `
		SELECT c.user_id, c.login, c.role, c.is_active,
		       COALESCE(p.first_name || ' ' || p.last_name, c.login) as full_name,
		       p.avatar_url
		FROM identity.credentials c
		LEFT JOIN identity.profiles p ON c.user_id = p.id
		WHERE c.user_id = $1
	`
	var user dto.UserInfo
	var isActive bool
	var avatarURL *string

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Login,
		&user.Role,
		&isActive,
		&user.FullName,
		&avatarURL,
	)
	if err != nil {
		return nil, false, err
	}
	if avatarURL != nil {
		user.AvatarURL = *avatarURL
	}
	return &user, isActive, nil
}

func (r *AuthRepo) CreateCredentials(ctx context.Context, login, passwordHash, role string) (int64, error) {
	query := `
		INSERT INTO identity.credentials (login, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3::identity.user_role, true, NOW(), NOW())
		RETURNING user_id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, login, passwordHash, role).Scan(&id)
	return id, err
}

func (r *AuthRepo) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	query := `
		UPDATE identity.credentials
		SET password_hash = $1, updated_at = NOW()
		WHERE user_id = $2
	`
	_, err := r.db.Exec(ctx, query, newHash, userID)
	return err
}

func (r *AuthRepo) UpdateStatus(ctx context.Context, userID int64, isActive bool) error {
	query := `
		UPDATE identity.credentials
		SET is_active = $1, updated_at = NOW()
		WHERE user_id = $2
	`
	_, err := r.db.Exec(ctx, query, isActive, userID)
	return err
}

func (r *AuthRepo) CreateSession(ctx context.Context, s *dto.Session) error {
	query := `
		INSERT INTO identity.sessions (user_id, refresh_token, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, s.UserID, s.RefreshToken, s.UserAgent, s.IPAddress, s.ExpiresAt, s.CreatedAt)
	return err
}

func (r *AuthRepo) GetSession(ctx context.Context, refreshToken string) (*dto.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM identity.sessions
		WHERE refresh_token = $1
	`
	var s dto.Session
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(
		&s.ID, &s.UserID, &s.RefreshToken, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *AuthRepo) DeleteSession(ctx context.Context, refreshToken string) error {
	query := `DELETE FROM identity.sessions WHERE refresh_token = $1`
	_, err := r.db.Exec(ctx, query, refreshToken)
	return err
}

func (r *AuthRepo) DeleteUserSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM identity.sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
