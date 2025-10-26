package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/storage"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type StorageCache struct {
	redis *redis.Client
}

func NewRedisCache(r *redis.Client) *StorageCache {
	return &StorageCache{
		redis: r,
	}
}

type SessionCache interface {
	Set(ctx context.Context, session *auth.RefreshSession) error
	Get(ctx context.Context, refreshToken string) (*auth.RefreshSession, error)
	Delete(ctx context.Context, refreshToken string) error
}

func buildKey(token string) string {
	return auth.SessionKeyPrefix + token
}

func (r *StorageCache) Set(ctx context.Context, session *auth.RefreshSession) error {
	payload, err := json.Marshal(session)
	if err != nil {
		storage.Logger.Error("error in setting cache value, marshal error", zap.Error(err))
		return err
	}
	ttl := time.Until(session.ExpiresAt)
	if ttl < time.Second {
		return fmt.Errorf("error, session expired")
	}
	return r.redis.Set(ctx, buildKey(session.RefreshToken), payload, ttl).Err()
}

func (r *StorageCache) Get(ctx context.Context, refreshToken string) (*auth.RefreshSession, error) {
	keyBuild := buildKey(refreshToken)
	data, err := r.redis.Get(ctx, keyBuild).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("key not exist")
		}
		storage.Logger.Error("error during getting key", zap.Error(err))
		return nil, err
	}
	session := &auth.RefreshSession{}

	if err = json.Unmarshal([]byte(data), session); err != nil {
		r.redis.Del(ctx, keyBuild)
		return nil, err
	}
	return session, nil
}

func (r *StorageCache) Delete(ctx context.Context, refreshToken string) error {
	return r.redis.Del(ctx, buildKey(refreshToken)).Err()
}
