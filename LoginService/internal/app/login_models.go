package app

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix     = "session:"
	AccessTokenLifetime  = 15 * time.Minute
	RefreshTokenLifetime = 3 * 24 * time.Hour
	RefreshTokenLength   = 32
)

type Credential struct {
	UserID    string
	Login     string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

type RefreshSession struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Storage struct {
	pgx *pgxpool.Pool
}

type StorageCache struct {
	redis *redis.Client
}

func NewCacheStorage(r *redis.Client) *StorageCache {
	return &StorageCache{
		redis: r,
	}
}

func NewStorage(p *pgxpool.Pool) *Storage {
	return &Storage{
		pgx: p,
	}
}
