package session

import (
	"context"
	"log"
	"time"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app"
)

type CachedSessionStorage struct {
	db    SessionStorage
	cache SessionCache
}

func NewCachedSessionStorage(db SessionStorage, cache SessionCache) *CachedSessionStorage {
	return &CachedSessionStorage{
		db:    db,
		cache: cache,
	}
}

func (s *CachedSessionStorage) GetSessionByToken(ctx context.Context, tokenHash string) (*auth.RefreshSession, error) {
	session, err := s.cache.Get(ctx, tokenHash)
	if err == nil {
		return session, nil
	}

	sessionFromDB, err := s.db.GetSessionByToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, sessionFromDB)

	return sessionFromDB, nil
}

func (s *CachedSessionStorage) CreateSession(ctx context.Context, session *auth.RefreshSession) error {
	err := s.db.CreateSession(ctx, session)
	if err == nil {
		_ = s.cache.Set(ctx, session)
		return nil
	}
	return err
}

func (s *CachedSessionStorage) UpdateSession(ctx context.Context, oldTokenHash string, newTokenHash string, newExpiresAt time.Time) (*auth.RefreshSession, error) {
	updatedSession, err := s.db.UpdateSession(ctx, oldTokenHash, newTokenHash, newExpiresAt)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, oldTokenHash)
	_ = s.cache.Set(ctx, updatedSession)

	return updatedSession, nil
}
func (s *CachedSessionStorage) DeleteSession(ctx context.Context, token string) error {
	log.Printf("using deleteSession func")
	err := s.db.DeleteSession(ctx, token)
	if err != nil {
		log.Printf("error deleting session from db: %v", err)
		return err
	}

	cacheErr := s.cache.Delete(ctx, token)
	if cacheErr != nil {
		log.Printf("error deleting session from cache: %v", cacheErr)
	}

	log.Printf("session deleted successfully from both db and cache")
	return nil
}
