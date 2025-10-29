package employee

import (
	"context"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type CachedEmployeeStorage struct {
	db    EmployeeStorage
	cache *ReferenceCache
}

func NewCachedEmployeeStorage(db EmployeeStorage, cache *ReferenceCache) *CachedEmployeeStorage {
	return &CachedEmployeeStorage{
		db:    db,
		cache: cache,
	}
}

func (s *CachedEmployeeStorage) CreateProfile(ctx context.Context, tx pgx.Tx, regData *emp.RegisterData) (*emp.Profile, error) {
	return s.db.CreateProfile(ctx, tx, regData)
}

func (s *CachedEmployeeStorage) GetProfile(ctx context.Context, userID string) (*emp.Profile, error) {
	return s.db.GetProfile(ctx, userID)
}

func (s *CachedEmployeeStorage) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
	return s.db.ListProfile(ctx, pageSize, pageNum, departmentID, positionID)
}

func (s *CachedEmployeeStorage) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	return s.db.UpdateProfile(ctx, userID, updProf)
}

func (s *CachedEmployeeStorage) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return s.db.BeginTransaction(ctx)
}

func (s *CachedEmployeeStorage) CreateDepartment(ctx context.Context, name string) (*emp.Department, error) {
	department, err := s.db.CreateDepartment(ctx, name)
	if err == nil {
		_ = s.cache.InvalidateDepartments(ctx)
	}
	return department, err
}

func (s *CachedEmployeeStorage) GetDepartment(ctx context.Context, id string) (*emp.Department, error) {
	department, err := s.cache.GetDepartment(ctx, id)
	if err == nil && department != nil {
		app.Logger.Debug("department found in cache", zap.String("id", id))
		return department, nil
	}

	department, err = s.db.GetDepartment(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetDepartment(ctx, department)
	return department, nil
}

func (s *CachedEmployeeStorage) ListDepartments(ctx context.Context) ([]*emp.Department, error) {
	departments, err := s.cache.GetDepartmentsList(ctx)
	if err == nil && departments != nil {
		app.Logger.Debug("departments list found in cache")
		return departments, nil
	}

	departments, err = s.db.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetDepartmentsList(ctx, departments)
	return departments, nil
}

func (s *CachedEmployeeStorage) UpdateDepartment(ctx context.Context, id, name string) (*emp.Department, error) {
	department, err := s.db.UpdateDepartment(ctx, id, name)
	if err == nil {
		_ = s.cache.InvalidateDepartment(ctx, id)
	}
	return department, err
}

func (s *CachedEmployeeStorage) DeleteDepartment(ctx context.Context, id string) error {
	err := s.db.DeleteDepartment(ctx, id)
	if err == nil {
		_ = s.cache.InvalidateDepartment(ctx, id)
	}
	return err
}

func (s *CachedEmployeeStorage) CreatePosition(ctx context.Context, name string) (*emp.Position, error) {
	position, err := s.db.CreatePosition(ctx, name)
	if err == nil {
		_ = s.cache.InvalidatePositions(ctx)
	}
	return position, err
}

func (s *CachedEmployeeStorage) GetPosition(ctx context.Context, id string) (*emp.Position, error) {
	position, err := s.cache.GetPosition(ctx, id)
	if err == nil && position != nil {
		app.Logger.Debug("position found in cache", zap.String("id", id))
		return position, nil
	}

	position, err = s.db.GetPosition(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetPosition(ctx, position)
	return position, nil
}

func (s *CachedEmployeeStorage) ListPositions(ctx context.Context) ([]*emp.Position, error) {
	positions, err := s.cache.GetPositionsList(ctx)
	if err == nil && positions != nil {
		app.Logger.Debug("positions list found in cache")
		return positions, nil
	}

	positions, err = s.db.ListPositions(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetPositionsList(ctx, positions)
	return positions, nil
}

func (s *CachedEmployeeStorage) UpdatePosition(ctx context.Context, id, name string) (*emp.Position, error) {
	position, err := s.db.UpdatePosition(ctx, id, name)
	if err == nil {
		_ = s.cache.InvalidatePosition(ctx, id)
	}
	return position, err
}

func (s *CachedEmployeeStorage) DeletePosition(ctx context.Context, id string) error {
	err := s.db.DeletePosition(ctx, id)
	if err == nil {
		_ = s.cache.InvalidatePosition(ctx, id)
	}
	return err
}

func (s *CachedEmployeeStorage) CreateSkill(ctx context.Context, name string) (*emp.Skill, error) {
	skill, err := s.db.CreateSkill(ctx, name)
	if err == nil {
		_ = s.cache.InvalidateSkills(ctx)
	}
	return skill, err
}

func (s *CachedEmployeeStorage) ListSkills(ctx context.Context) ([]*emp.Skill, error) {
	skills, err := s.cache.GetSkillsList(ctx)
	if err == nil && skills != nil {
		app.Logger.Debug("skills list found in cache")
		return skills, nil
	}

	skills, err = s.db.ListSkills(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetSkillsList(ctx, skills)
	return skills, nil
}

func (s *CachedEmployeeStorage) AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error {
	return s.db.AddSkillToEmployee(ctx, employeeID, skillID)
}

func (s *CachedEmployeeStorage) RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error {
	return s.db.RemoveSkillFromEmployee(ctx, employeeID, skillID)
}
