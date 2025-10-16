package app

import (
	"context"
	"fmt"
	"log"
	"time"
)

type loginCore struct {
	sessionStorage    SessionStorage
	credentialStorage CredentialStorage
	authenticator     Authenticator
}

type CoreLogic interface {
	Login(ctx context.Context, login, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, oldRefreshToken, userAgent, ipAddress string) (newAccessToken, newRefreshToken string, err error)
	CreateCredentials(ctx context.Context, userID, login, password, role string) error
}

func NewCore(credsStorage CredentialStorage, sessStorage SessionStorage, auth Authenticator) CoreLogic {
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
	refreshSes := &RefreshSession{
		UserID:       checkUser.UserID,
		RefreshToken: hashedRefToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(RefreshTokenLifetime),
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
		return "", "", err
	}
	if time.Now().After(ses.ExpiresAt) {
		_ = l.sessionStorage.DeleteSession(ctx, oldRefreshToken)
		return "", "", err
	}
	credentials, err := l.credentialStorage.GetCrendentialsByID(ctx, ses.UserID)
	if err != nil {
		return "", "", err
	}
	accToken, refToken, err := l.authenticator.GenerateTokens(credentials.UserID, credentials.Role)
	if err != nil {
		return "", "", err
	}
	newHashToken := l.authenticator.HashRefreshToken(refToken)
	_, err = l.sessionStorage.UpdateSession(ctx, oldTokenHash, newHashToken, time.Now().Add(RefreshTokenLifetime))
	if err != nil {
		return "", "", err
	}
	return accToken, refToken, nil
}

func (l *loginCore) CreateCredentials(ctx context.Context, userID, login, password, role string) error {
	hashedPassword, err := l.authenticator.HashPassword(password)
	if err != nil {
		return err
	}
	credential := &Credential{
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
