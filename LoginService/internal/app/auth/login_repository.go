package auth

import (
	"context"
	"log"

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

type CredentialStorage interface {
	GetCrendentialsByUser(ctx context.Context, login string) (*auth.Credential, error)
	GetCrendentialsByID(ctx context.Context, userID string) (*auth.Credential, error)
	ChangeUserStatus(ctx context.Context, userID string, isActive bool) error
	CreateCredentials(ctx context.Context, cr *auth.Credential) error
	PasswordHashUpdate(ctx context.Context, passwordHash, userID string) error
}

func (s *Storage) GetCrendentialsByUser(ctx context.Context, login string) (*auth.Credential, error) {
	result := &auth.Credential{}
	query := `SELECT user_id, login, password_hash, role, created_at, updated_at, is_active
			  FROM auth.credentials WHERE login = $1`
	err := s.pgx.QueryRow(ctx, query, login).Scan(&result.UserID, &result.Login, &result.Password, &result.Role, &result.CreatedAt, &result.UpdatedAt, &result.IsActive)
	if err != nil {
		log.Printf("error during sql expression at method GetCredentialsByUser: %v", err)
		return nil, err
	}
	return result, nil
}

func (s *Storage) GetCrendentialsByID(ctx context.Context, userID string) (*auth.Credential, error) {
	result := &auth.Credential{}
	query := `SELECT user_id, login, password_hash, role, created_at, updated_at, is_active
			  FROM auth.credentials WHERE user_id = $1`
	err := s.pgx.QueryRow(ctx, query, userID).Scan(
		&result.UserID,
		&result.Login,
		&result.Password,
		&result.Role,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		log.Printf("error during sql expression at method GetCredentialsByID: %v", err)
		return nil, err
	}
	return result, nil
}

func (s *Storage) ChangeUserStatus(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE auth.credentials SET is_active = $1 WHERE user_id = $2`
	_, err := s.pgx.Exec(ctx, query, isActive, userID)
	if err != nil {
		log.Printf("error during sql expression at method ChangeUserStatus: %v", err)
		return err
	}
	return nil
}

func (s *Storage) CreateCredentials(ctx context.Context, cr *auth.Credential) error {
	query := `INSERT INTO auth.credentials (user_id, role, login, password_hash)
				VALUES ($1, $2, $3, $4)`
	tags, err := s.pgx.Exec(ctx, query, cr.UserID, cr.Role, cr.Login, cr.Password)
	if err != nil {
		log.Printf("error during sql expression at method CreateCredentials: %v", err)
		return err
	}
	if tags.RowsAffected() == 0 {
		log.Printf("unable to insert data at method CreateCredentials")
		return err
	}
	return nil
}

func (s *Storage) PasswordHashUpdate(ctx context.Context, passwordHash string, userID string) error {
	query := `UPDATE auth.credentials SET password_hash = $1 WHERE user_id = $2`
	_, err := s.pgx.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		log.Printf("error during sql expression at method PasswordHashUpdate: %v", err)
		return err
	}
	return nil
}
