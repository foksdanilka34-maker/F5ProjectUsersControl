package project

import "github.com/redis/go-redis/v9"

type CacheStorage struct {
	r *redis.Client
}

func NewCacheStorage(r *redis.Client) *CacheStorage {
	return &CacheStorage{
		r: r,
	}
}

type ProjectCacheStorage interface {
	
}
