package core

import (
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/repo"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTAuthenticator struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTAuthenticator(secret string, accessTTL, refreshTTL time.Duration) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func NewAuthenticator(secret string, accessTTL, refreshTTL time.Duration) *JWTAuthenticator {
	return NewJWTAuthenticator(secret, accessTTL, refreshTTL)
}

func (a *JWTAuthenticator) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (a *JWTAuthenticator) CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (a *JWTAuthenticator) GenerateTokens(userID int64, role string) (string, string, error) {
	now := time.Now()

	accessClaims := &repo.CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(a.secret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &repo.CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(a.secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (a *JWTAuthenticator) ValidateRefreshToken(tokenString string) (*repo.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &repo.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*repo.CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}


