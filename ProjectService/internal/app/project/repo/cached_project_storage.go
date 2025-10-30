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

func (c *CachedEmployeeStorage) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	project, err := c.cache.Get(ctx, projectID)
	if err != nil {
		project, err = c.db.GetProject(ctx, projectID)
		if err == nil {
			_ = c.cache.Set(ctx, project)
		}
	}
	return project, err
}
