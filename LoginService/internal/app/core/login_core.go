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
	log.Printf("Login called: login=%s, userAgent=%s, ipAddress=%s", login, userAgent, ipAddress)
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

	log.Printf("Login successful: userID=%s, login=%s", checkUser.UserID, login)
	return accToken, refToken, nil
}

func (l *loginCore) Logout(ctx context.Context, refreshToken string) error {
	log.Printf("Logout called: refreshToken=%s", refreshToken)
	if refreshToken == "" {
		log.Printf("Logout: empty token")
		return fmt.Errorf("empty token")
	}
	tokenHash := l.authenticator.HashRefreshToken(refreshToken)
	err := l.sessionStorage.DeleteSession(ctx, tokenHash)
	if err != nil {
		return err
	}
	log.Printf("Logout successful: refreshToken=%s", refreshToken)
	return nil
}

func (l *loginCore) Refresh(ctx context.Context, oldRefreshToken, userAgent, ipAddress string) (newAccessToken, newRefreshToken string, err error) {
	log.Printf("Refresh called: oldRefreshToken=%s, userAgent=%s, ipAddress=%s", oldRefreshToken, userAgent, ipAddress)
	oldTokenHash := l.authenticator.HashRefreshToken(oldRefreshToken)
	ses, err := l.sessionStorage.GetSessionByToken(ctx, oldTokenHash)
	if err != nil {
		log.Printf("error getting session: %v", err)
		return "", "", err
	}
	if time.Now().After(ses.ExpiresAt) {
		_ = l.sessionStorage.DeleteSession(ctx, oldRefreshToken)
		log.Printf("token is expired: userID=%s", ses.UserID)
		return "", "", err
	}
	credentials, err := l.credentialStorage.GetCrendentialsByID(ctx, ses.UserID)
	if err != nil {
		log.Printf("error getting credentials: %v", err)
		return "", "", err
	}
	accToken, refToken, err := l.authenticator.GenerateTokens(credentials.UserID, credentials.Role)
	if err != nil {
		log.Printf("error during generation token: %v", err)
		return "", "", err
	}
	newHashToken := l.authenticator.HashRefreshToken(refToken)
	_, err = l.sessionStorage.UpdateSession(ctx, oldTokenHash, newHashToken, time.Now().Add(auth.RefreshTokenLifetime))
	if err != nil {
		log.Printf("error during updating session: %v", err)
		return "", "", err
	}
	log.Printf("Refresh successful: userID=%s", credentials.UserID)
	return accToken, refToken, nil
}

func (l *loginCore) CreateCredentials(ctx context.Context, userID, login, password, role string) error {
	log.Printf("CreateCredentials called: userID=%s, login=%s, role=%s", userID, login, role)
	hashedPassword, err := l.authenticator.HashPassword(password)
	if err != nil {
		log.Printf("error hashing password: %v", err)
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
		log.Printf("error creating credentials: %v", err)
		return err
	}
	log.Printf("CreateCredentials successful: userID=%s, login=%s", userID, login)
	return nil
}

func (l *loginCore) ChangeUserStatus(ctx context.Context, userID string, isActive bool) error {
	log.Printf("ChangeUserStatus called: userID=%s, isActive=%t", userID, isActive)
	err := l.credentialStorage.ChangeUserStatus(ctx, userID, isActive)
	if err != nil {
		log.Printf("error changing user status: %v", err)
		return err
	}
	log.Printf("ChangeUserStatus successful: userID=%s, isActive=%t", userID, isActive)
	return nil
}

func (l *loginCore) ChangePassword(ctx context.Context, userID, newPassword string) error {
	log.Printf("ChangePassword called: userID=%s", userID)
	hashedPassword, err := l.authenticator.HashPassword(newPassword)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return err
	}
	err = l.credentialStorage.PasswordHashUpdate(ctx, hashedPassword, userID)
	if err != nil {
		log.Printf("error updating password: %v", err)
		return err
	}
	log.Printf("ChangePassword successful: userID=%s", userID)
	return nil
}
