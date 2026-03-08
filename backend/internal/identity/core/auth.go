package core

import (
	"context"
	"strconv"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/repo"
	"github.com/jackc/pgx/v5"
)

type AuthRepository interface {
	GetCredentialsByLogin(ctx context.Context, login string) (*repo.Credential, error)
	GetCredentialsByUserID(ctx context.Context, userID int64) (*repo.Credential, error)
	CreateCredentials(ctx context.Context, tx pgx.Tx, cred *repo.Credential) (int64, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateStatus(ctx context.Context, userID int64, isActive bool) error
	CreateSession(ctx context.Context, session *repo.RefreshSession) error
	GetSessionByToken(ctx context.Context, token string) (*repo.RefreshSession, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID int64) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type Authenticator interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	GenerateTokens(userID int64, role string) (accessToken, refreshToken string, err error)
	ValidateRefreshToken(token string) (*repo.CustomClaims, error)
}

type AuthService struct {
	repo          AuthRepository
	authenticator Authenticator
	refreshTTL    time.Duration
}

func NewAuthService(repo AuthRepository, auth Authenticator, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		repo:          repo,
		authenticator: auth,
		refreshTTL:    refreshTTL,
	}
}

func (s *AuthService) Login(ctx context.Context, login, password, userAgent, ipAddress string) (int64, string, string, error) {
	cred, err := s.repo.GetCredentialsByLogin(ctx, login)
	if err != nil {
		return 0, "", "", err
	}

	if !cred.IsActive {
		return 0, "", "", ErrUserInactive
	}

	if !s.authenticator.CheckPassword(password, cred.PasswordHash) {
		return 0, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.authenticator.GenerateTokens(cred.UserID, cred.Role)
	if err != nil {
		return 0, "", "", err
	}

	session := &repo.RefreshSession{
		UserID:       cred.UserID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return 0, "", "", err
	}

	return cred.UserID, accessToken, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteSession(ctx, refreshToken)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken, userAgent, ipAddress string) (string, string, int64, error) {
	session, err := s.repo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return "", "", 0, ErrInvalidToken
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.DeleteSession(ctx, refreshToken)
		return "", "", 0, ErrTokenExpired
	}

	cred, err := s.repo.GetCredentialsByUserID(ctx, session.UserID)
	if err != nil {
		return "", "", 0, err
	}

	if !cred.IsActive {
		return "", "", 0, ErrUserInactive
	}

	_ = s.repo.DeleteSession(ctx, refreshToken)

	newAccessToken, newRefreshToken, err := s.authenticator.GenerateTokens(cred.UserID, cred.Role)
	if err != nil {
		return "", "", 0, err
	}

	newSession := &repo.RefreshSession{
		UserID:       cred.UserID,
		RefreshToken: newRefreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateSession(ctx, newSession); err != nil {
		return "", "", 0, err
	}

	return newAccessToken, newRefreshToken, cred.UserID, nil
}

func (s *AuthService) CreateCredentials(ctx context.Context, tx pgx.Tx, login, password, role string) (int64, error) {
	hash, err := s.authenticator.HashPassword(password)
	if err != nil {
		return 0, err
	}

	cred := &repo.Credential{
		Login:        login,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.repo.CreateCredentials(ctx, tx, cred)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, newPassword string) error {
	hash, err := s.authenticator.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, userID, hash)
}

func (s *AuthService) ChangePasswordWithOldPassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	cred, err := s.repo.GetCredentialsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if !s.authenticator.CheckPassword(oldPassword, cred.PasswordHash) {
		return ErrInvalidCredentials
	}

	return s.ChangePassword(ctx, userID, newPassword)
}

func (s *AuthService) ValidateToken(tokenString string) (*repo.CustomClaims, error) {
	return s.authenticator.ValidateRefreshToken(tokenString)
}

func (s *AuthService) ChangeStatus(ctx context.Context, userID int64, isActive bool) error {
	if !isActive {
		_ = s.repo.DeleteUserSessions(ctx, userID)
	}
	return s.repo.UpdateStatus(ctx, userID, isActive)
}

func ParseUserID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}


