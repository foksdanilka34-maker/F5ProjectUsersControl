package core

import (
	"context"
	"errors"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
)

var (
	ErrNotFound = errors.New("not found")
)

type OrgRepository interface {
	CreateDepartment(ctx context.Context, name string) (*dto.DepartmentDTO, error)
	GetDepartment(ctx context.Context, id int64) (*dto.DepartmentDTO, error)
	ListDepartments(ctx context.Context) ([]dto.DepartmentDTO, error)
	UpdateDepartment(ctx context.Context, id int64, name string) (*dto.DepartmentDTO, error)
	DeleteDepartment(ctx context.Context, id int64) error

	CreatePosition(ctx context.Context, name string) (*dto.PositionDTO, error)
	GetPosition(ctx context.Context, id int64) (*dto.PositionDTO, error)
	ListPositions(ctx context.Context) ([]dto.PositionDTO, error)
	UpdatePosition(ctx context.Context, id int64, name string) (*dto.PositionDTO, error)
	DeletePosition(ctx context.Context, id int64) error

	CreateSkill(ctx context.Context, name string) (*dto.SkillDTO, error)
	ListSkills(ctx context.Context) ([]dto.SkillDTO, error)
	DeleteSkill(ctx context.Context, id int64) error
}

type OrgService struct {
	repo OrgRepository
}

func NewOrgService(repo OrgRepository) *OrgService {
	return &OrgService{repo: repo}
}

func (s *OrgService) CreateDepartment(ctx context.Context, name string) (dto.DepartmentDTO, error) {
	if name == "" {
		return dto.DepartmentDTO{}, errors.New("name is required")
	}
	res, err := s.repo.CreateDepartment(ctx, name)
	if err != nil {
		return dto.DepartmentDTO{}, err
	}
	return *res, nil
}

func (s *OrgService) GetDepartment(ctx context.Context, id int64) (dto.DepartmentDTO, error) {
	res, err := s.repo.GetDepartment(ctx, id)
	if err != nil {
		return dto.DepartmentDTO{}, err
	}
	if res == nil {
		return dto.DepartmentDTO{}, ErrNotFound
	}
	return *res, nil
}

func (s *OrgService) ListDepartments(ctx context.Context) ([]dto.DepartmentDTO, error) {
	return s.repo.ListDepartments(ctx)
}

func (s *OrgService) UpdateDepartment(ctx context.Context, id int64, name string) (dto.DepartmentDTO, error) {
	if name == "" {
		return dto.DepartmentDTO{}, errors.New("name is required")
	}
	res, err := s.repo.UpdateDepartment(ctx, id, name)
	if err != nil {
		return dto.DepartmentDTO{}, err
	}
	return *res, nil
}

func (s *OrgService) DeleteDepartment(ctx context.Context, id int64) error {
	return s.repo.DeleteDepartment(ctx, id)
}

func (s *OrgService) CreatePosition(ctx context.Context, name string) (dto.PositionDTO, error) {
	if name == "" {
		return dto.PositionDTO{}, errors.New("name is required")
	}
	res, err := s.repo.CreatePosition(ctx, name)
	if err != nil {
		return dto.PositionDTO{}, err
	}
	return *res, nil
}

func (s *OrgService) GetPosition(ctx context.Context, id int64) (dto.PositionDTO, error) {
	res, err := s.repo.GetPosition(ctx, id)
	if err != nil {
		return dto.PositionDTO{}, err
	}
	if res == nil {
		return dto.PositionDTO{}, ErrNotFound
	}
	return *res, nil
}

func (s *OrgService) ListPositions(ctx context.Context) ([]dto.PositionDTO, error) {
	return s.repo.ListPositions(ctx)
}

func (s *OrgService) UpdatePosition(ctx context.Context, id int64, name string) (dto.PositionDTO, error) {
	if name == "" {
		return dto.PositionDTO{}, errors.New("name is required")
	}
	res, err := s.repo.UpdatePosition(ctx, id, name)
	if err != nil {
		return dto.PositionDTO{}, err
	}
	return *res, nil
}

func (s *OrgService) DeletePosition(ctx context.Context, id int64) error {
	return s.repo.DeletePosition(ctx, id)
}

func (s *OrgService) CreateSkill(ctx context.Context, name string) (dto.SkillDTO, error) {
	if name == "" {
		return dto.SkillDTO{}, errors.New("name is required")
	}
	res, err := s.repo.CreateSkill(ctx, name)
	if err != nil {
		return dto.SkillDTO{}, err
	}
	return *res, nil
}

func (s *OrgService) ListSkills(ctx context.Context) ([]dto.SkillDTO, error) {
	return s.repo.ListSkills(ctx)
}

func (s *OrgService) DeleteSkill(ctx context.Context, id int64) error {
	return s.repo.DeleteSkill(ctx, id)
}
