package core

import (
	"context"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

// ProjectRepository - интерфейс репозитория проектов
type ProjectRepository interface {
	Create(ctx context.Context, p *repo.Project) (int64, error)
	GetByID(ctx context.Context, id int64) (*repo.Project, error)
	List(ctx context.Context, pageSize, offset int, status string, ownerID int64, memberID int64) ([]*repo.Project, int, error)
	Update(ctx context.Context, id int64, name, description, status *string, startDate, endDate *time.Time) error
	Delete(ctx context.Context, id int64) error
	GetUserProjects(ctx context.Context, userID int64) ([]*repo.Project, error)
	ExistsByName(ctx context.Context, name string) (bool, error)

	AddMember(ctx context.Context, projectID, userID int64, role string) error
	RemoveMember(ctx context.Context, projectID, userID int64) error
	GetMembers(ctx context.Context, projectID int64) ([]*repo.ProjectMember, error)
	IsMember(ctx context.Context, projectID, userID int64) (bool, error)
	GetTaskStats(ctx context.Context, projectID int64) (*repo.TaskStats, error)
}

// ProjectService - сервис проектов
type ProjectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

type CreateProjectRequest struct {
	Name        string
	Description string
	OwnerID     int64
	StartDate   *time.Time
	EndDate     *time.Time
}

func (s *ProjectService) CreateProject(ctx context.Context, req *CreateProjectRequest) (*repo.Project, error) {
	// Проверка уникальности имени
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProjectNameExists
	}

	project := &repo.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		OwnerID:     req.OwnerID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	id, err := s.repo.Create(ctx, project)
	if err != nil {
		return nil, err
	}
	project.ID = id

	// Add owner as member
	_ = s.repo.AddMember(ctx, id, req.OwnerID, "owner")

	return project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id int64) (*repo.Project, error) {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrNotFound
	}

	stats, _ := s.repo.GetTaskStats(ctx, id)
	project.TaskStats = stats

	return project, nil
}

type ListProjectsFilter struct {
	PageSize   int
	PageNumber int
	Status     string
	OwnerID    int64
	MemberID   int64
}

func (s *ProjectService) ListProjects(ctx context.Context, filter *ListProjectsFilter) ([]*repo.Project, int, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	pageNumber := filter.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	offset := (pageNumber - 1) * pageSize

	return s.repo.List(ctx, pageSize, offset, filter.Status, filter.OwnerID, filter.MemberID)
}

func (s *ProjectService) GetUserProjects(ctx context.Context, userID int64) ([]*repo.Project, error) {
	return s.repo.GetUserProjects(ctx, userID)
}

type UpdateProjectRequest struct {
	Name        *string
	Description *string
	Status      *string
	StartDate   *time.Time
	EndDate     *time.Time
}

func (s *ProjectService) UpdateProject(ctx context.Context, id int64, req *UpdateProjectRequest) (*repo.Project, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if err := s.repo.Update(ctx, id, req.Name, req.Description, req.Status, req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
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

func (s *ProjectService) GetMembers(ctx context.Context, projectID int64) ([]*repo.ProjectMember, error) {
	return s.repo.GetMembers(ctx, projectID)
}

func (s *ProjectService) CheckAccess(ctx context.Context, projectID, userID int64) (bool, error) {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, ErrNotFound
	}

	if project.OwnerID == userID {
		return true, nil
	}

	return s.repo.IsMember(ctx, projectID, userID)
}
