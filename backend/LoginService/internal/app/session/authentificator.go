package session

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"crypto/rand"
	"crypto/sha256"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator interface {
	GenerateTokens(userID, role string) (accessToken, refreshToken string, err error)
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hashedPassword string) error
	HashRefreshToken(token string) string
}

type JWTAuthenticator struct {
	jwtSecretKey []byte
}

func NewAuthenticator(secretKey string) (Authenticator, error) {
	if len(secretKey) < auth.RefreshTokenLength {
		return nil, fmt.Errorf("jwt secret key must be at least 32 characters long")
	}
	return &JWTAuthenticator{
		jwtSecretKey: []byte(secretKey),
	}, nil
}

func (a *JWTAuthenticator) GenerateTokens(userID, role string) (accessToken, refreshToken string, err error) {
	customClaims := &auth.CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auth.AccessTokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	accessToken, err = token.SignedString(a.jwtSecretKey)
	if err != nil {
		log.Printf("error during signing token %v", err)
		return "", "", err
	}

	refreshTokenBytes := make([]byte, auth.RefreshTokenLength)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken = base64.URLEncoding.EncodeToString(refreshTokenBytes)

	return accessToken, refreshToken, nil
}

func (a *JWTAuthenticator) HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", fmt.Errorf("password is not secure")
	}

	val, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func (a *JWTAuthenticator) CheckPasswordHash(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (a *JWTAuthenticator) HashRefreshToken(token string) string {
	hash := sha256.New()
	hash.Write([]byte(token))

	return fmt.Sprintf("%x", hash.Sum(nil))
}
