package core

import (
	"context"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/repo"
	"github.com/jackc/pgx/v5"
)

type ProfileRepository interface {
	Create(ctx context.Context, tx pgx.Tx, p *repo.Profile) error
	GetByID(ctx context.Context, id int64) (*repo.Profile, error)
	List(ctx context.Context, pageSize, offset int, departmentID, positionID int64) ([]*repo.Profile, int, error)
	Update(ctx context.Context, id int64, firstName, lastName *string, positionID, departmentID *int64, email, avatarURL *string) error

	CreateDepartment(ctx context.Context, name string) (int64, error)
	GetDepartment(ctx context.Context, id int64) (*repo.Department, error)
	ListDepartments(ctx context.Context) ([]*repo.Department, error)
	UpdateDepartment(ctx context.Context, id int64, name string) error
	DeleteDepartment(ctx context.Context, id int64) error

	CreatePosition(ctx context.Context, name string) (int64, error)
	GetPosition(ctx context.Context, id int64) (*repo.Position, error)
	ListPositions(ctx context.Context) ([]*repo.Position, error)
	UpdatePosition(ctx context.Context, id int64, name string) error
	DeletePosition(ctx context.Context, id int64) error

	CreateSkill(ctx context.Context, name string) (int64, error)
	ListSkills(ctx context.Context) ([]repo.Skill, error)
	DeleteSkill(ctx context.Context, id int64) error
	AddSkillToProfile(ctx context.Context, profileID, skillID int64) error
	RemoveSkillFromProfile(ctx context.Context, profileID, skillID int64) error

	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type EmployeeEventPublisher interface {
	PublishEmployeeCreated(ctx context.Context, userID int64, fullName string) error
	PublishEmployeeUpdated(ctx context.Context, userID int64, fullName string) error
	PublishEmployeeDeleted(ctx context.Context, userID int64) error
}

type ProfileService struct {
	repo        ProfileRepository
	authService *AuthService
	publisher   EmployeeEventPublisher
}

func NewProfileService(repo ProfileRepository, authService *AuthService, publisher EmployeeEventPublisher) *ProfileService {
	return &ProfileService{
		repo:        repo,
		authService: authService,
		publisher:   publisher,
	}
}

type CreateProfileRequest struct {
	FirstName    string
	LastName     string
	PositionID   int64
	DepartmentID *int64
	Email        string
	Login        string
	Password     string
	Role         string
	HireDate     time.Time
}

func (s *ProfileService) CreateProfile(ctx context.Context, req *CreateProfileRequest) (*repo.Profile, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	role := req.Role
	if role == "" {
		role = "employee"
	}

	userID, err := s.authService.CreateCredentials(ctx, tx, req.Login, req.Password, role)
	if err != nil {
		return nil, err
	}

	hireDate := req.HireDate
	if hireDate.IsZero() {
		hireDate = time.Now()
	}

	profile := &repo.Profile{
		ID:           userID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PositionID:   req.PositionID,
		DepartmentID: req.DepartmentID,
		Email:        req.Email,
		HireDate:     hireDate,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, tx, profile); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	profile.Login = req.Login
	profile.Role = role
	profile.IsActive = true

	if s.publisher != nil {
		fullName := profile.FirstName + " " + profile.LastName
		_ = s.publisher.PublishEmployeeCreated(ctx, userID, fullName)
	}

	return profile, nil
}

func (s *ProfileService) GetProfile(ctx context.Context, id int64) (*repo.Profile, error) {
	return s.repo.GetByID(ctx, id)
}

type ListProfilesFilter struct {
	PageSize     int
	PageNumber   int
	DepartmentID int64
	PositionID   int64
}

func (s *ProfileService) ListProfiles(ctx context.Context, filter *ListProfilesFilter) ([]*repo.Profile, int, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	pageNumber := filter.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	offset := (pageNumber - 1) * pageSize

	return s.repo.List(ctx, pageSize, offset, filter.DepartmentID, filter.PositionID)
}

type UpdateProfileRequest struct {
	FirstName    *string
	LastName     *string
	PositionID   *int64
	DepartmentID *int64
	Email        *string
	AvatarURL    *string
}

func (s *ProfileService) UpdateProfile(ctx context.Context, id int64, req *UpdateProfileRequest) (*repo.Profile, error) {
	if err := s.repo.Update(ctx, id, req.FirstName, req.LastName, req.PositionID, req.DepartmentID, req.Email, req.AvatarURL); err != nil {
		return nil, err
	}

	profile, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.publisher != nil {
		fullName := profile.FirstName + " " + profile.LastName
		_ = s.publisher.PublishEmployeeUpdated(ctx, id, fullName)
	}

	return profile, nil
}

func (s *ProfileService) ChangeUserStatus(ctx context.Context, userID int64, isActive bool) error {
	return s.authService.ChangeStatus(ctx, userID, isActive)
}

func (s *ProfileService) CreateDepartment(ctx context.Context, name string) (*repo.Department, error) {
	id, err := s.repo.CreateDepartment(ctx, name)
	if err != nil {
		return nil, err
	}
	return &repo.Department{ID: id, Name: name}, nil
}

func (s *ProfileService) GetDepartment(ctx context.Context, id int64) (*repo.Department, error) {
	return s.repo.GetDepartment(ctx, id)
}

func (s *ProfileService) ListDepartments(ctx context.Context) ([]*repo.Department, error) {
	return s.repo.ListDepartments(ctx)
}

func (s *ProfileService) UpdateDepartment(ctx context.Context, id int64, name string) (*repo.Department, error) {
	if err := s.repo.UpdateDepartment(ctx, id, name); err != nil {
		return nil, err
	}
	return &repo.Department{ID: id, Name: name}, nil
}

func (s *ProfileService) DeleteDepartment(ctx context.Context, id int64) error {
	return s.repo.DeleteDepartment(ctx, id)
}

func (s *ProfileService) CreatePosition(ctx context.Context, name string) (*repo.Position, error) {
	id, err := s.repo.CreatePosition(ctx, name)
	if err != nil {
		return nil, err
	}
	return &repo.Position{ID: id, Name: name}, nil
}

func (s *ProfileService) GetPosition(ctx context.Context, id int64) (*repo.Position, error) {
	return s.repo.GetPosition(ctx, id)
}

func (s *ProfileService) ListPositions(ctx context.Context) ([]*repo.Position, error) {
	return s.repo.ListPositions(ctx)
}

func (s *ProfileService) UpdatePosition(ctx context.Context, id int64, name string) (*repo.Position, error) {
	if err := s.repo.UpdatePosition(ctx, id, name); err != nil {
		return nil, err
	}
	return &repo.Position{ID: id, Name: name}, nil
}

func (s *ProfileService) DeletePosition(ctx context.Context, id int64) error {
	return s.repo.DeletePosition(ctx, id)
}

func (s *ProfileService) CreateSkill(ctx context.Context, name string) (*repo.Skill, error) {
	id, err := s.repo.CreateSkill(ctx, name)
	if err != nil {
		return nil, err
	}
	return &repo.Skill{ID: id, Name: name}, nil
}

func (s *ProfileService) ListSkills(ctx context.Context) ([]repo.Skill, error) {
	return s.repo.ListSkills(ctx)
}

func (s *ProfileService) DeleteSkill(ctx context.Context, id int64) error {
	return s.repo.DeleteSkill(ctx, id)
}

func (s *ProfileService) AddSkillToProfile(ctx context.Context, profileID, skillID int64) error {
	return s.repo.AddSkillToProfile(ctx, profileID, skillID)
}

func (s *ProfileService) RemoveSkillFromProfile(ctx context.Context, profileID, skillID int64) error {
	return s.repo.RemoveSkillFromProfile(ctx, profileID, skillID)
}


