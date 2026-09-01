package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
)

type ProjectRepository interface {
	Create(ctx context.Context, p *dto.ProjectDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.ProjectDTO, error)
	List(ctx context.Context, filter dto.ListProjectsFilter) ([]dto.ProjectDTO, int, error)
	Update(ctx context.Context, id int64, req dto.UpdateProjectRequest) error
	Delete(ctx context.Context, id int64) error

	AddMember(ctx context.Context, projectID, userID int64, role string) error
	RemoveMember(ctx context.Context, projectID, userID int64) error
	GetMembers(ctx context.Context, projectID int64) ([]dto.ProjectMemberDTO, error)
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(r *repo.RepositoryRegistry) error) error
}

type ProjectService struct {
	repo      ProjectRepository
	txManager TxManager
}

func NewProjectService(repo ProjectRepository, txManager TxManager) *ProjectService {
	return &ProjectService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, req dto.CreateProjectRequest) (dto.ProjectDTO, error) {
	if req.Name == "" || req.ManagerID == 0 {
		return dto.ProjectDTO{}, errors.New("name and manager_id are required")
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			dueDate = &t
		}
	}

	project := &dto.ProjectDTO{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerID,
		Status:      "ACTIVE",
		DueDate:     dueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var id int64
	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		var err error
		id, err = r.Project().Create(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		project.ID = id

		if err := r.Project().AddMember(ctx, id, req.ManagerID, "manager"); err != nil {
			return err
		}

		// Insert Outbox Event
		eventPayload := dto.ProjectEventPayload{
			EventID:   uuid.New().String(),
			ProjectID: id,
			Name:      project.Name,
			Status:    project.Status,
			Timestamp: time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err = r.Outbox().Insert(ctx, "project.event.created", payloadBytes)
		return err
	})

	if err != nil {
		return dto.ProjectDTO{}, err
	}

	return *project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id int64) (dto.ProjectDTO, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.ProjectDTO{}, err
	}
	if p == nil {
		return dto.ProjectDTO{}, ErrNotFound
	}
	return *p, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, filter dto.ListProjectsFilter) (dto.ListProjectsResponse, error) {
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageNumber <= 0 {
		filter.PageNumber = 1
	}

	projects, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return dto.ListProjectsResponse{}, err
	}

	return dto.ListProjectsResponse{
		Projects:   projects,
		TotalCount: total,
	}, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, id int64, req dto.UpdateProjectRequest) (dto.ProjectDTO, error) {
	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if err := r.Project().Update(ctx, id, req); err != nil {
			return err
		}

		eventPayload := dto.ProjectEventPayload{
			EventID:   uuid.New().String(),
			ProjectID: id,
			Timestamp: time.Now(),
		}
		if req.Name != nil {
			eventPayload.Name = *req.Name
		}
		if req.Status != nil {
			eventPayload.Status = *req.Status
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "project.event.updated", payloadBytes)
		return err
	})

	if err != nil {
		return dto.ProjectDTO{}, err
	}

	return s.GetProject(ctx, id)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id int64) error {
	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if err := r.Project().Delete(ctx, id); err != nil {
			return err
		}
		eventPayload := dto.ProjectEventPayload{
			EventID:   uuid.New().String(),
			ProjectID: id,
			Timestamp: time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "project.event.deleted", payloadBytes)
		return err
	})
}

func (s *ProjectService) AddMember(ctx context.Context, projectID, userID int64, role string) error {
	if role == "" {
		role = "member"
	}
	return s.repo.AddMember(ctx, projectID, userID, role)
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID int64) error {
	return s.repo.RemoveMember(ctx, projectID, userID)
}

func (s *ProjectService) ListMembers(ctx context.Context, projectID int64) ([]dto.ProjectMemberDTO, error) {
	return s.repo.GetMembers(ctx, projectID)
}
