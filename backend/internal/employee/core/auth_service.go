package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrSessionNotFound    = errors.New("session not found or expired")
	ErrUnauthorized       = errors.New("unauthorized")
)

type AuthRepository interface {
	GetByLogin(ctx context.Context, login string) (*dto.UserInfo, string, bool, error)
	GetByID(ctx context.Context, userID int64) (*dto.UserInfo, bool, error)
	CreateCredentials(ctx context.Context, login, passwordHash, role string) (int64, error)
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
	UpdateStatus(ctx context.Context, userID int64, isActive bool) error
	CreateSession(ctx context.Context, s *dto.Session) error
	GetSession(ctx context.Context, refreshToken string) (*dto.Session, error)
	DeleteSession(ctx context.Context, refreshToken string) error
	DeleteUserSessions(ctx context.Context, userID int64) error
}

type TokenClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo       AuthRepository
	txManager  TxManager
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(
	repo AuthRepository,
	txManager TxManager,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		repo:       repo,
		txManager:  txManager,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	user, passHash, isActive, err := s.repo.GetByLogin(ctx, req.Login)
	if err != nil {
		return dto.LoginResponse{}, ErrInvalidCredentials
	}

	if !isActive {
		return dto.LoginResponse{}, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(req.Password)); err != nil {
		return dto.LoginResponse{}, ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken := uuid.New().String()
	session := &dto.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    req.UserAgent,
		IPAddress:    req.IPAddress,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return dto.LoginResponse{}, fmt.Errorf("failed to save session: %w", err)
	}

	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.repo.DeleteSession(ctx, refreshToken)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken, userAgent, ipAddress string) (dto.RefreshResponse, error) {
	session, err := s.repo.GetSession(ctx, refreshToken)
	if err != nil || session == nil {
		return dto.RefreshResponse{}, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.DeleteSession(ctx, refreshToken)
		return dto.RefreshResponse{}, ErrSessionNotFound
	}

	user, isActive, err := s.repo.GetByID(ctx, session.UserID)
	if err != nil || !isActive {
		return dto.RefreshResponse{}, ErrUnauthorized
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return dto.RefreshResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken := uuid.New().String()
	newSession := &dto.Session{
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
	}

	err = s.txManager.WithinTx(ctx, func(r Repositories) error {
		if err := r.Auth().DeleteSession(ctx, refreshToken); err != nil {
			return err
		}
		return r.Auth().CreateSession(ctx, newSession)
	})
	if err != nil {
		return dto.RefreshResponse{}, fmt.Errorf("failed to rotate session: %w", err)
	}

	return dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID int64) (dto.UserInfo, error) {
	user, isActive, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return dto.UserInfo{}, err
	}
	if !isActive {
		return dto.UserInfo{}, ErrUserInactive
	}
	return *user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.txManager.WithinTx(ctx, func(r Repositories) error {
		if err := r.Auth().UpdatePassword(ctx, userID, string(hash)); err != nil {
			return err
		}
		return r.Auth().DeleteUserSessions(ctx, userID)
	})
}

func (s *AuthService) ValidateToken(tokenStr string) (int64, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", ErrUnauthorized
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return 0, "", ErrUnauthorized
	}

	return claims.UserID, claims.Role, nil
}

func (s *AuthService) generateAccessToken(userID int64, role string) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
