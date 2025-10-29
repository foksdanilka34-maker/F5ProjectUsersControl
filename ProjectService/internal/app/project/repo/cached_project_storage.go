package project

import (

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	"context"
)

type CachedEmployeeStorage struct {
	db    ProjectStorage
	cache *CacheStorage
}

func NewCachedProjectStorage(db ProjectStorage, cache *CacheStorage) *CachedEmployeeStorage {
	return &CachedEmployeeStorage{
		db:    db,
		cache: cache,
	}
}

func (c *CachedEmployeeStorage) CreateProject(ctx context.Context, createProject *models.CreateProjectRequest) (*models.Project, error) {
	return c.db.CreateProject(ctx, createProject)
}
