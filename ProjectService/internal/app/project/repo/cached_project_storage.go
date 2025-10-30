package project

import (
	"context"
	"log"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
)

type CachedProjectStorage struct {
	db    ProjectStorage
	cache *CacheStorage
}

func NewCachedProjectStorage(db ProjectStorage, cache *CacheStorage) *CachedProjectStorage {
	return &CachedProjectStorage{
		db:    db,
		cache: cache,
	}
}

func (c *CachedProjectStorage) CreateProject(ctx context.Context, createProject *models.CreateProjectRequest) (*models.Project, error) {
	return c.db.CreateProject(ctx, createProject)
}

func (c *CachedProjectStorage) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	project, err := c.cache.Get(ctx, projectID)
	if err != nil {
		project, err = c.db.GetProject(ctx, projectID)
		if err == nil {
			_ = c.cache.Set(ctx, project)
		}
	}
	return project, err
}

func (c *CachedProjectStorage) ListProjects(ctx context.Context, listProject *models.ListProjectsFilter) (*models.ProjectsListResponse, error) {
	return c.db.ListProjects(ctx, listProject)
}

func (c *CachedProjectStorage) UpdateProject(ctx context.Context, updRequest *models.UpdateProjectRequest) (*models.Project, error) {
	updProject, err := c.db.UpdateProject(ctx, updRequest)
	if err != nil {
		log.Printf("error updatingProject %v", err)
		return nil, err
	}
	if updProject != nil {
		_ = c.cache.Set(ctx, updProject)
	}
	return updProject, nil
}