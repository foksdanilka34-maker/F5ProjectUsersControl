package auth

import (
	"context"
	"fmt"
	"time"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
	credential "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/auth"
	session "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/session"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/storage"
	"go.uber.org/zap"
)

type loginCore struct {
	sessionStorage    session.SessionStorage
	credentialStorage credential.CredentialStorage
	authenticator     session.Authenticator
}

type CoreLogic interface {
	Login(ctx context.Context, login, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, oldRefreshToken, userAgent, ipAddress string) (newAccessToken, newRefreshToken string, err error)
	CreateCredentials(ctx context.Context, userID, login, password, role string) error
	ChangeUserStatus(ctx context.Context, userID string, isActive bool) error
	ChangePassword(ctx context.Context, userID, newPassword string) error
}

func NewCore(credsStorage credential.CredentialStorage, sessStorage session.SessionStorage, auth session.Authenticator) CoreLogic {
	return &loginCore{
		credentialStorage: credsStorage,
		sessionStorage:    sessStorage,
		authenticator:     auth,
	}
}

func (l *loginCore) Login(ctx context.Context, login, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error) {
	storage.Logger.Info("Login called", zap.String("login", login), zap.String("userAgent", userAgent), zap.String("ipAddress", ipAddress))
	checkUser, err := l.credentialStorage.GetCrendentialsByUser(ctx, login)
	if err != nil {
		return "", "", err
	}
	passwordCheck := l.authenticator.CheckPasswordHash(password, checkUser.Password)
	if passwordCheck != nil {
		return "", "", err
	}
	accToken, refToken, err := l.authenticator.GenerateTokens(checkUser.UserID, checkUser.Role)
	if err != nil {
		return "", "", err
	}
	hashedRefToken := l.authenticator.HashRefreshToken(refToken)
	refreshSes := &auth.RefreshSession{
		UserID:       checkUser.UserID,
		RefreshToken: hashedRefToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(auth.RefreshTokenLifetime),
	}
	err = l.sessionStorage.CreateSession(ctx, refreshSes)
	if err != nil {
		return "", "", err
	}

	storage.Logger.Info("Login successful", zap.String("userID", checkUser.UserID), zap.String("login", login))
	return accToken, refToken, nil
}

func (l *loginCore) Logout(ctx context.Context, refreshToken string) error {
	storage.Logger.Info("Logout called", zap.String("refreshToken", refreshToken))
	if refreshToken == "" {
		storage.Logger.Warn("Logout: empty token")
		return fmt.Errorf("empty token")
	}
	tokenHash := l.authenticator.HashRefreshToken(refreshToken)
	err := l.sessionStorage.DeleteSession(ctx, tokenHash)
	if err != nil {
		return err
	}
	storage.Logger.Info("Logout successful", zap.String("refreshToken", refreshToken))
	return nil
}

func (l *loginCore) Refresh(ctx context.Context, oldRefreshToken, userAgent, ipAddress string) (newAccessToken, newRefreshToken string, err error) {
	storage.Logger.Info("Refresh called", zap.String("oldRefreshToken", oldRefreshToken), zap.String("userAgent", userAgent), zap.String("ipAddress", ipAddress))
	oldTokenHash := l.authenticator.HashRefreshToken(oldRefreshToken)
	ses, err := l.sessionStorage.GetSessionByToken(ctx, oldTokenHash)
	if err != nil {
		storage.Logger.Error("error getting session", zap.Error(err))
		return "", "", err
	}
	if time.Now().After(ses.ExpiresAt) {
		_ = l.sessionStorage.DeleteSession(ctx, oldRefreshToken)
		storage.Logger.Warn("token is expired", zap.String("userID", ses.UserID))
		return "", "", err
	}
	credentials, err := l.credentialStorage.GetCrendentialsByID(ctx, ses.UserID)
	if err != nil {
		storage.Logger.Error("error getting credentials", zap.Error(err))
		return "", "", err
	}
	accToken, refToken, err := l.authenticator.GenerateTokens(credentials.UserID, credentials.Role)
	if err != nil {
		storage.Logger.Error("error during generation token", zap.Error(err))
		return "", "", err
	}
	newHashToken := l.authenticator.HashRefreshToken(refToken)
	_, err = l.sessionStorage.UpdateSession(ctx, oldTokenHash, newHashToken, time.Now().Add(auth.RefreshTokenLifetime))
	if err != nil {
		storage.Logger.Error("error during updating session", zap.Error(err))
		return "", "", err
	}
	storage.Logger.Info("Refresh successful", zap.String("userID", credentials.UserID))
	return accToken, refToken, nil
}

func (l *loginCore) CreateCredentials(ctx context.Context, userID, login, password, role string) error {
	storage.Logger.Info("CreateCredentials called", zap.String("userID", userID), zap.String("login", login), zap.String("role", role))
	hashedPassword, err := l.authenticator.HashPassword(password)
	if err != nil {
		storage.Logger.Error("error hashing password", zap.Error(err))
		return err
	}
	credential := &auth.Credential{
		UserID:   userID,
		Login:    login,
		Password: hashedPassword,
		Role:     role,
	}
	err = l.credentialStorage.CreateCredentials(ctx, credential)
	if err != nil {
		storage.Logger.Error("error creating credentials", zap.Error(err))
		return err
	}
	storage.Logger.Info("CreateCredentials successful", zap.String("userID", userID), zap.String("login", login))
	return nil
}

func (l *loginCore) ChangeUserStatus(ctx context.Context, userID string, isActive bool) error {
	storage.Logger.Info("ChangeUserStatus called", zap.String("userID", userID), zap.Bool("isActive", isActive))
	err := l.credentialStorage.ChangeUserStatus(ctx, userID, isActive)
	if err != nil {
		storage.Logger.Error("error changing user status", zap.Error(err))
		return err
	}
	storage.Logger.Info("ChangeUserStatus successful", zap.String("userID", userID), zap.Bool("isActive", isActive))
	return nil
}

func (l *loginCore) ChangePassword(ctx context.Context, userID, newPassword string) error {
	storage.Logger.Info("ChangePassword called", zap.String("userID", userID))
	hashedPassword, err := l.authenticator.HashPassword(newPassword)
	if err != nil {
		storage.Logger.Error("error hashing password", zap.Error(err))
		return err
	}
	err = l.credentialStorage.PasswordHashUpdate(ctx, hashedPassword, userID)
	if err != nil {
		storage.Logger.Error("error updating password", zap.Error(err))
		return err
	}
	storage.Logger.Info("ChangePassword successful", zap.String("userID", userID))
	return nil
}
