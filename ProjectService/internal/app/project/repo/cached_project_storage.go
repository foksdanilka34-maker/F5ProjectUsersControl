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

func (c *CachedProjectStorage) DeleteProject(ctx context.Context, projectID string) error {
	err := c.db.DeleteProject(ctx, projectID)
	if err != nil {
		log.Printf("error deleting project from db: %v", err)
		return err
	}

	err = c.cache.Delete(ctx, projectID)
	if err != nil {
		log.Printf("error deleting project from cache, but deleted from db: %v", err)
	}
	return nil
}

func (c *CachedProjectStorage) CreateTask(ctx context.Context, createTask *models.CreateTaskRequest) (*models.Task, error) {
	return c.db.CreateTask(ctx, createTask)
}

func (c *CachedProjectStorage) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	return c.db.GetTask(ctx, taskID)
}

func (c *CachedProjectStorage) UpdateTask(ctx context.Context, updRequest *models.UpdateTaskRequest) (*models.Task, error) {
	return c.db.UpdateTask(ctx, updRequest)
}

func (c *CachedProjectStorage) DeleteTask(ctx context.Context, taskID string) error {
	return c.db.DeleteTask(ctx, taskID)
}

func (c *CachedProjectStorage) MoveTask(ctx context.Context, moveRequest *models.MoveTaskRequest) (*models.Task, error) {
	return c.db.MoveTask(ctx, moveRequest)
}

func (c *CachedProjectStorage) AssignTask(ctx context.Context, assignRequest *models.AssignTaskRequest) (*models.Task, error) {
	return c.db.AssignTask(ctx, assignRequest)
}

func (c *CachedProjectStorage) ListTasksByProject(ctx context.Context, filter *models.ListTasksFilter) (*models.TasksListResponse, error) {
	return c.db.ListTasksByProject(ctx, filter)
}

func (c *CachedProjectStorage) AddMemberToProject(ctx context.Context, projectID, userID string) error {
	return c.db.AddMemberToProject(ctx, projectID, userID)
}

func (c *CachedProjectStorage) RemoveMemberFromProject(ctx context.Context, projectID, userID string) error {
	return c.db.RemoveMemberFromProject(ctx, projectID, userID)
}

func (c *CachedProjectStorage) ListProjectMembers(ctx context.Context, projectID string) (*models.ProjectMembersResponse, error) {
	return c.db.ListProjectMembers(ctx, projectID)
}

func (c *CachedProjectStorage) RecordStatusChange(ctx context.Context, taskID string, fromStatus, toStatus models.TaskStatus, actorID *string) error {
	return c.db.RecordStatusChange(ctx, taskID, fromStatus, toStatus, actorID)
}

func (c *CachedProjectStorage) GetTaskStatusHistory(ctx context.Context, taskID string, pageSize, pageNumber int32) (*models.TaskStatusHistoryResponse, error) {
	return c.db.GetTaskStatusHistory(ctx, taskID, pageSize, pageNumber)
}

func (c *CachedProjectStorage) GetProjectMetrics(ctx context.Context, projectID string) (*models.ProjectMetrics, error) {
	return c.db.GetProjectMetrics(ctx, projectID)
}
