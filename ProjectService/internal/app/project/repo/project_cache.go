package project

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	"github.com/redis/go-redis/v9"
)

type CacheStorage struct {
	r *redis.Client
}

const (
	projectKey = "project:"
)

func NewCacheStorage(r *redis.Client) *CacheStorage {
	return &CacheStorage{
		r: r,
	}
}

func buildKey(projectID string) string {
	return projectKey + projectID
}

type ProjectCacheStorage interface {
	Set(ctx context.Context, project *models.Project) error
	Get(ctx context.Context, projectID string) (project *models.Project, err error)
}

func (c *CacheStorage) Set(ctx context.Context, project *models.Project) error {
	payload, err := json.Marshal(project)
	if err != nil {
		log.Printf("error marshaling object: %v", err)
		return err
	}
	return c.r.Set(ctx, projectKey+project.ID, payload, 0).Err()
}

func (c *CacheStorage) Get(ctx context.Context, projectID string) (project *models.Project, err error) {
	getProject, err := c.r.Get(ctx, buildKey(projectID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Printf("project not found in cache")
			return nil, err
		}
		log.Printf("system error: %v", err)
		return nil, err
	}

	retProject := &models.Project{}
	if unm := json.Unmarshal([]byte(getProject), retProject); unm != nil {
		c.r.Del(ctx, buildKey(projectID))
		log.Printf("error during unmarshaling: %v", err)
		return nil, err
	}
	return retProject, nil
}
