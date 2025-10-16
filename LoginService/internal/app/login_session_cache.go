package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionCache interface {
	Set(ctx context.Context, session *RefreshSession) error
	Get(ctx context.Context, refreshToken string) (*RefreshSession, error)
	Delete(ctx context.Context, refreshToken string) error
}

func buildKey(token string) string {
	return sessionKeyPrefix + token
}

func (r *StorageCache) Set(ctx context.Context, session *RefreshSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		log.Printf("error in setting cache value, marshal error %v", err)
		return err
	}
	ttl := time.Until(session.ExpiresAt)
	if ttl < time.Second {
		return fmt.Errorf("error, session expired")
	}
	return r.redis.Set(ctx, buildKey(session.RefreshToken), data, ttl).Err()
}

func (r *StorageCache) Get(ctx context.Context, refreshToken string) (*RefreshSession, error) {
	keyBuild := buildKey(refreshToken)
	data, err := r.redis.Get(ctx, keyBuild).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("key not exist")
		}
		log.Printf("error during getting key, %v", err)
		return nil, err
	}
	session := &RefreshSession{}

	if err = json.Unmarshal([]byte(data), session); err != nil {
		r.redis.Del(ctx, keyBuild)
		return nil, err
	}
	return session, nil
}

func (r *StorageCache) Delete(ctx context.Context, refreshToken string) error {
	return r.redis.Del(ctx, buildKey(refreshToken)).Err()
}
