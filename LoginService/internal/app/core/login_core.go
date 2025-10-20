package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
	credential "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/auth"
	session "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/session"
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

	return accToken, refToken, nil
}

func (l *loginCore) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		log.Printf("Empty token")
		return fmt.Errorf("empty token")
	}
	tokenHash := l.authenticator.HashRefreshToken(refreshToken)
	return l.sessionStorage.DeleteSession(ctx, tokenHash)
}

func (l *loginCore) Refresh(ctx context.Context, oldRefreshToken, userAgent, ipAddress string) (newAccessToken, newRefreshToken string, err error) {
	oldTokenHash := l.authenticator.HashRefreshToken(oldRefreshToken)
	ses, err := l.sessionStorage.GetSessionByToken(ctx, oldTokenHash)
	if err != nil {
		log.Printf("error getting session")
		return "", "", err
	}
	if time.Now().After(ses.ExpiresAt) {
		_ = l.sessionStorage.DeleteSession(ctx, oldRefreshToken)
		log.Printf("token is expired")
		return "", "", err
	}
	credentials, err := l.credentialStorage.GetCrendentialsByID(ctx, ses.UserID)
	if err != nil {
		log.Printf("error getting credentials")
		return "", "", err
	}
	accToken, refToken, err := l.authenticator.GenerateTokens(credentials.UserID, credentials.Role)
	if err != nil {
		log.Printf("error during generation token")
		return "", "", err
	}
	newHashToken := l.authenticator.HashRefreshToken(refToken)
	_, err = l.sessionStorage.UpdateSession(ctx, oldTokenHash, newHashToken, time.Now().Add(auth.RefreshTokenLifetime))
	if err != nil {
		log.Printf("error during updating session")
		return "", "", err
	}
	return accToken, refToken, nil
}

func (l *loginCore) CreateCredentials(ctx context.Context, userID, login, password, role string) error {
	hashedPassword, err := l.authenticator.HashPassword(password)
	if err != nil {
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
		return err
	}
	return nil
}

func (l *loginCore) ChangeUserStatus(ctx context.Context, userID string, isActive bool) error {
	return l.credentialStorage.ChangeUserStatus(ctx, userID, isActive)
}

func (l *loginCore) ChangePassword(ctx context.Context, userID, newPassword string) error {
	hashedPassword, err := l.authenticator.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return l.credentialStorage.PasswordHashUpdate(ctx, hashedPassword, userID)
}
